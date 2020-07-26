package server

import (
	"bufio"
	"crypto/tls"
	"errors"
	"net"
	"sync"

	"example-tls/messages"
)

// OnStartListener ...
type OnStartListener func()

// OnStopListener ...
type OnStopListener func()

// OnNewConnectionListener ...
type OnNewConnectionListener func(addr string)

// OnMessageListener ...
type OnMessageListener func(message messages.Message)

// OnErrorListener ...
type OnErrorListener func(err error)

// Server ...
type Server interface {
	Start(addres string, config *tls.Config) error

	Stop() error

	OnStart(listener OnStartListener)

	OnStop(listener OnStopListener)

	OnNewConnection(listener OnNewConnectionListener)

	OnMessage(listener OnMessageListener)

	OnError(listener OnErrorListener)
}

// Create ...
func Create() (Server, error) {
	return &server{}, nil
}

type server struct {
	mutex     sync.Mutex
	isRunning bool

	conns []net.Conn

	onStartListener         OnStartListener
	onStopListener          OnStopListener
	onNewConnectionListener OnNewConnectionListener
	onMessageListener       OnMessageListener
	onErrorListener         OnErrorListener
}

func (server *server) Start(address string, config *tls.Config) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if server.isRunning {
		return errors.New("Server was already start")
	}

	listener, err := tls.Listen("tcp", address, config)
	if err != nil {
		return err
	}

	server.isRunning = true
	server.emitStart()

	go server.handleListener(listener)

	return nil
}

func (server *server) Stop() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if !server.isRunning {
		return errors.New("Server not start yet")
	}

	server.isRunning = false
	server.emitStop()

	return nil
}

func (server *server) OnStart(listener OnStartListener) {
	server.onStartListener = listener
}

func (server *server) OnStop(listener OnStopListener) {
	server.onStopListener = listener
}

func (server *server) OnNewConnection(listener OnNewConnectionListener) {
	server.onNewConnectionListener = listener
}

func (server *server) OnMessage(listener OnMessageListener) {
	server.onMessageListener = listener
}

func (server *server) OnError(listener OnErrorListener) {
	server.onErrorListener = listener
}

func (server *server) handleListener(listener net.Listener) {
	defer listener.Close()

	for server.isRunning {
		conn, err := listener.Accept()
		if err != nil {
			server.emitError(err)
			continue
		}

		server.emitNowConnection(conn.RemoteAddr().String())
		server.addConn(conn)

		go server.handleConnection(conn)
	}
}

func (server *server) handleConnection(conn net.Conn) {
	defer server.removeConn(conn)
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for server.isRunning {
		raw, err := reader.ReadBytes(messages.DELIM)
		if err != nil {
			server.emitError(err)
			return
		}

		message, err := messages.Decode(raw)
		if err != nil {
			server.emitError(err)
			continue
		}

		server.handleMessage(conn, message)
	}
}

func (server *server) handleMessage(conn net.Conn, message messages.Message) {
	server.emitMessage(message)

	switch message := message.(type) {
	case *messages.Text:
		server.sendBroadcastMessage(message)
	default:
		server.emitError(
			errors.New("Unexpected message type " + message.Type().String()),
		)
	}
}

func (server *server) sendMessage(conn net.Conn, message messages.Message) {
	raw, err := messages.Encode(message)
	if err != nil {
		server.emitError(err)
		return
	}

	_, err = conn.Write(raw)
	if err != nil {
		server.emitError(err)
	}
}

func (server *server) sendBroadcastMessage(message messages.Message) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	for _, conn := range server.conns {
		server.sendMessage(conn, message)
	}
}

func (server *server) addConn(conn net.Conn) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	server.conns = append(server.conns, conn)
}

func (server *server) removeConn(conn net.Conn) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	for i, _conn := range server.conns {
		if _conn == conn {
			server.conns = append(server.conns[:i], server.conns[i+1:]...)
		}
	}
}

func (server *server) emitStart() {
	if server.onStartListener != nil {
		go server.onStartListener()
	}
}

func (server *server) emitStop() {
	if server.onStopListener != nil {
		go server.onStopListener()
	}
}

func (server *server) emitNowConnection(address string) {
	if server.onNewConnectionListener != nil {
		go server.onNewConnectionListener(address)
	}
}

func (server *server) emitMessage(message messages.Message) {
	if server.onMessageListener != nil {
		go server.onMessageListener(message)
	}
}

func (server *server) emitError(err error) {
	if server.onErrorListener != nil {
		go server.onErrorListener(err)
	}
}

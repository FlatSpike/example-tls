package client

import (
	"bufio"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"time"

	"example-tls/messages"
)

// OnStartListener ...
type OnStartListener func(address string)

// OnStopListener ...
type OnStopListener func()

// OnMessageListener ...
type OnTextListener func(text messages.Text)

// OnErrorListener ...
type OnErrorListener func(err error)

// Client ...
type Client interface {
	Start(string, *tls.Config) error

	Stop() error

	Text(string) error

	OnStart(OnStartListener)

	OnStop(OnStopListener)

	OnText(OnTextListener)

	OnError(OnErrorListener)
}

// Create ...
func Create(name string) (Client, error) {
	return &client{name: name}, nil
}

type client struct {
	mutex sync.Mutex
	conn  net.Conn

	name      string
	isRunning bool

	onStrartListener OnStartListener
	onStopListener   OnStopListener
	onTextListener   OnTextListener
	onErrorListener  OnErrorListener
}

func (client *client) Start(address string, config *tls.Config) error {
	if client.isRunning {
		return errors.New("Client was already start")
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	conn, err := tls.Dial("tcp", "127.0.0.1:443", config)
	if err != nil {
		return err
	}

	client.conn = conn
	client.isRunning = true
	client.emitStart(conn.RemoteAddr().String())

	go client.handleConnection()

	return nil
}

func (client *client) Stop() error {
	if !client.isRunning {
		return errors.New("Client was not start yet")
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.isRunning = false
	client.emitStop()

	return nil
}

func (client *client) Text(text string) error {
	if !client.isRunning {
		return errors.New("Client was not start yet")
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.sendMessage(messages.Text{
		Name: client.name,
		Text: text,
		Time: time.Now(),
	})

	return nil
}

func (client *client) OnStart(listener OnStartListener) {
	client.onStrartListener = listener
}

func (client *client) OnStop(listener OnStopListener) {
	client.onStopListener = listener
}

func (client *client) OnText(listener OnTextListener) {
	client.onTextListener = listener
}

func (client *client) OnError(listener OnErrorListener) {
	client.onErrorListener = listener
}

func (client *client) handleConnection() {
	defer client.conn.Close()

	reader := bufio.NewReader(client.conn)
	for client.isRunning {
		rewMessage, err := reader.ReadBytes(messages.DELIM)
		if err != nil {
			client.emitError(err)
			return
		}

		message, err := messages.Decode(rewMessage)
		if err != nil {
			client.emitError(err)
			continue
		}

		client.handleMessage(message)
	}
}

func (client *client) handleMessage(message messages.Message) {
	switch message := message.(type) {
	case *messages.Text:
		client.emitText(*message)
	default:
		client.emitError(
			errors.New("Unexpected message type " + message.Type().String()),
		)
	}
}

func (client *client) sendMessage(message messages.Message) {
	bytes, err := messages.Encode(message)
	if err != nil {
		client.emitError(err)
		return
	}

	_, err = client.conn.Write(bytes)
	if err != nil {
		client.emitError(err)
	}
}

func (client *client) emitStart(address string) {
	if client.onStrartListener != nil {
		go client.onStrartListener(address)
	}
}

func (client *client) emitStop() {
	if client.onStopListener != nil {
		go client.onStopListener()
	}
}

func (client *client) emitText(text messages.Text) {
	if client.onTextListener != nil {
		go client.onTextListener(text)
	}
}

func (client *client) emitError(err error) {
	if client.onErrorListener != nil {
		go client.onErrorListener(err)
	}
}

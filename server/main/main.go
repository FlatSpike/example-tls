package main

import (
	"crypto/tls"
	"log"
	"sync"

	"example-tls/messages"
	"example-tls/server"
)

//nolint:deadcode,unused
func main() {
	log.SetFlags(log.Lshortfile)

	var wg sync.WaitGroup

	cer, err := tls.LoadX509KeyPair(
		"certificates/cert.crt",
		"certificates/cert.key",
	)
	if err != nil {
		log.Println(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	server, err := server.Create()
	if err != nil {
		log.Println(err)
		return
	}

	server.OnStart(func() {
		log.Println("OnStart")
	})

	server.OnStop(func() {
		log.Println("OnStop")
		wg.Done()
	})

	server.OnNewConnection(func(address string) {
		log.Println("OnNewConnection", address)
	})

	server.OnMessage(func(message messages.Message) {
		log.Println("OnMessage", message)
	})

	server.OnError(func(err error) {
		log.Println("OnError", err)
	})

	wg.Add(1)
	err = server.Start(":443", config)
	if err != nil {
		log.Println(err)
		return
	}

	wg.Wait()
}

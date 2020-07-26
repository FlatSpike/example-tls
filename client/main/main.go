package main

import (
	"crypto/tls"
	"crypto/x509"
	"example-tls/messages"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"sync"
	"time"

	"example-tls/client"
)

func main() {
	log.SetFlags(log.Lshortfile)

	var name string
	fmt.Println("Init client")
	fmt.Print("Type in name: ")
	fmt.Scanln(&name)

	certFile, err := ioutil.ReadFile("certificates/ca.crt")
	if err != nil {
		log.Println(err)
		return
	}

	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM(certFile)
	if !ok {
		panic("failed to parse root certificate")
	}

	conf := &tls.Config{RootCAs: certpool}

	client, err := client.Create(name)
	if err != nil {
		log.Println(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	client.OnStart(func(address string) {
		fmt.Println("Client started")
		fmt.Println("Connected to", address)

		wg.Done()
	})

	client.OnStop(func() {
		fmt.Println("Client stoped")
	})

	client.OnText(func(text messages.Text) {
		log.Println("OnText", text)
	})

	client.OnError(func(err error) {
		log.Println("OnError", err)
	})

	err = client.Start("127.0.0.1:443", conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	wg.Wait()
	err = sendMessages(client, 100)
	if err != nil {
		panic(err)
	}
}

func sendMessages(client client.Client, count int) error {
	for i := 0; i < count; i++ {
		err := client.Text("ms " + strconv.Itoa(i))
		if err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

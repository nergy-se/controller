package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func main() {
	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
	})
	// Create signals channel to run server until interrupted
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		server.Close()
		done <- true
	}()

	// Create the new MQTT Server.

	// Allow all connections.
	_ = server.AddHook(new(auth.AllowHook), nil)

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP(listeners.Config{ID: "t1", Address: ":1883"})
	err := server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}
	// server.Subscribe()
	// callbackFn := func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	//     server.Log.Info("inline client received message from subscription", "client", cl.ID, "subscriptionId", sub.Identifier, "topic", pk.TopicName, "payload", string(pk.Payload))
	// }
	// server.Subscribe("direct/#", 1, callbackFn)
	// example here
	// https://github.com/mochi-mqtt/server/blob/main/examples/direct/main.go

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Run server until interrupted
	<-done

	// Cleanup
}

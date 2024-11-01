package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/sirupsen/logrus"
)

func main() {
	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
	})
	// Create signals channel to run server until interrupted
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer cancel()
	// Create the new MQTT Server.

	// Allow all connections.
	_ = server.AddHook(new(auth.AllowHook), nil)

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP(listeners.Config{ID: "t1", Address: ":1883"})
	err := server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {

		err := server.Subscribe("p1ib/#", 1, func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
			server.Log.Info("inline client received message from subscription", "client", cl.ID, "subscriptionId", sub.Identifier, "topic", pk.TopicName, "payload", string(pk.Payload))
		})
		if err != nil {
			logrus.Error(err)
			return
		}

		//TODO listen for only this topic
		/*
		   p1ib/sensor_state:

		   {
		     "p1ib_hourly_active_import_q1_q4": 76215.335,
		     "p1ib_hourly_active_export_q2_q3": 12925.573,
		     "p1ib_hourly_reactive_import_q1_q2": 7721.648,
		     "p1ib_hourly_reactive_export_q3_q4": 11314.6,
		     "p1ib_active_power_plus_q1_q4": 4.396,
		     "p1ib_active_power_minus_q2_q3": 0,
		     "p1ib_reactive_power_plus_q1_q2": 0,
		     "p1ib_reactive_power_minus_q3_q4": 0.883,
		     "p1ib_active_power_plus_l1": 0.764,
		     "p1ib_active_power_minus_l1": 0,
		     "p1ib_active_power_plus_l2": 0.9,
		     "p1ib_active_power_minus_l2": 0,
		     "p1ib_active_power_plus_l3": 2.732,
		     "p1ib_active_power_minus_l3": 0,
		     "p1ib_reactive_power_plus_l1": 0,
		     "p1ib_reactive_power_minus_l1": 0.32,
		     "p1ib_reactive_power_plus_l2": 0,
		     "p1ib_reactive_power_minus_l2": 0.402,
		     "p1ib_reactive_power_plus_l3": 0,
		     "p1ib_reactive_power_minus_l3": 0.149,
		     "p1ib_voltage_l1": 233.6,
		     "p1ib_voltage_l2": 234,
		     "p1ib_voltage_l3": 232.3,
		     "p1ib_current_l1": 3.5,
		     "p1ib_current_l2": 4.2,
		     "p1ib_current_l3": 11.8,
		     "p1ib_firmware": "54aa555",
		     "p1ib_update_available": "no",
		     "p1ib_import_export_l1": 0.764,
		     "p1ib_import_export_l2": 0.9,
		     "p1ib_import_export_l3": 2.732,
		     "p1ib_import_export": 4.396,
		     "p1ib_rssi": "-58",
		     "p1ib_meter": "Aidon",
		     "p1ib_telegrams_crc_ok": 312677,
		     "p1ib_ip": "192.168.0.232",
		     "p1ib_wifi_mac": "b0b21ca00a68"
		   }
		*/

	}()

	// Run server until interrupted
	<-ctx.Done()
	server.Close()

	// Cleanup
}

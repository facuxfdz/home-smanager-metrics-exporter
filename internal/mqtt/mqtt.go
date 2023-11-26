package mqtt

import (
	"fmt"
	"log"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	fmt.Println("Connected to MQTT broker")
}

var connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

// create a new client
func BuildClient() MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883")
	currentTimestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	clientId := "home_smanager_client_" + string(currentTimestamp)
	log.Printf("Client ID: %s\n", clientId)
	opts.SetClientID(clientId)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := MQTT.NewClient(opts)
	return client
}

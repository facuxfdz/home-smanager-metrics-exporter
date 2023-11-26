package mqtt

import (
	"fmt"
	"math/rand"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	deviceRooms = map[string]string{
		"device1": "living_room",
		"device2": "living_room",
		"device3": "bedroom",
		"device4": "kitchen",
		"device5": "bedroom",
		"device6": "bathroom",
		"device7": "bedroom",
		"device8": "bathroom",
	}
)

func generateRandomEnergyConsumption() float64 {
	return rand.Float64() * 50.0
}

func MockPublish(client MQTT.Client, topic string, qos byte) MQTT.Token {
	rand.Seed(time.Now().UnixNano())
	for devId, room := range deviceRooms {
		deviceType := "sensor"
		energyConsumption := generateRandomEnergyConsumption()
		payload := fmt.Sprintf("%s:%s:%s:%.2f", devId, room, deviceType, energyConsumption)
		token := client.Publish(topic, qos, false, payload)
		token.Wait()
		time.Sleep(1 * time.Second)
	}
	return nil
}

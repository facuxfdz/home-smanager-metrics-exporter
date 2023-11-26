package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/facuxfdz/home-smanager-metrics-exporter/internal/mqtt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	energyTotalConsumption   *prometheus.CounterVec
	currentEnergyConsumption *prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {

	m := &metrics{
		energyTotalConsumption: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "energy_total_consumption",
				Help: "Total energy consumption for all devices in the smart home",
			},
			[]string{"device", "room", "type"},
		),
		currentEnergyConsumption: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "current_energy_consumption",
				Help: "Current energy consumption for all devices in the smart home",
			},
			[]string{"device", "room", "type"},
		),
	}

	reg.MustRegister(m.energyTotalConsumption)
	reg.MustRegister(m.currentEnergyConsumption)
	return m
}

func parseBoolOrDefault(value string, defaultValue bool) bool {

	_, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return true
}

var (
	enableMockPublisher = parseBoolOrDefault(os.Getenv("ENABLE_MOCK_PUBLISHER"), false)
)

func init() {
	log.Printf("Mock publisher enabled: %t\n", enableMockPublisher)
}

func msgHandler(client MQTT.Client, msg MQTT.Message, m *metrics) {
	// get device id from message with the format of "deviceID:room:type:energyConsumption"
	payload := string(msg.Payload())
	// split the message by the colon
	splitPayload := strings.Split(payload, ":")
	// get the device id
	deviceID := splitPayload[0]
	// get the room
	room := splitPayload[1]
	// get the type
	deviceType := splitPayload[2]
	// get the energy consumption
	energyConsumption, _ := strconv.ParseFloat(strings.TrimSpace(splitPayload[3]), 64)

	// handle metrics
	m.energyTotalConsumption.WithLabelValues(deviceID, room, deviceType).Add(energyConsumption)
	m.currentEnergyConsumption.WithLabelValues(deviceID, room, deviceType).Set(energyConsumption)

	log.Printf("Device ID: %s, Room: %s, Type: %s, Energy Consumption: %f\n", deviceID, room, deviceType, energyConsumption)
}

func main() {

	// create a non-global registry
	reg := prometheus.NewRegistry()

	// register the custom metrics
	m := NewMetrics(reg)

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
		log.Print("Starting prometheus server on port 8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	client := mqtt.BuildClient()
	token := client.Connect()
	token.Wait()

	tokensub := client.Subscribe("home-smanager/test", 1, func(client MQTT.Client, msg MQTT.Message) {
		msgHandler(client, msg, m)
	})
	tokensub.Wait()

	log.Printf("Subscribed to topic: %s\n", "home-smanager/test")

	ticker := time.NewTicker(8 * time.Second)
	defer ticker.Stop()

	publisherClient := mqtt.BuildClient()
	publisherToken := publisherClient.Connect()
	publisherToken.Wait()

	if enableMockPublisher {
		go func() {
			for {
				<-ticker.C
				mqtt.MockPublish(publisherClient, "home-smanager/test", 0)
			}
		}()
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	<-signalCh

	client.Disconnect(250)
	publisherClient.Disconnect(250)

}

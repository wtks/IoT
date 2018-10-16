package main

import (
	"fmt"
	"github.com/d2r2/go-bsbmp"
	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	ClientID         = "rpizerow_bme280"
	PubTopicTemp     = "/home/temperature"
	PubTopicHumid    = "/home/humidity"
	PubTopicPressure = "/home/pressure"
	PubTopicMetrics  = "/home/metrics"
)

var (
	MQTTHost = os.Getenv("MQTT_HOST")
)

func main() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, os.Kill)

	// init mqtt client
	mqttOpt := mqtt.NewClientOptions()
	mqttOpt.AddBroker(fmt.Sprintf("tcp://%s:1883", MQTTHost))
	mqttOpt.SetClientID(ClientID)

	client := mqtt.NewClient(mqttOpt)
	defer client.Disconnect(250)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	// init i2c
	i, err := i2c.NewI2C(0x76, 1)
	if err != nil {
		log.Fatal(err)
	}
	defer i.Close()
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

	sensor, err := bsbmp.NewBMP(bsbmp.BME280, i)
	if err != nil {
		log.Fatal(err)
	}
	logger.ChangePackageLogLevel("bsbmp", logger.InfoLevel)

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

	BreakFor:
		for {
			select {
			case <-stop:
				break BreakFor
			case <-ticker.C:
				temp, err := sensor.ReadTemperatureC(bsbmp.ACCURACY_STANDARD)
				if err != nil {
					log.Fatal(err)
				}
				_, humid, err := sensor.ReadHumidityRH(bsbmp.ACCURACY_STANDARD)
				if err != nil {
					log.Fatal(err)
				}
				pressure, err := sensor.ReadPressurePa(bsbmp.ACCURACY_STANDARD)
				if err != nil {
					log.Fatal(err)
				}
				pressure /= 100
				publish(client, PubTopicMetrics, fmt.Sprintf(`{"t":%2.2f,"h":%0.2f,"p":%f}`, temp, humid, pressure))
				publish(client, PubTopicTemp, fmt.Sprintf("%2.2f", temp))
				publish(client, PubTopicHumid, fmt.Sprintf("%0.2f", humid))
				publish(client, PubTopicPressure, fmt.Sprintf("%f", pressure))
			}
		}
	}()

	<-sigint
	close(stop)
}

func publish(client mqtt.Client, topic, body string) {
	token := client.Publish(topic, 0, false, body)
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
}

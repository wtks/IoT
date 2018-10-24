package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/djthorpe/gopi"
	_ "github.com/djthorpe/gopi-hw/sys/lirc"
	_ "github.com/djthorpe/gopi/sys/logger"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/wtks/A75C4269"
	"log"
	"os"
	"os/signal"
)

const (
	ClientID = "rpizerow_aircon"
	SubTopic = "/aircon/action"
	PubTopic = "/aircon/state"
)

var (
	MQTTHost     = os.Getenv("MQTT_HOST")
	MQTTUserName = os.Getenv("MQTT_USERNAME")
	MQTTPassword = os.Getenv("MQTT_PASSWORD")
)

func main() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, os.Kill)

	// init mqtt client
	mqttOpt := mqtt.NewClientOptions()
	mqttOpt.AddBroker(fmt.Sprintf("tcp://%s:1883", MQTTHost))
	mqttOpt.SetUsername(MQTTUserName)
	mqttOpt.SetPassword(MQTTPassword)
	mqttOpt.SetClientID(ClientID)

	client := mqtt.NewClient(mqttOpt)
	defer client.Disconnect(250)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	config := gopi.NewAppConfig("lirc")

	recv := make(chan mqtt.Message)
	token := client.Subscribe(SubTopic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		recv <- msg
	})
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	os.Exit(gopi.CommandLineTool(config, func(app *gopi.AppInstance, done chan<- struct{}) error {
		if app.LIRC == nil {
			return errors.New("missing LIRC module")
		}

		for {
			select {
			case <-sigint:
				done <- gopi.DONE
				return nil
			case msg := <-recv:
				c := A75C4269.Controller{}
				if err := json.Unmarshal(msg.Payload(), &c); err != nil {
					app.Logger.Error(err.Error())
					break
				}

				if err := app.LIRC.PulseSend(c.GetRawSignal()); err != nil {
					return err
				}

				payload, _ := json.Marshal(c)
				token := client.Publish(PubTopic, 1, true, string(payload))
				if token.Wait() && token.Error() != nil {
					app.Logger.Error(token.Error().Error())
					break
				}
			}
		}

		done <- gopi.DONE
		return nil
	}))
}

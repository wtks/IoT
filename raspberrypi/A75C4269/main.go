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
)

const (
	ClientID = "rpizerow_aircon"
	SubTopic = "/aircon/action"
	PubTopic = "/aircon/state"
)

var MQTTHost = os.Getenv("MQTT_HOST")

func main() {
	// init mqtt client
	mqttOpt := mqtt.NewClientOptions()
	mqttOpt.AddBroker(fmt.Sprintf("tcp://%s:1883", MQTTHost))
	mqttOpt.SetClientID(ClientID)

	client := mqtt.NewClient(mqttOpt)
	defer client.Disconnect(250)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	config := gopi.NewAppConfig("lirc")

	os.Exit(gopi.CommandLineTool(config, func(app *gopi.AppInstance, done chan<- struct{}) error {
		if app.LIRC == nil {
			return errors.New("missing LIRC module")
		}

		token := client.Subscribe(SubTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
			c := A75C4269.Controller{}
			if err := json.Unmarshal(msg.Payload(), &c); err != nil {
				app.Logger.Error(err.Error())
				return
			}

			if err := app.LIRC.PulseSend(c.GetRawSignal()); err != nil {
				app.Logger.Error(err.Error())
				return
			}

			token := client.Publish(PubTopic, 0, false, string(msg.Payload()))
			if token.Wait() && token.Error() != nil {
				app.Logger.Error(token.Error().Error())
				return
			}
		})

		if token.Wait() && token.Error() != nil {
			return token.Error()
		}

		done <- gopi.DONE
		return nil
	}))
}

package internal

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/webishdev/fritze-mqtt/log"
)

func StartMQTT(mqttChan chan byte, broker string, port int, topic string) error {
	brokerURL := fmt.Sprintf("tcp://%s:%d", broker, port)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID("fritze-mqtt")
	opts.SetDefaultPublishHandler(messagePubHandler)
	//opts.SetUsername("fritze")
	//opts.SetPassword("mq")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	token := client.Subscribe(topic, 1, nil)

	token.Wait()
	log.Info("Subscribed to topic %s", topic)

	select {
	case <-mqttChan:
		{
			client.Disconnect(0)
			return nil
		}
	}
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Info("Received message: %s from topic: %s", msg.Payload(), msg.Topic())
}

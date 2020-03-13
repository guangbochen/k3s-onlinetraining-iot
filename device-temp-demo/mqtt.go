package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yosssi/gmq/mqtt/client"
)

type Message struct {
	PropertyName string
	Result       string
	TimeStamp    int64
}

func ConnectToMQTT(clientID string, mq Mqtt) (*client.Client, error) {
	logrus.Printf("connecting to the mqtt server: %s", mq.Server)
	cli := client.New(&client.Options{
		ErrorHandler: func(err error) {
			logrus.Println(err)
		},
	})
	defer cli.Terminate()

	err := cli.Connect(&client.ConnectOptions{
		Network:      "tcp",
		Address:      mq.Server,
		ClientID:     []byte(clientID),
		CleanSession: true,
	})
	if err != nil {
		logrus.Errorf("error connecting to the mqtt server: %s", err.Error())
		return nil, err
	}
	return cli, nil
}

// PublishToMQTT publish the message to the MQTT server
func PublishToMQTT(cli *client.Client, mq Mqtt, message []byte) error {
	topic := strings.TrimSuffix(mq.Topic, "/")
	err := cli.Publish(&client.PublishOptions{
		QoS:       byte(mq.Qos),
		TopicName: []byte(topic),
		Message:   message,
	})
	if err != nil {
		return err
	}
	return nil
}

// SubscribeToMQTT subscribe to the MQTT topic
func SubscribeToMQTT(cli *client.Client, mq Mqtt) error {
	topic := fmt.Sprintf("%s/#", strings.TrimSuffix(mq.Topic, "/"))
	err := cli.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte(topic),
				QoS:         byte(mq.Qos),
				Handler: func(topicName, message []byte) {
					logrus.Printf("Subscribed MQTT topic: %s, message: %s", topic, string(message))
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func FormatMessage(property, value string) Message {
	var message Message
	message.PropertyName = property
	message.Result = fmt.Sprintf("%s", value)
	message.TimeStamp = time.Now().UTC().UnixNano()
	return message
}

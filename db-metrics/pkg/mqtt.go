package pkg

import (
	"github.com/sirupsen/logrus"
	"github.com/yosssi/gmq/mqtt/client"
)

type Message struct {
	PropertyName string
	Result       string
	TimeStamp    int64
}

type Mqtt struct {
	Server string `json:"server"`
	Topic  string `json:"topic"`
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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"time"

	"db-metrics/pkg"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	influxClient "github.com/influxdata/influxdb1-client"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/yosssi/gmq/mqtt"
	mqClient "github.com/yosssi/gmq/mqtt/client"
)

type influxDB struct {
	url      string `json:"url"`
	port     int    `json:"port"`
	username string `json:"username"`
	password string `json:"password"`
}

const (
	VERSION = "0.0.0-dev"
)

func main() {
	app := cli.NewApp()
	app.Name = "grafana-metrics"
	app.Usage = "helps to collecting grafana metrics"
	app.Version = VERSION
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "debug",
		},
		cli.StringFlag{
			Name:   "database",
			EnvVar: "DATABASE",
			Value:  "",
			Usage:  "specify influxDB database name.",
		},
		cli.StringFlag{
			Name:   "influxdb_server",
			EnvVar: "INFLUXDB_SERVER",
			Value:  "",
			Usage:  "specify influxDB server url.",
		},
		cli.StringFlag{
			Name:   "influxdb_username",
			EnvVar: "INFLUXDB_USERNAME",
			Value:  "",
			Usage:  "specify influxDB username.",
		},
		cli.StringFlag{
			Name:   "influxdb_password",
			EnvVar: "INFLUXDB_PASSWORD",
			Value:  "",
			Usage:  "specify influxDB password.",
		},
		cli.IntFlag{
			Name:   "influxdb_port",
			EnvVar: "INFLUXDB_PORT",
			Usage:  "specify influxDB port number.",
		},
		cli.StringFlag{
			Name:   "mqtt_server",
			EnvVar: "MQTT_SERVER",
			Value:  "127.0.0.1:1883",
			Usage:  "specify mqtt server.",
		},
		cli.StringFlag{
			Name:   "mqtt_topic",
			EnvVar: "MQTT_TOPIC",
			Value:  "",
			Usage:  "specify mqtt subscribe topic.",
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}

}

func run(c *cli.Context) {
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ctx := signals.SetupSignalHandler(context.Background())
	// start the influxDB client
	server := c.String("influxdb_server")
	username := c.String("influxdb_username")
	password := c.String("influxdb_password")
	port := c.Int("influxdb_port")
	dbConfig, err := initInfluxDB(server, username, password, port)
	if err != nil {
		logrus.Fatal(err)
	}

	conn, err := initInfluxDBClient(dbConfig)
	if err != nil {
		logrus.Fatal(err)
	}

	_, _, err = conn.Ping()
	if err != nil {
		logrus.Fatalf("failed to ping influxDB with err: %s.\n", err.Error())
	}

	// init mqtt client and subscribe to the topic
	mq := pkg.Mqtt{
		Server: c.String("mqtt_server"),
		Topic:  c.String("mqtt_topic"),
	}
	cli, err := pkg.ConnectToMQTT("influxDB_client", mq)
	if err != nil {
		logrus.Fatal(err)
	}

	if err = subscribeToMQTT(cli, mq, conn); err != nil {
		logrus.Fatal(err)
	}

	<-ctx.Done()
	// Disconnect the Network Connection.
	if err := cli.Disconnect(); err != nil {
		logrus.Fatal(err.Error())
	}
}

func initInfluxDBClient(c *influxDB) (*influxClient.Client, error) {
	logrus.Printf("connecting to the influxDB server:%s:%d\n", c.url, c.port)
	host, err := url.Parse(fmt.Sprintf("http://%s:%d", c.url, c.port))
	if err != nil {
		return nil, err
	}

	// NOTE: this assumes you've setup a user and have setup shell env variables,
	// namely INFLUX_USER/INFLUX_PWD. If not just omit Username/Password below.
	conf := influxClient.Config{
		URL:       *host,
		Username:  c.username,
		Password:  c.password,
		UnsafeSsl: true,
	}
	client, err := influxClient.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// SubscribeToMQTT subscribe to the MQTT topic
func subscribeToMQTT(cli *mqClient.Client, mq pkg.Mqtt, conn *influxClient.Client) error {
	err := cli.Subscribe(&mqClient.SubscribeOptions{
		SubReqs: []*mqClient.SubReq{
			&mqClient.SubReq{
				TopicFilter: []byte(mq.Topic),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					logrus.Printf("Subscribed MQTT topic: %s, message: %s", mq.Topic, string(message))
					messageData := pkg.Message{}
					err := json.Unmarshal(message, &messageData)
					if err != nil {
						logrus.Errorf("failed to decode mqtt message: %s, error: %s\n", string(message), err.Error())
						return
					}
					writeToDB(messageData, conn)
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func writeToDB(message pkg.Message, conn *influxClient.Client) {
	re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
	submatchall := re.FindAllString(message.Result, -1)
	if len(submatchall) < 2 {
		logrus.Errorf("failed to sub-struct elements from string:%s, sub total len:%d.", message.Result, len(submatchall))
	}
	point := influxClient.Point{
		Measurement: "temperature",
		Tags: map[string]string{
			"name": "temp",
		},
		Fields: map[string]interface{}{
			"temperature": submatchall[0],
			"humid":       submatchall[1],
		},
		Time:      time.Unix(0, message.TimeStamp),
		Precision: "ns",
	}

	bps := influxClient.BatchPoints{
		Points:          []influxClient.Point{point},
		Database:        "mydb",
		RetentionPolicy: "autogen",
	}
	_, err := conn.Write(bps)
	if err != nil {
		logrus.Errorf("failed to write to influxDB with err: %s", err.Error())
	}
	logrus.Infof("success write data into the influxDB")
}

func initInfluxDB(server, username, password string, port int) (*influxDB, error) {
	if len(server) == 0 {
		return nil, fmt.Errorf("invalid influxdb Server URL %s", server)
	}
	if len(username) == 0 {
		return nil, fmt.Errorf("invalid influxdb username %s", username)
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("invalid influxdb password len=%d", len(password))
	}
	if port == 0 {
		return nil, fmt.Errorf("invalid influxdb port: %d", port)
	}

	return &influxDB{
		url:      server,
		port:     port,
		username: username,
		password: password,
	}, nil
}

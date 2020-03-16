package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"time"

	"github.com/bettercap/gatt"
	"github.com/bettercap/gatt/examples/option"
	"github.com/urfave/cli"
	"k8s.io/klog"
)

var (
	Version   = "v0.0.0-dev"
	GitCommit = "HEAD"
)

func main() {
	app := cli.NewApp()
	app.Name = "bluetooth-device"
	app.Version = fmt.Sprintf("%s (%s)", Version, GitCommit)
	app.Usage = "Bluetooth device adaptor"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "device_name",
			EnvVar: "DEVICE_NAME",
			Value:  "",
		},
		cli.StringFlag{
			Name:   "device_mac_address",
			EnvVar: "DEVICE_MAC_ADDRESS",
			Value:  "",
		},
		cli.StringFlag{
			Name:   "device_mqtt",
			EnvVar: "DEVICE_MQTT",
			Value:  "",
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		klog.Fatal(err)
	}
}

func run(c *cli.Context) {
	klog.Info("Starting bluetooth device adaptor")
	// handle device options
	name := c.String("device_name")
	macAddress := c.String("device_mac_address")
	deviceMQTT := c.String("device_mqtt")

	if len(name) == 0 && len(macAddress) == 0 {
		klog.Fatal("device name or device MAC address is required.")
	}

	if len(deviceMQTT) == 0 {
		klog.Fatal("mqtt config is required for device.")
	}
	mqtt, err := loadMqttConfig(deviceMQTT)
	if err != nil {
		klog.Fatal(err)
	}

	dv := Device{
		Name:       name,
		MacAddress: macAddress,
		Mqtt:       mqtt,
	}

	cli, err := ConnectToMQTT("bluetooth-temp", dv.Mqtt)
	if err != nil {
		log.Fatalf("Failed to connect to the mqtt server, err: %s\n", err)
		return
	}
	mqttCli = cli

	// subscribe to the MQTT device topic
	err = SubscribeToMQTT(cli, dv.Mqtt)
	if err != nil {
		log.Fatalf("Failed to subscribe the mqtt server, err: %s\n", err)
		return
	}

	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	for {
		// Register handlers.
		d.Handle(
			gatt.PeripheralDiscovered(dv.onPeriphDiscovered),
			gatt.PeripheralConnected(dv.onPeriphConnected),
			gatt.PeripheralDisconnected(dv.onPeriphDisconnected),
		)

		d.Init(dv.onStateChanged)
		<-done
		fmt.Println("Done")
		fmt.Println("Sleep for 10 seconds")
		time.Sleep(10 * time.Second)
	}

	// disconnect mqtt cli
	if err := cli.Disconnect(); err != nil {
		logrus.Fatalln(err)
	}
}

func loadMqttConfig(mqttStr string) (Mqtt, error) {
	mqtt := Mqtt{}
	if len(mqttStr) != 0 {
		err := json.Unmarshal([]byte(mqttStr), &mqtt)
		if err != nil {
			return mqtt, fmt.Errorf("failed to unmarshall mqtt env:%s, err: %s", mqttStr, err.Error())
		}
	}
	return mqtt, nil
}

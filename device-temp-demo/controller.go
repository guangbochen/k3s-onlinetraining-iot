package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bettercap/gatt"
	"github.com/sirupsen/logrus"
	"github.com/yosssi/gmq/mqtt/client"
)

type Device struct {
	Name       string
	MacAddress string
	Mqtt       Mqtt
}

var mqttCli *client.Client

type Mqtt struct {
	Server string
	Topic  string
	Qos    int
}

var done = make(chan struct{})

func (device *Device) onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("Scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func (device *Device) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	name := device.Name
	addr := device.MacAddress
	if name != "" && a.LocalName != name {
		return
	}

	if addr != "" && strings.ToUpper(p.ID()) != strings.ToUpper(addr) {
		return
	}

	// Stop scanning once we've got the peripheral we're looking for.
	fmt.Printf("Stop scanning and found device %s\n", a.LocalName)
	p.Device().StopScanning()
	fmt.Printf("Peripheral ID:%s, name:%s\n", p.ID(), p.Name())
	p.Device().Connect(p)
}

func (device *Device) onPeriphConnected(p gatt.Peripheral, err error) {
	fmt.Println("Connected")
	defer p.Device().CancelConnection(p)

	if err := p.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}

	// Discovery services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	for _, s := range ss {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		fmt.Println(msg)

		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			msg := "  Characteristic  " + c.UUID().String()
			if len(c.Name()) > 0 {
				msg += " (" + c.Name() + ")"
			}
			msg += "\n    properties    " + c.Properties().String()
			fmt.Println(msg)

			// Read the characteristic, if possible.
			if (c.Properties() & gatt.CharRead) != 0 {
				b, err := p.ReadCharacteristic(c)
				if err != nil {
					fmt.Printf("Failed to read characteristic, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Discovery descriptors
			ds, err := p.DiscoverDescriptors(nil, c)
			if err != nil {
				fmt.Printf("Failed to discover descriptors, err: %s\n", err)
				continue
			}

			for _, d := range ds {
				msg := "  Descriptor      " + d.UUID().String()
				if len(d.Name()) > 0 {
					msg += " (" + d.Name() + ")"
				}
				fmt.Println(msg)

				// Read descriptor (could fail, if it's not readable)
				b, err := p.ReadDescriptor(d)
				if err != nil {
					fmt.Printf("Failed to read descriptor, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Subscribe the characteristic, if possible.
			if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
				f := func(c *gatt.Characteristic, b []byte, err error) {
					fmt.Printf("notified: % X | %q\n", b, b)
					message := FormatMessage("temp-data", fmt.Sprintf("%q", b))
					mqttMessage, err := json.Marshal(message)
					if err != nil {
						logrus.Errorf("Failed to format mqtt message, error: %s", err)
						return
					}
					PublishToMQTT(mqttCli, device.Mqtt, mqttMessage)
				}
				if err := p.SetNotifyValue(c, f); err != nil {
					fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
					continue
				}
			}

		}
		fmt.Println()
	}

	fmt.Printf("Waiting for 5 seconds to get some notifiations, if any.\n")
	time.Sleep(5 * time.Second)
}

func (device *Device) onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected")
	close(done)
}

package internal

import (
	"github.com/webishdev/fritze-mqtt/fritzbox"
	"github.com/webishdev/fritze-mqtt/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var sigs chan os.Signal

func StartController(fc fritzbox.FritzClient, username string, password string) error {
	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	session, errLogin := fc.Login(username, password)
	if errLogin != nil {
		return errLogin
	}

	deviceChan := make(chan []fritzbox.Device)

	go handler(deviceChan)

	return loop(fc, session, deviceChan)
}

func loop(fc fritzbox.FritzClient, session fritzbox.Session, deviceChan chan []fritzbox.Device) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		devices, errDevices := getDevices(fc, session)
		if errDevices != nil {
			return errDevices
		}
		deviceChan <- devices
		select {
		case sig := <-sigs:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				errLogout := fc.Logout(session)
				return errLogout
			}
		case <-ticker.C:
		}
	}
}

func getDevices(fc fritzbox.FritzClient, session fritzbox.Session) ([]fritzbox.Device, error) {
	devices, errDevices := fc.GetDevices(session)
	if errDevices != nil {
		return nil, errDevices
	}

	return devices, nil
}

func handler(deviceChan chan []fritzbox.Device) {
	identifierToDevice := map[string]fritzbox.Device{}
	for {
		select {
		case devices := <-deviceChan:
			log.Info("Received %d devices\n", len(devices))
			for _, device := range devices {
				current, exists := identifierToDevice[device.Identifier]
				if exists {
					if device.Triggered && current.StateValue == device.StateValue {
						log.Info("Device %s: %s, [%s] is currently triggered", device.Identifier, device.Name, device.Description)
					}
					if current.StateValue != device.StateValue {
						log.Info("Device %s: %s, [%s] changed from %d to %d", device.Identifier, device.Name, device.Description, current.StateValue, device.StateValue)
					}
					identifierToDevice[device.Identifier] = device
				} else {
					identifierToDevice[device.Identifier] = device
					log.Debug("New device %s: %s, [%s]", device.Identifier, device.Name, device.Description)
					continue
				}

			}
		}
	}
}

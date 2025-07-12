package internal

import (
	"fmt"
	"github.com/webishdev/fritze-mqtt/fritzbox"
	"github.com/webishdev/fritze-mqtt/log"
)

func ListDevices(fc fritzbox.FritzClient, username string, password string) error {
	log.SetLogLevel(10)
	session, errLogin := fc.Login(username, password)
	if errLogin != nil {
		return errLogin
	}

	devices, errLogin := fc.GetDevices(session)
	if errLogin != nil {
		return errLogin
	}

	for _, device := range devices {
		fmt.Printf("%s: %s, [%s]\n", device.Identifier, device.Name, device.Description)
	}

	errLogout := fc.Logout(session)
	if errLogout != nil {
		return errLogout
	}

	return nil
}

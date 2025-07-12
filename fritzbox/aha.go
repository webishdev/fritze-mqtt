package fritzbox

import (
	"encoding/xml"
	"fmt"
	"github.com/webishdev/fritze-mqtt/log"
	"io"
	"strings"
	"time"
)

type DeviceFunction uint32

const (
	HANFUNDevice DeviceFunction = iota
	_
	Light
	_
	AlarmSensor
	AVMButton
	AVMHeatingController
	AVMPowerMeter
	TemperatureSensor
	AVMOutletSwitch
	AVMDECTRepeater
	AVMMicrophone
	_
	HANFUNUnit
	_
	SimpleOnOffDevice
	DimmableLevelDevice
	ColorAdjustableLight
	Blinds
	_
	HumiditySensor
)

type DeviceType uint32

const (
	TypeSimpleButton                DeviceType = 273
	TypeSimpleOnOffSwitchable       DeviceType = 256
	TypeSimpleOnOffSwitch           DeviceType = 257
	TypeACOutlet                    DeviceType = 262
	TypeACOutletSimplePowerMetering DeviceType = 263
	TypeSimpleLight                 DeviceType = 264
	TypeDimmableLight               DeviceType = 265
	TypeDimmerSwitch                DeviceType = 266
	TypeColorBulb                   DeviceType = 277
	TypeDimmableColorBulb           DeviceType = 278
	TypeBlind                       DeviceType = 281
	TypeLamellar                    DeviceType = 282
	TypeSimpleDetector              DeviceType = 512
	TypeDoorOpenCloseDetector       DeviceType = 513
	TypeWindowOpenCloseDetector     DeviceType = 514
	TypeMotionDetector              DeviceType = 515
	TypeFloodDetector               DeviceType = 518
	TypeGlassBreakDetector          DeviceType = 519
	TypeVibrationDetector           DeviceType = 520
	TypeSiren                       DeviceType = 640
)

type DeviceInterfaces uint32

const (
	InterfaceKeepAlive       DeviceInterfaces = 277
	InterfaceAlert           DeviceInterfaces = 256
	InterfaceOnOff           DeviceInterfaces = 512
	InterfaceLevelCtrl       DeviceInterfaces = 513
	InterfaceColorCtrl       DeviceInterfaces = 514
	InterfaceOpenClose       DeviceInterfaces = 516
	InterfaceOpenCloseConfig DeviceInterfaces = 517
	InterfaceSimpleButton    DeviceInterfaces = 772
	InterfaceOTAUpdate       DeviceInterfaces = 1024
)

type Device struct {
	id           int
	ProductName  string
	Identifier   string
	Manufacturer string
	FwVersion    string
	Name         string
	Description  string
	StateValue   int
	Triggered    bool
	Functions    []DeviceFunction
}

type deviceList struct {
	XMLName   xml.Name `xml:"devicelist"`
	FwVersion string   `xml:"fwversion,attr,omitempty"`
	Version   string   `xml:"version,attr,omitempty"`
	Devices   []device `xml:"device"`
}

type device struct {
	Id              int           `xml:"id,attr,omitempty"` // internal id
	ProductName     string        `xml:"productname,attr,omitempty"`
	Identifier      string        `xml:"identifier,attr,omitempty"` // AIN, MAC
	Manufacturer    string        `xml:"manufacturer,attr,omitempty"`
	FwVersion       string        `xml:"fwversion,attr,omitempty"`
	FunctionBitmask uint32        `xml:"functionbitmask,attr,omitempty"`
	Name            string        `xml:"name"`
	IsLowBattery    *bool         `xml:"batterylow,omitempty"`
	BatteryLevel    *byte         `xml:"battery,omitempty"`
	Present         bool          `xml:"present"`
	TXBusy          bool          `xml:"txbusy"`
	OnOff           *DeviceOnOff  `xml:"simpleonoff,omitempty"`
	Alert           *DeviceAlert  `xml:"alert,omitempty"`
	Button          *DeviceButton `xml:"button,omitempty"`
	UnitInfo        *unitInfo     `xml:"etsiunitinfo,omitempty"`
}

type unitInfo struct {
	DeviceID   int              `xml:"etsideviceid"`
	UnitType   DeviceType       `xml:"unittype"`
	Interfaces DeviceInterfaces `xml:"interfaces"`
}

type DeviceOnOff struct {
	State int `xml:"state"`
}

type DeviceAlert struct {
	State           int   `xml:"state"`
	LastAlertChange int64 `xml:"lastalertchgtimestamp"` // 1752247238
}

type DeviceSwitch struct { // Switchable power outlet
	State      int    `xml:"state"`
	Mode       string `xml:"mode"`
	Lock       int    `xml:"lock"`
	DeviceLock int    `xml:"devicelock"`
}

type DeviceButton struct {
	Identifier  string `xml:"identifier,attr,omitempty"`
	Id          string `xml:"id,attr,omitempty"`
	LastPressed int64  `xml:"lastpressedtimestamp"` // 1752247238
	Name        string `xml:"name,omitempty"`
}

func getDeviceListInfos(fc *fritzClient, s Session) ([]Device, error) {
	resp, err := fc.client.Get(fmt.Sprintf("%s/webservices/homeautoswitch.lua?sid=%s&switchcmd=getdevicelistinfos", fc.baseURL, s.GetSID()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dl deviceList
	if unmarshalErr := xml.Unmarshal(body, &dl); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	idToInternalDevice := map[int]*device{}

	for _, d := range dl.Devices {
		idToInternalDevice[d.Id] = &d
	}

	log.PrintXML(dl)

	var devices []Device

	for _, d := range dl.Devices {
		if d.UnitInfo == nil {
			continue
		}
		relatedDevice := idToInternalDevice[d.UnitInfo.DeviceID]
		if relatedDevice == nil {
			continue
		}
		functions := parseFunctionBitmask(d.FunctionBitmask)
		useName := relatedDevice.Name
		useState := -1
		isTriggered := false
		var descriptionParts []string
		if relatedDevice.Name != d.Name {
			useName = relatedDevice.Name + " (" + d.Name + ")"
		}
		if d.OnOff != nil {
			descriptionParts = append(descriptionParts, fmt.Sprintf("on_off=%d", d.OnOff.State))
			useState = d.OnOff.State
		}
		if d.Alert != nil {
			lastChange := time.Unix(d.Alert.LastAlertChange, 0)
			descriptionParts = append(descriptionParts, fmt.Sprintf("alert=%d, lastchange=%s", d.Alert.State, lastChange.Format(time.DateTime)))
			useState = d.Alert.State
			isTriggered = d.Alert.State == 1
		}
		if d.Button != nil {
			lastPressed := time.Unix(d.Button.LastPressed, 0)
			descriptionParts = append(descriptionParts, "button")
			descriptionParts = append(descriptionParts, fmt.Sprintf("lastpressed=%s", lastPressed.Format(time.DateTime)))
		}

		if relatedDevice.BatteryLevel != nil {
			descriptionParts = append(descriptionParts, fmt.Sprintf("battery=%d%%", *relatedDevice.BatteryLevel))
		}

		useDescription := strings.Join(descriptionParts, ", ")
		current := Device{
			id:           relatedDevice.Id,
			ProductName:  relatedDevice.ProductName,
			Identifier:   relatedDevice.Identifier,
			Manufacturer: relatedDevice.Manufacturer,
			FwVersion:    relatedDevice.FwVersion,
			Name:         useName,
			Description:  useDescription,
			StateValue:   useState,
			Triggered:    isTriggered,
			Functions:    functions,
		}

		devices = append(devices, current)
	}

	return devices, nil
}

func parseFunctionBitmask(bitmask uint32) []DeviceFunction {

	functions := toDeviceFunctions(bitmask)

	return functions
}

func toDeviceFunctions(num uint32) []DeviceFunction {
	var functions []DeviceFunction
	// Prüfe jedes mögliche Bit
	for i := 0; i <= 20; i++ {
		flag := uint32(i) + 1
		if num&(1<<flag) != 0 {
			// Prüfe, ob es eine definierte Funktion für dieses Bit gibt
			f := DeviceFunction(i + 1)
			if isDefinedDeviceFunction(f) {
				functions = append(functions, f)
			}
		}
	}
	return functions
}

func isDefinedDeviceFunction(f DeviceFunction) bool {
	switch f {
	case HANFUNDevice, Light, AlarmSensor, AVMButton,
		AVMHeatingController, AVMPowerMeter, TemperatureSensor,
		AVMOutletSwitch, AVMDECTRepeater, AVMMicrophone,
		HANFUNUnit, SimpleOnOffDevice, DimmableLevelDevice,
		ColorAdjustableLight, Blinds, HumiditySensor:
		return true
	default:
		return false
	}
}

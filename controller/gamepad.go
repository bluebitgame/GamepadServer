package controller

import (
	"time"

	"github.com/boombuler/hid"
	"github.com/go-toast/toast"
)

var HIDDeviceMap map[string]*hid.DeviceInfo

func CreateLocalControllerService() {
	HIDDeviceMap = make(map[string]*hid.DeviceInfo)

	for {
		for dev := range hid.Devices() {
			if dev.VendorId == 0x57e && dev.ProductId == 0x2009 {
				_, ok := HIDDeviceMap[dev.Path]
				if !ok {
					go NewSwitchProController(dev)
				}
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func Notification(title, message string) {
	notification := toast.Notification{
		AppID:   "GamepadServer",
		Title:   title,
		Message: message,
	}

	notification.Push()
}

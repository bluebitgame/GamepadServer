package controller

import (
	"encoding/binary"
	"fmt"

	"github.com/boombuler/hid"
)

const (
	SwitchProControllerButtonY             = 0
	SwitchProControllerButtonX             = 1
	SwitchProControllerButtonB             = 2
	SwitchProControllerButtonA             = 3
	SwitchProControllerButtonRightShoulder = 6
	SwitchProControllerButtonRightTrigger  = 7

	SwitchProControllerButtonMinus      = 0
	SwitchProControllerButtonPlus       = 1
	SwitchProControllerButtonRightThumb = 2
	SwitchProControllerButtonLeftThumb  = 3
	SwitchProControllerButtonHome       = 4
	SwitchProControllerButtonCapture    = 5

	SwitchProControllerButtonDown         = 0
	SwitchProControllerButtonUp           = 1
	SwitchProControllerButtonRight        = 2
	SwitchProControllerButtonLeft         = 3
	SwitchProControllerButtonLeftShoulder = 6
	SwitchProControllerButtonLeftTrigger  = 7
)

func NewSwitchProController(DeviceInfo *hid.DeviceInfo) {
	HIDDeviceMap[DeviceInfo.Path] = DeviceInfo
	defer delete(HIDDeviceMap, DeviceInfo.Path)

	device, err := DeviceInfo.Open()
	if err != nil {
		fmt.Println(err)
	}
	defer device.Close()

	var controller SwitchProController
	controller.Name = DeviceInfo.Manufacturer + " " + DeviceInfo.Product
	controller.Device = device
	err = controller.Init()
	if err != nil {
		Notification("unable to init controller", err.Error())
		return
	}

	Notification(controller.Name, " Connected")

	emulator, err := NewEmulator(func(vibration Vibration) {})
	if err != nil {
		Notification("unable to start ViGEm client", err.Error())
		return
	}
	defer emulator.Close()

	ctr, err := emulator.CreateXbox360Controller()
	if err != nil {
		Notification("unable to create emulated Xbox 360 controller", err.Error())
		return
	}
	defer ctr.Close()

	err = ctr.Connect()
	if err != nil {
		Notification("unable to connect to emulated Xbox 360 controller", err.Error())
		return
	}
	defer ctr.Disconnect()

	for {
		raw_buf, err := controller.Device.Read()
		if err != nil {
			Notification(controller.Name, err.Error())
			return
		}
		if len(raw_buf) < 12 {
			Notification(controller.Name, "Disconnected")
			return
		}

		if raw_buf[0] == 0x30 {
			ButtonsR := raw_buf[3]
			ButtonsM := raw_buf[4]
			ButtonsL := raw_buf[5]

			report := Xbox360ControllerReport{}
			report.MaybeSetButton(Xbox360ControllerButtonX, ButtonsR&(1<<SwitchProControllerButtonY) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonY, ButtonsR&(1<<SwitchProControllerButtonX) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonA, ButtonsR&(1<<SwitchProControllerButtonB) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonB, ButtonsR&(1<<SwitchProControllerButtonA) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonRightShoulder, ButtonsR&(1<<SwitchProControllerButtonRightShoulder) != 0)
			if ButtonsR&(1<<SwitchProControllerButtonRightTrigger) != 0 {
				report.SetRightTrigger(255)
			}

			report.MaybeSetButton(Xbox360ControllerButtonStart, ButtonsM&(1<<SwitchProControllerButtonPlus) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonBack, ButtonsM&(1<<SwitchProControllerButtonMinus) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonRightThumb, ButtonsM&(1<<SwitchProControllerButtonRightThumb) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonLeftThumb, ButtonsM&(1<<SwitchProControllerButtonLeftThumb) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonGuide, ButtonsM&(1<<SwitchProControllerButtonHome) != 0)

			report.MaybeSetButton(Xbox360ControllerButtonUp, ButtonsL&(1<<SwitchProControllerButtonUp) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonDown, ButtonsL&(1<<SwitchProControllerButtonDown) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonLeft, ButtonsL&(1<<SwitchProControllerButtonLeft) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonRight, ButtonsL&(1<<SwitchProControllerButtonRight) != 0)
			report.MaybeSetButton(Xbox360ControllerButtonLeftShoulder, ButtonsL&(1<<SwitchProControllerButtonLeftShoulder) != 0)
			if ButtonsL&(1<<SwitchProControllerButtonLeftTrigger) != 0 {
				report.SetLeftTrigger(255)
			}

			LeftThumbX, LeftThumbY := controller.StickCalLeft.StickCalibrate(uint16(raw_buf[7]&0xF)<<8|uint16(raw_buf[6]), (uint16(raw_buf[8])<<4)|uint16(raw_buf[7]>>4))
			RightThumbX, RightThumbY := controller.StickCalRight.StickCalibrate(uint16(raw_buf[10]&0xF)<<8|uint16(raw_buf[9]), (uint16(raw_buf[11])<<4)|uint16(raw_buf[10]>>4))

			if ButtonsL&(1<<SwitchProControllerButtonLeftTrigger) != 0 {
				var gyr_x float32 = 0.0
				var gyr_y float32 = 0.0
				for i := 0; i < 3; i++ {
					var acc_r, gyr_r [3]int16
					acc_r[0] = int16(binary.LittleEndian.Uint16(raw_buf[13+i*12:]))
					acc_r[1] = int16(binary.LittleEndian.Uint16(raw_buf[15+i*12:]))
					acc_r[2] = int16(binary.LittleEndian.Uint16(raw_buf[17+i*12:]))
					gyr_r[0] = int16(binary.LittleEndian.Uint16(raw_buf[19+i*12:]))
					gyr_r[1] = int16(binary.LittleEndian.Uint16(raw_buf[21+i*12:]))
					gyr_r[2] = int16(binary.LittleEndian.Uint16(raw_buf[23+i*12:]))

					gyr_x += float32(gyr_r[2]) / 2048.0
					gyr_y += float32(gyr_r[1]) / 2048.0
				}

				if gyr_x > 0.0 {
					gyr_x += 0.1
				} else {
					gyr_x -= 0.1
				}
				if gyr_y > 0.0 {
					gyr_y += 0.1
				} else {
					gyr_y -= 0.1
				}
				RightThumbX -= gyr_x
				RightThumbY -= gyr_y
			}

			if LeftThumbX > 1.0 {
				LeftThumbX = 1.0
			} else if LeftThumbX < -1.0 {
				LeftThumbX = -1.0
			}

			if LeftThumbY > 1.0 {
				LeftThumbY = 1.0
			} else if LeftThumbY < -1.0 {
				LeftThumbY = -1.0
			}

			if RightThumbX > 1.0 {
				RightThumbX = 1.0
			} else if RightThumbX < -1.0 {
				RightThumbX = -1.0
			}

			if RightThumbY > 1.0 {
				RightThumbY = 1.0
			} else if RightThumbY < -1.0 {
				RightThumbY = -1.0
			}

			report.SetLeftThumb(int16(LeftThumbX*32767), int16(LeftThumbY*32767))
			report.SetRightThumb(int16(RightThumbX*32767), int16(RightThumbY*32767))

			ctr.Send(&report)
		}
	}
}

type SwitchProController struct {
	Name          string
	Device        hid.Device
	CommmandID    byte
	StickCalLeft  StickCalibration
	StickCalRight StickCalibration
	AccNeutral    [3]uint16
	AccSensiti    [3]uint16
	GyrNeutral    [3]uint16
	GyrSensiti    [3]uint16
}

func (Controller *SwitchProController) Init() error {
	Controller.CommmandID = 0

	//Blink Home Light
	_, err := Controller.Subcommand(0x38, []byte{0x1F, 0xFF, 0x00})
	if err != nil {
		return err
	}

	// Request Factory Calibration Data
	response, err := Controller.ReadSPI([]byte{0x3d, 0x60, 0x00, 0x00, 9})
	if err != nil {
		return err
	}
	Controller.StickCalLeft.MaxX = uint16(response[1]&0xF)<<8 | uint16(response[0])
	Controller.StickCalLeft.MaxY = (uint16(response[2]) << 4) | uint16(response[1]>>4)
	Controller.StickCalLeft.CenterX = uint16(response[4]&0xF)<<8 | uint16(response[3])
	Controller.StickCalLeft.CenterY = (uint16(response[5]) << 4) | uint16(response[4]>>4)
	Controller.StickCalLeft.MinX = uint16(response[7]&0xF)<<8 | uint16(response[6])
	Controller.StickCalLeft.MinY = (uint16(response[8]) << 4) | uint16(response[7]>>4)

	response, err = Controller.ReadSPI([]byte{0x86, 0x60, 0x00, 0x00, 16})
	if err != nil {
		return err
	}
	Controller.StickCalLeft.DeadZone = uint16(response[4]&0xF)<<8 | uint16(response[3])

	// Request Factory Calibration Data
	response, err = Controller.ReadSPI([]byte{0x46, 0x60, 0x00, 0x00, 9})
	if err != nil {
		return err
	}
	Controller.StickCalRight.MaxX = uint16(response[7]&0xF)<<8 | uint16(response[6])
	Controller.StickCalRight.MaxY = (uint16(response[8]) << 4) | uint16(response[7]>>4)
	Controller.StickCalRight.CenterX = uint16(response[1]&0xF)<<8 | uint16(response[0])
	Controller.StickCalRight.CenterY = (uint16(response[2]) << 4) | uint16(response[1]>>4)
	Controller.StickCalRight.MinX = uint16(response[4]&0xF)<<8 | uint16(response[3])
	Controller.StickCalRight.MinY = (uint16(response[5]) << 4) | uint16(response[4]>>4)

	response, err = Controller.ReadSPI([]byte{0x98, 0x60, 0x00, 0x00, 16})
	if err != nil {
		return err
	}
	Controller.StickCalRight.DeadZone = uint16(response[4]&0xF)<<8 | uint16(response[3])

	response, err = Controller.ReadSPI([]byte{0x20, 0x60, 0x00, 0x00, 10})
	if err != nil {
		return err
	}
	Controller.AccNeutral[0] = binary.LittleEndian.Uint16(response[0:2])
	Controller.AccNeutral[1] = binary.LittleEndian.Uint16(response[2:4])
	Controller.AccNeutral[2] = binary.LittleEndian.Uint16(response[4:6])

	response, err = Controller.ReadSPI([]byte{0x26, 0x60, 0x00, 0x00, 10})
	if err != nil {
		return err
	}
	Controller.AccSensiti[0] = binary.LittleEndian.Uint16(response[0:2])
	Controller.AccSensiti[1] = binary.LittleEndian.Uint16(response[2:4])
	Controller.AccSensiti[2] = binary.LittleEndian.Uint16(response[4:6])

	response, err = Controller.ReadSPI([]byte{0x2C, 0x60, 0x00, 0x00, 10})
	if err != nil {
		return err
	}
	Controller.GyrNeutral[0] = binary.LittleEndian.Uint16(response[0:2])
	Controller.GyrNeutral[1] = binary.LittleEndian.Uint16(response[2:4])
	Controller.GyrNeutral[2] = binary.LittleEndian.Uint16(response[4:6])

	response, err = Controller.ReadSPI([]byte{0x32, 0x60, 0x00, 0x00, 10})
	if err != nil {
		return err
	}
	Controller.GyrSensiti[0] = binary.LittleEndian.Uint16(response[0:2])
	Controller.GyrSensiti[1] = binary.LittleEndian.Uint16(response[2:4])
	Controller.GyrSensiti[2] = binary.LittleEndian.Uint16(response[4:6])

	// Set Player LED
	_, err = Controller.Subcommand(0x30, []byte{1})
	if err != nil {
		return err
	}
	// Set IMU
	_, err = Controller.Subcommand(0x40, []byte{0x1, 1})
	if err != nil {
		return err
	}
	_, err = Controller.Subcommand(0x48, []byte{0x1})
	if err != nil {
		return err
	}
	_, err = Controller.Subcommand(0x3, []byte{0x30})
	if err != nil {
		return err
	}

	//Home Light
	_, err = Controller.Subcommand(0x38, []byte{0x1F, 0x60, 0x60})
	if err != nil {
		return err
	}

	return nil
}

func (Controller *SwitchProController) Subcommand(sc byte, cmd []byte) ([]byte, error) {
	header := []byte{0x1, Controller.CommmandID, 0x0, 0x1, 0x40, 0x40, 0x0, 0x1, 0x40, 0x40, sc}
	Controller.CommmandID++
	err := Controller.Device.Write(append(header, cmd...))
	if err != nil {
		return nil, err
	}
	for i := 0; i < 10; i++ {
		response, err := Controller.Device.Read()
		if err != nil {
			return nil, err
		}
		if response[0] == 0x21 && response[14] == sc {
			return response, nil
		}
	}
	return nil, nil
}

func (Controller *SwitchProController) ReadSPI(cmd []byte) ([]byte, error) {
	for i := 0; i < 10; i++ {
		response, err := Controller.Subcommand(0x10, cmd)
		if err != nil {
			return nil, err
		}
		if response[15] == cmd[0] && response[16] == cmd[1] {
			return response[20:], nil
		}
	}

	return nil, nil
}

type StickCalibration struct {
	MaxX     uint16
	MaxY     uint16
	CenterX  uint16
	CenterY  uint16
	MinX     uint16
	MinY     uint16
	DeadZone uint16
}

func (StickCal StickCalibration) StickCalibrate(ThumbX uint16, ThumbY uint16) (float32, float32) {
	x := float32(ThumbX) - float32(StickCal.CenterX)
	y := float32(ThumbY) - float32(StickCal.CenterY)
	if x*x+y*y < float32(StickCal.DeadZone*StickCal.DeadZone) {
		return 0.0, 0.0
	}

	if x > 0.0 {
		x /= float32(StickCal.MaxX)
	} else {
		x /= float32(StickCal.MinX)
	}

	if y > 0.0 {
		y /= float32(StickCal.MaxY)
	} else {
		y /= float32(StickCal.MinY)
	}

	return x, y
}

package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func CreateWebControllerService(Addr string) *http.Server {
	setupRoutes()
	return &http.Server{Addr: ":3080"}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

type ControllerMsg struct {
	Buttons []byte  `json:"buttons"`
	Axes    []int16 `json:"axes"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Notification("Gamepad Server", err.Error())
		return
	}
	defer conn.Close()

	msgType, msgJson, err := conn.ReadMessage()
	if err != nil {
		Notification("Bad message", string(msgJson)+"\r\n"+err.Error())
		return
	}
	ClientName := string(msgJson)
	Notification(ClientName+" Connected", conn.RemoteAddr().String())

	emulator, err := NewEmulator(func(vibration Vibration) {})
	if err != nil {
		Notification("unable to start ViGEm client", err.Error())
	}
	defer emulator.Close()

	ctr, err := emulator.CreateXbox360Controller()
	if err != nil {
		Notification("unable to create emulated Xbox 360 controller", err.Error())
	}
	defer ctr.Close()

	err = ctr.Connect()
	if err != nil {
		Notification("unable to connect to emulated Xbox 360 controller", err.Error())
		return
	}
	defer ctr.Disconnect()

	if err = conn.WriteMessage(msgType, msgJson); err != nil {
		return
	}

	for {
		_, msgJson, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg ControllerMsg
		err = json.Unmarshal(msgJson, &msg)
		if err != nil {
			Notification("Bad message", string(msgJson)+"\r\n"+err.Error())
			return
		}

		report := Xbox360ControllerReport{}
		for i := 0; i < len(msg.Buttons) && i < 17; i++ {
			report.SetButtonFromGamePadAPI(i, msg.Buttons[i])
		}

		if len(msg.Axes) >= 2 {
			report.SetLeftThumb(msg.Axes[0], -msg.Axes[1])
		}

		if len(msg.Axes) >= 4 {
			report.SetRightThumb(msg.Axes[2], -msg.Axes[3])
		}

		ctr.Send(&report)
	}
}

func setupRoutes() {
	fileServer := http.FileServer(http.Dir("./web/"))
	http.Handle("/", fileServer)
	http.HandleFunc("/ws", wsEndpoint)
}

func (r *Xbox360ControllerReport) SetButtonFromGamePadAPI(button int, value byte) {
	switch button {
	case 0:
		r.MaybeSetButton(Xbox360ControllerButtonA, value != 0)
	case 1:
		r.MaybeSetButton(Xbox360ControllerButtonB, value != 0)
	case 2:
		r.MaybeSetButton(Xbox360ControllerButtonX, value != 0)
	case 3:
		r.MaybeSetButton(Xbox360ControllerButtonY, value != 0)
	case 4:
		r.MaybeSetButton(Xbox360ControllerButtonLeftShoulder, value != 0)
	case 5:
		r.MaybeSetButton(Xbox360ControllerButtonRightShoulder, value != 0)
	case 6:
		r.SetLeftTrigger(value)
	case 7:
		r.SetRightTrigger(value)
	case 8:
		r.MaybeSetButton(Xbox360ControllerButtonBack, value != 0)
	case 9:
		r.MaybeSetButton(Xbox360ControllerButtonStart, value != 0)
	case 10:
		r.MaybeSetButton(Xbox360ControllerButtonLeftThumb, value != 0)
	case 11:
		r.MaybeSetButton(Xbox360ControllerButtonRightThumb, value != 0)
	case 12:
		r.MaybeSetButton(Xbox360ControllerButtonUp, value != 0)
	case 13:
		r.MaybeSetButton(Xbox360ControllerButtonDown, value != 0)
	case 14:
		r.MaybeSetButton(Xbox360ControllerButtonLeft, value != 0)
	case 15:
		r.MaybeSetButton(Xbox360ControllerButtonRight, value != 0)
	case 16:
		r.MaybeSetButton(Xbox360ControllerButtonGuide, value != 0)
	}
}

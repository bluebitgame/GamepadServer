package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"./controller"
	"./icon"
	"github.com/getlantern/systray"
	"github.com/go-toast/toast"
	"github.com/gorilla/websocket"
	"github.com/rodolfoag/gow32"
)

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
		notification("Gamepad Server", err.Error())
		return
	}
	defer conn.Close()

	msgType, msgJson, err := conn.ReadMessage()
	if err != nil {
		notification("Bad message", string(msgJson)+"\r\n"+err.Error())
		return
	}
	ClientName := string(msgJson)
	notification(ClientName+" Connected", conn.RemoteAddr().String())

	emulator, err := controller.NewEmulator(func(vibration controller.Vibration) {})
	if err != nil {
		notification("unable to start ViGEm client", err.Error())
	}
	defer emulator.Close()

	ctr, err := emulator.CreateXbox360Controller()
	if err != nil {
		notification("unable to create emulated Xbox 360 controller", err.Error())
	}
	defer ctr.Close()

	err = ctr.Connect()
	if err != nil {
		notification("unable to connect to emulated Xbox 360 controller", err.Error())
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
			notification("Bad message", string(msgJson)+"\r\n"+err.Error())
			return
		}

		report := controller.Xbox360ControllerReport{}
		for i := 0; i < len(msg.Buttons) && i < 17; i++ {
			report.SetButtonFromGamePad(i, msg.Buttons[i])
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

var HTTPServer *http.Server

func main() {
	_, err := gow32.CreateMutex("GamepadServer")
	if err != nil {
		notification("", "GamepadServer is already running!")
	} else {
		systray.Run(onReady, onExit)
	}
}

func onReady() {
	systray.SetIcon(icon.Data)
	mQuit := systray.AddMenuItem("Quit", "")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	setupRoutes()
	HTTPServer = &http.Server{Addr: ":3080"}
	HTTPServer.ListenAndServe()
}

func onExit() {
	HTTPServer.Close()
}

func notification(title, message string) {
	notification := toast.Notification{
		AppID:   "GamepadServer",
		Title:   title,
		Message: message,
	}

	notification.Push()
}

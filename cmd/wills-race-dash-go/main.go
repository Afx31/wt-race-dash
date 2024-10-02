package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"go.einride.tech/can/pkg/socketcan"
)

var addr = flag.String("addr", ":8080", "http service address")
const configCanDevice = "vcan0"
const configStopDataloggingId = uint32(105)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type CanData struct {
	Rpm uint16
	Speed uint16
	Gear uint8
	Voltage uint8
	Iat uint16
	Ect uint16
	Tps uint16
	Map uint16
	LambdaRatio uint16
	OilTemp uint16
	OilPressure uint16
}

func handleWs(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Error upgrading WebSocket: ", err)
        return
    }
    defer conn.Close()

		// ---------- CANBus data ----------
		canConn, _ := socketcan.DialContext(context.Background(), "can", configCanDevice)
		defer canConn.Close()
		canRecv:= socketcan.NewReceiver(canConn)

		canData := CanData{}

		for canRecv.Receive() {
			frame := canRecv.Frame()

			switch frame.ID {
				case 660:
				//case 1632:
					canData.Rpm = binary.BigEndian.Uint16(frame.Data[0:2])
					canData.Speed = binary.BigEndian.Uint16(frame.Data[2:4])
					canData.Gear = frame.Data[4]
					canData.Voltage = frame.Data[5] / 10
			}

			jsonData, err := json.Marshal(canData)
			if err != nil {
				log.Println("Json Marshal: ", err)
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				fmt.Println("Write error: ", err)
				return
			}
		}

		// ----------------------------------------
    // for {
    //     // Reading message
    //     // messageType, p, err := conn.ReadMessage()
    //     // if err != nil {
    //     //     log.Println("Read error: ", err)
    //     //     return
    //     // }
    //     // fmt.Println("Received: ", string(p))

    //     // Writing message back
		// 		message := []byte("Test message from backend")
		// 		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		// 			log.Println("Write error:", err)
		// 			return
		// 	}
    // }
}

func main() {
    fmt.Println("--- Server running ---")

    // Serve all static files from the 'web' directory
    fs := http.FileServer(http.Dir("../../web"))
    http.Handle("/", fs)

    // Handle WebSocket connection
    http.HandleFunc("/ws", handleWs)

		fmt.Println("Server starting at :8080")
		err := http.ListenAndServe(*addr, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
}

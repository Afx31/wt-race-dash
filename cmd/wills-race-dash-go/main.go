package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
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

// Data conversion consts
// Oil Temp
const A = 0.0014222095
const B = 0.00023729017
const C = 9.3273998E-8
// Oil Pressure
const originalLow float64 = 0 //0.5
const originalHigh float64 = 5 //4.5
const desiredLow float64 = -100 //0
const desiredHigh float64 = 1100 //1000

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
				// case 1632:
					canData.Rpm = binary.BigEndian.Uint16(frame.Data[0:2])
					canData.Speed = binary.BigEndian.Uint16(frame.Data[2:4])
					canData.Gear = frame.Data[4]
					canData.Voltage = frame.Data[5] / 10
				case 661:
				// case 1633:
					canData.Iat = binary.BigEndian.Uint16(frame.Data[0:2])
					canData.Ect = binary.BigEndian.Uint16(frame.Data[2:4])
				case 662:
				// case 1634:
					canData.Tps = binary.BigEndian.Uint16(frame.Data[0:2])
					canData.Map = binary.BigEndian.Uint16(frame.Data[2:4]) / 10
				case 664:
				// case 1636:
					canData.LambdaRatio = 32768 / binary.BigEndian.Uint16(frame.Data[0:2])
				case 667:
				// case 1639:
					// Oil Temp
					oilTempResistance := binary.BigEndian.Uint16(frame.Data[0:2])
					kelvinTemp := 1 / (A + B * math.Log(float64(oilTempResistance)) + C * math.Pow(math.Log(float64(oilTempResistance)), 3))
					canData.OilTemp = uint16(kelvinTemp - 273.15)
		
					// Oil Pressure
					oilPressureResistance := float64(binary.BigEndian.Uint16(frame.Data[2:4])) / 819.2
					kPaValue := ((float64(oilPressureResistance) - originalLow) / (originalHigh - originalLow) * (desiredHigh - desiredLow)) + desiredLow
					canData.OilPressure = uint16(math.Round(kPaValue * 0.145038)) // Convert to psi
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

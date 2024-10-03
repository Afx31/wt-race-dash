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
	// "time"

	"github.com/gorilla/websocket"
	"go.einride.tech/can/pkg/socketcan"
	"github.com/stratoberry/go-gpsd"
)

var addr = flag.String("addr", ":8080", "http service address")
const configCanDevice = "vcan0"
const configStopDataloggingId = uint32(105)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

// Data conversion constants
const (
	// Oil Temp
	A = 0.0014222095
	B = 0.00023729017
	C = 9.3273998E-8
	// Oil Pressure
	originalLow float64 = 0 //0.5
	originalHigh float64 = 5 //4.5
	desiredLow float64 = -100 //0
	desiredHigh float64 = 1100 //1000
)

type CanData struct {
	Type uint8
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

type GpsData struct {
	Type uint8
	Lon float64
	Lat float64
	Time string //time.Time
}

func handleGps(conn *websocket.Conn) {
	// Connect to the GPSD server
	gps, err := gpsd.Dial("localhost:2947")
	if err != nil {
		log.Fatal("Failed to connect to GPSD: %v", err)
	}

	gpsData := GpsData{}
	gpsData.Type = 2

	// Define a reporting filter
	tpvFilter := func(r interface{}) {
		report := r.(*gpsd.TPVReport)
		// fmt.Println(report.Time, " | ", report.Lon, " | ", report.Lat)
		gpsData.Lon = report.Lon
		gpsData.Lat = report.Lat
		gpsData.Time = fmt.Sprintf("%0.2d:%0.2d:%0.2d:%0.3d", report.Time.Hour(), report.Time.Minute(), report.Time.Second(), report.Time.Nanosecond()/1000000)

		jsonData, err := json.Marshal(gpsData)
		if err != nil {
			log.Fatal("Json Marshall error (GPS): ", err)
		}

		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			fmt.Println("Write error (GPS): ", err)
			return
		}
	}

	gps.AddFilter("TPV", tpvFilter)

	done := gps.Watch()
	<-done
	// .. some time later ....
	gps.Close()
}

func handleWs(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Error upgrading WebSocket: ", err)
        return
    }
    defer conn.Close()


		// ---------- GPS ----------
		go handleGps(conn)


		// ---------- CANBus data ----------
		canConn, _ := socketcan.DialContext(context.Background(), "can", configCanDevice)
		defer canConn.Close()
		canRecv:= socketcan.NewReceiver(canConn)

		canData := CanData{}
		canData.Type = 1

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
				log.Println("Json Marshal error (CAN): ", err)
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				fmt.Println("Write error (CAN): ", err)
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

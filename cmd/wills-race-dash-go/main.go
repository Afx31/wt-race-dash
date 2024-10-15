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
	"time"

	"wills-race-dash-go/internal/tracks"

	"github.com/gorilla/websocket"
	"github.com/stratoberry/go-gpsd"
	"go.einride.tech/can/pkg/socketcan"
)

var wsConn *websocket.Conn
var addr = flag.String("addr", ":8080", "http service address")
var configCanDevice = "vcan0"
var configStopDataloggingId = uint32(105)

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

type CurrentLapData struct {
	Type int8
  LapStartTime time.Time
	CurrentLapTime int32
}

type LapStats struct {
  Type int8
  LapCounter int8
	BestLapTime int32
	PbLapTime int32
	PreviousLapTime int32
}

var lapStats = LapStats{Type: 3, LapCounter: 1}
// **********************************************************************************************************

// Testing
func containsCurrentCoordinates(arr []float64, coordinate float64) bool {
  for _, v := range arr {
    if v == coordinate {
      return true
    }
  }
  return false
}

func containsCurrentCoordinates2(min float64, max float64, current float64) bool {
  if (min < current && current > max) {
    return true
  }
  return false
}

// -- How to do this --
// We've defined the tracks start line coordinates
// Now we need to bring in the current GPS location, tak it and iterate to see if it's the tracks one
func (currentLapData *CurrentLapData) startFinishLineDetection(currentLat float64, currentLon float64, currentTime time.Time) {
  timeDiff := currentTime.Sub(currentLapData.LapStartTime)
  currentLapData.CurrentLapTime = int32(timeDiff.Milliseconds())

  // Testing: This will only go off the actual points
  // if containsCurrentCoordinates(tracks.TestLat[:], currentLat) && containsCurrentCoordinates(tracks.TestLon[:], currentLon) {
  //   fmt.Println("yay1")
  // }
  
  if containsCurrentCoordinates2(tracks.TestLatMin, tracks.TestLatMax, currentLat) {
    if containsCurrentCoordinates2(tracks.TestLonMin, tracks.TestLonMax, currentLon) {
      if currentLapData.CurrentLapTime < lapStats.BestLapTime || lapStats.BestLapTime == 0 {
        lapStats.BestLapTime = currentLapData.CurrentLapTime
      }
      if currentLapData.CurrentLapTime < lapStats.PbLapTime || lapStats.PbLapTime == 0 {
        lapStats.PbLapTime = currentLapData.CurrentLapTime
      }
      lapStats.PreviousLapTime = currentLapData.CurrentLapTime
      
      // Start the next lap
      currentLapData.LapStartTime = currentTime
      lapStats.LapCounter++;

      // Send up to client
      jsonData, err := json.Marshal(lapStats)
      if err != nil {
        log.Fatal("Json Marshall error (Lap Stats)")
      }
      if err := wsConn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
        fmt.Println("Write error (Lap Stats): ", err)
      }
    }
  }
}


func handleGpsLapTiming() {
	// Connect to the GPSD server
	gps, err := gpsd.Dial("localhost:2947")
	if err != nil {
		log.Fatal("Failed to connect to GPSD: ", err)
	}

	currentLapData := CurrentLapData{Type: 2}
  currentLapData.LapStartTime = time.Now().Round(100 * time.Millisecond)
  //currentLapData.CurrentLapStartTime = time.Now().Format("15:04:05.000000000")
    
	// Define a reporting filter
	tpvFilter := func(r interface{}) {
		report := r.(*gpsd.TPVReport)
    
    // ----- Convert report.Time from UTC to Australia/Sydney -----
    location, err := time.LoadLocation("Australia/Sydney")
    if err != nil {
      fmt.Println("Error loading location:", err)
      return
    }
    convertedTime := report.Time.In(location)

    // Main bulk of it for GPS/Lap Timing
    currentLapData.startFinishLineDetection(report.Lat, report.Lon, convertedTime)

		jsonData, err := json.Marshal(currentLapData)
		if err != nil {
			log.Fatal("Json Marshall error (GPS): ", err)
		}
    // Send up to client
		if err := wsConn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
        fmt.Println("Unexpected close error: ", err)
      } else {
        fmt.Println("Write error (GPS): ", err)
      }
      wsConn.Close()
			return
		}
	}
  
	gps.AddFilter("TPV", tpvFilter)
	done := gps.Watch()
	<-done
	gps.Close()
}


func handleWs(w http.ResponseWriter, r *http.Request) {
  var err error
  wsConn, err = upgrader.Upgrade(w, r, nil)
  if err != nil {
      log.Println("Error upgrading WebSocket: ", err)
      return
  }
  defer wsConn.Close()

  // ---------- Lap Timing ----------
  go handleGpsLapTiming()

  // ---------- CANBus data ----------
  canConn, _ := socketcan.DialContext(context.Background(), "can", configCanDevice)
  defer canConn.Close()
  canRecv:= socketcan.NewReceiver(canConn)

  canData := CanData{Type: 1}

  for canRecv.Receive() {
    frame := canRecv.Frame()
    
    switch frame.ID {
      case 660, 1632:
        canData.Rpm = binary.BigEndian.Uint16(frame.Data[0:2])
        canData.Speed = binary.BigEndian.Uint16(frame.Data[2:4])
        canData.Gear = frame.Data[4]
        canData.Voltage = frame.Data[5] / 10
      case 661, 1633:
        canData.Iat = binary.BigEndian.Uint16(frame.Data[0:2])
        canData.Ect = binary.BigEndian.Uint16(frame.Data[2:4])
      case 662, 1634:
        canData.Tps = binary.BigEndian.Uint16(frame.Data[0:2])
				if canData.Tps == 65535 { canData.Tps = 0	}
        canData.Map = binary.BigEndian.Uint16(frame.Data[2:4]) / 10
      case 664, 1636:
        canData.LambdaRatio = 32768 / binary.BigEndian.Uint16(frame.Data[0:2])
      case 667, 1639:
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
    if err := wsConn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
      fmt.Println("Write error (CAN): ", err)
      return
    }
  }
}


func main() {
    fmt.Println("---------- Server running ----------")

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

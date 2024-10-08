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
  CurrentLapStartTime time.Time
	CurrentLapTime int32
	BestLapTime int32
	PbLapTime int32
	PreviousLapTime int32
}

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
func (gpsData *GpsData) startFinishLineDetection(currentLat float64, currentLon float64, currentTime time.Time) {
  // Current lap time
  timeDiff := currentTime.Sub(gpsData.CurrentLapStartTime)
  gpsData.CurrentLapTime = int32(timeDiff.Milliseconds())
  
  // This will only go off the actual points
  // if containsCurrentCoordinates(tracks.TestLat[:], currentLat) && containsCurrentCoordinates(tracks.TestLon[:], currentLon) {
  //   fmt.Println("yay1")
  // }
  // Need to create a range of min/max of the lat/lon and thats our range to fall within
  if containsCurrentCoordinates2(tracks.TestLatMin, tracks.TestLatMax, currentLat) {
    if containsCurrentCoordinates2(tracks.TestLonMin, tracks.TestLonMax, currentLon) {
      if gpsData.CurrentLapTime < gpsData.BestLapTime || gpsData.BestLapTime == 0 {
        gpsData.BestLapTime = gpsData.CurrentLapTime
      }
      if gpsData.CurrentLapTime < gpsData.PbLapTime || gpsData.PbLapTime == 0 {
        gpsData.PbLapTime = gpsData.CurrentLapTime
      }
      gpsData.PreviousLapTime = gpsData.CurrentLapTime
      
      // Start the next lap
      gpsData.CurrentLapStartTime = currentTime
    }
  }
}

func handleGpsLapTiming(conn *websocket.Conn) {
	// Connect to the GPSD server
	gps, err := gpsd.Dial("localhost:2947")
	if err != nil {
		log.Fatal("Failed to connect to GPSD: %v", err)
	}

	gpsData := GpsData{}
	gpsData.Type = 2
  // TODO: maybe preset the times with System.Time first off?
  //gpsData.CurrentLaptStartTime = 
  
  //time.Now().Format("15:04:05.000000000")
  //gpsData.CurrentLapStartTime = time.Now().Format("15:04:05.000000000")
  gpsData.CurrentLapStartTime = time.Now().Round(100 * time.Millisecond)
  
	// Define a reporting filter
	tpvFilter := func(r interface{}) {
		report := r.(*gpsd.TPVReport)
		gpsData.Lat = report.Lat
    gpsData.Lon = report.Lon
    
    // ----- Convert report.Time from UTC to Australia/Sydney -----
    location, err := time.LoadLocation("Australia/Sydney")
    if err != nil {
      fmt.Println("Error loading location:", err)
      return
    }
    convertedTime := report.Time.In(location)

    // Main bulk of it for GPS/Lap Timing
    gpsData.startFinishLineDetection(report.Lat, report.Lon, convertedTime)


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
	gps.Close()
}

func handleWs(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Error upgrading WebSocket: ", err)
        return
    }
    defer conn.Close()


		// ---------- Lap Timing ----------
		go handleGpsLapTiming(conn)


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

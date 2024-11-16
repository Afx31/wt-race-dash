package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"wills-race-dash-go/internal/tracks"

	"github.com/gorilla/websocket"
	"github.com/stratoberry/go-gpsd"
	"go.einride.tech/can/pkg/socketcan"
)

type AppSettings struct {
  CanChannel string `json:"canChannel"`
  Track string `json:"track"`
	LoggingHertz int `json:"loggingHertz"`
  Car string `json:"car"`
}

type MySocket struct {
  conn *websocket.Conn
  mutex sync.Mutex
}

type CanData struct {
	Type int8
	Rpm uint16
	Speed uint16
	Gear uint8
	Voltage float32
	Iat uint16
	Ect uint16
	Tps uint16
	Map uint16
	LambdaRatio float64
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
  LapCount int8
	BestLapTime int32
	PbLapTime int32
	PreviousLapTime int32
}

var (
  appSettings *AppSettings
  addr = flag.String("addr", ":8080", "http service address")
  upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
  }
  lapStats = LapStats{Type: 3, LapCount: 0}
  canData = CanData{Type: 1}
  currentTrack tracks.Track

  // --- Data conversion constants ---
  // Oil Temp
	A = 0.0014222095
	B = 0.00023729017
	C = 9.3273998E-8
	// Oil Pressure
	originalLow float64 = 0 //0.5
	originalHigh float64 = 5 //4.5
	desiredLow float64 = -100 //0
	desiredHigh float64 = 1100 //1000

	// Testing
	stopChanDatalogging = make(chan struct{})   // Channel to signal the goroutine to stop
	wg sync.WaitGroup  // WaitGroup to ensure clean shutdown of goroutine
)
// **********************************************************************************************************

func isThisTheFinishLine(min float64, max float64, current float64) bool {
  return current >= min && current <= max
}



func (wsConn *MySocket) writeToClient(writeType int8, data []byte) {
  wsConn.mutex.Lock()
  defer wsConn.mutex.Unlock()

  err := wsConn.conn.WriteMessage(websocket.TextMessage, data)
  if err != nil {
    fmt.Println("Write error: ", writeType, err)
    return
  }
}

func (wsConn *MySocket) handleGpsLapTiming() {
	// Connect to the GPSD server
	gps, err := gpsd.Dial("localhost:2947")
	if err != nil {
		fmt.Println("Failed to connect to GPSD: ", err)
		return
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
    convertedCurrentTime := report.Time.In(location)

	
    // ---------- GPS/Lap Timing ----------
    timeDiff := convertedCurrentTime.Sub(currentLapData.LapStartTime)
    currentLapData.CurrentLapTime = int32(timeDiff.Milliseconds())

    // Testing
    //fmt.Println(report.Lat, ", ", report.Lon)

    if isThisTheFinishLine(currentTrack.LatMin, currentTrack.LatMax, report.Lat) && isThisTheFinishLine(currentTrack.LonMin, currentTrack.LonMax, report.Lon) {
			// Do lap stats
      if (currentLapData.CurrentLapTime < lapStats.BestLapTime) || lapStats.BestLapTime == 0 {
        lapStats.BestLapTime = currentLapData.CurrentLapTime
      }
      if (currentLapData.CurrentLapTime < lapStats.PbLapTime) || lapStats.PbLapTime == 0 {
        lapStats.PbLapTime = currentLapData.CurrentLapTime
      }
      lapStats.PreviousLapTime = currentLapData.CurrentLapTime
      
      // Start the next lap
      currentLapData.LapStartTime = convertedCurrentTime
      lapStats.LapCount++;

      // Send up to client
      jsonData, err := json.Marshal(lapStats)
      if err != nil {
        log.Fatal("Json Marshall error (Lap Stats)")
      }
      wsConn.writeToClient(lapStats.Type, jsonData)

			// TESTING
			time.Sleep(10 * time.Second)
    }

		// jsonData2, err := json.Marshal(lapStats)
		// if err != nil {
		// 	log.Fatal("Json Marshall error (Lap Stats)")
		// }
		// wsConn.writeToClient(3, jsonData2)

		jsonData, err := json.Marshal(currentLapData)
		if err != nil {
			log.Fatal("Json Marshall error (GPS): ", err)
		}
    wsConn.writeToClient(currentLapData.Type, jsonData)
	}
  
	gps.AddFilter("TPV", tpvFilter)
	done := gps.Watch()
	<-done
	gps.Close()
}


func (wsConn *MySocket) handleCanBusData() {
  // ---------- CANBus data ----------
  canConn, _ := socketcan.DialContext(context.Background(), "can", appSettings.CanChannel)
  defer canConn.Close()
  canRecv:= socketcan.NewReceiver(canConn)

  // ---------- Datalogging ----------
	wg.Add(1)
  go func() {
    defer wg.Done()
    cmd := exec.Command("/home/pi/dev/wt-datalogging/bin/wt-datalogging")
    output, err := cmd.Output()
    if err != nil {
      fmt.Println("Error running datalogging: ", err)
    }
    fmt.Println(string(output))
  }()

  for canRecv.Receive() {
    frame := canRecv.Frame()
    
    switch frame.ID {
      // case 69, 105:        
      case 660, 1632:
        canData.Rpm = binary.BigEndian.Uint16(frame.Data[0:2])
        canData.Speed = binary.BigEndian.Uint16(frame.Data[2:4])
        canData.Gear = frame.Data[4]
        canData.Voltage = float32(frame.Data[5]) / 10.0
      case 661, 1633:
        canData.Iat = binary.BigEndian.Uint16(frame.Data[0:2])
        canData.Ect = binary.BigEndian.Uint16(frame.Data[2:4])
      case 662, 1634:
        canData.Tps = binary.BigEndian.Uint16(frame.Data[0:2])
				if canData.Tps == 65535 { canData.Tps = 0	}
        canData.Map = binary.BigEndian.Uint16(frame.Data[2:4]) / 10
      case 664, 1636:
        canData.LambdaRatio = math.Round(float64(32768.0) / float64(binary.BigEndian.Uint16(frame.Data[0:2])) * 100) / 100
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
    wsConn.writeToClient(canData.Type, jsonData)
  }
}


func handleWs(w http.ResponseWriter, r *http.Request) {
  // ---------- Web Socket Setup ----------
  wsConn := MySocket{}
  var err error
  wsConn.conn, err = upgrader.Upgrade(w, r, nil)
  if err != nil {
      log.Println("Error upgrading WebSocket: ", err)
      return
  }
  defer wsConn.conn.Close()

  // ===============================================================  
  // ---------- Handle CAN data ----------
  wg.Add(1)
  go func() {
    defer wg.Done()
    wsConn.handleCanBusData()
  }()

  // ---------- Lap Timing ----------
  wg.Add(1)
  go func() {
    defer wg.Done()
    wsConn.handleGpsLapTiming()
  }()

  wg.Wait()
  // ===============================================================
}


func main() {
  fmt.Println("---------- Server running ----------")

	// -------------------- Read in settings file first --------------------
	settingsFile, err := os.Open("/home/pi/dev/wt-racedash-settings.json")
	if err != nil {
		log.Fatal("Error: Cannot read in settings file")
	}
	defer settingsFile.Close()

	data, _ := io.ReadAll(settingsFile)
  json.Unmarshal(data, &appSettings)
  currentTrack = tracks.Tracks[appSettings.Track]
  // ---------------------------------------------------------------------

  // Serve all static files from the 'web' directory
  fs := http.FileServer(http.Dir("../web"))
  http.Handle("/", fs)

  // Handle WebSocket connection
  http.HandleFunc("/ws", handleWs)

  fmt.Println("Server starting at :8080")
  err = http.ListenAndServe(*addr, nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

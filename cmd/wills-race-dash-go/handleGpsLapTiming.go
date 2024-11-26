package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"wills-race-dash-go/internal/tracks"

	"github.com/stratoberry/go-gpsd"
)

type CurrentLapData struct {
	Type           int8
	LapStartTime   time.Time
	CurrentLapTime uint32

	// TODO: Testing
  PreviousLat    float64
	PreviousLon    float64
}

type LapStats struct {
	Type            int8
	LapCount        uint8
	BestLapTime     uint32
	PbLapTime       uint32
	PreviousLapTime uint32
}

var (
	lapStats      = LapStats{Type: 3, LapCount: 0}
	currentTrack  tracks.Track
)

func isThisTheFinishLine(min float64, max float64, current float64) bool {
	//fmt.Println(min, max, current)
	return current >= min && current <= max
}

// TODO: Testing
func isThisTheFinishLine2(currentLat float64, currentLon float64, previousLat float64, previousLon float64) bool {
	movementVecLat := currentLat - previousLat
	movementVecLon := currentLon - previousLon
	lineVecLat := currentTrack.LatMax - currentTrack.LatMin
	lineVecLon := currentTrack.LonMax - currentTrack.LonMin

	crossProduct := lineVecLat*movementVecLon - lineVecLon*movementVecLat

	if crossProduct > 0 {
		fmt.Println("Car is moving forward across the line.")
		return true
	} else if crossProduct < 0 {
		fmt.Println("Car is moving backward across the line.")
		return false
	} else {
		fmt.Println("Car is moving along the line (no crossing).")
		return false
	}
}

func (wsConn *MySocket) HandleGpsLapTiming() {
	var gps *gpsd.Session
	var err error

	// Connect to the GPSD server
	for {
		gps, err = gpsd.Dial("localhost:2947")
		if err != nil {
			fmt.Println("Failed to connect to GPSD: ", err)
			fmt.Println("Retrying in 10 seconds...")
			time.Sleep(10 * time.Second)
			continue
		}

		fmt.Println("Connected to GPSD")
		break
	}
	defer gps.Close()
	
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
		currentLapData.CurrentLapTime = uint32(timeDiff.Milliseconds())

		// TODO: Testing
		
		// fmt.Println(report.Lat, "|", report.Lon)
		//fmt.Println(isThisTheFinishLine(currentTrack.LatMin, currentTrack.LatMax, report.Lat) && isThisTheFinishLine(currentTrack.LonMin, currentTrack.LonMax, report.Lon))

		//if isThisTheFinishLine(currentTrack.LatMin, currentTrack.LatMax, report.Lat) && isThisTheFinishLine(currentTrack.LonMin, currentTrack.LonMax, report.Lon) {
		if isThisTheFinishLine2(report.Lat, report.Lon, currentLapData.PreviousLat, currentLapData.PreviousLon) {
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
			lapStats.LapCount++

			// Send up to client
			jsonData, err := json.Marshal(lapStats)
			if err != nil {
				log.Fatal("Json Marshall error (Lap Stats)")
			}

      wsConn.writeToClient(lapStats.Type, jsonData)
		}

		// Testing
		currentLapData.PreviousLat = report.Lat
		currentLapData.PreviousLon = report.Lon

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
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"wt-race-dash/internal/tracks"

	"github.com/gorilla/websocket"
)

type AppSettings struct {
	CanChannel		string	`json:"canChannel"`
	Track					string	`json:"track"`
  LapTiming			bool		`json:"lapTiming"`
	LoggingHertz	int			`json:"loggingHertz"`
	CarOrEcu			string	`json:"carOrEcu"`
	WarningAlerts	bool		`json:"warningAlerts"`
	WarningValues	map[string]int
}

type MySocket struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

var (
	appSettings *AppSettings
	addr        = flag.String("addr", ":8080", "http service address")
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// Testing
	stopChanDatalogging = make(chan struct{}) // Channel to signal the goroutine to stop
	wg sync.WaitGroup // WaitGroup to ensure clean shutdown of goroutine
)

// **********************************************************************************************************

func (wsConn *MySocket) writeToClient(writeType int8, data []byte) {
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()

  err := wsConn.conn.WriteMessage(websocket.TextMessage, data)
  if err != nil {
    fmt.Println("Write error: ", writeType, err)
    return
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
		wsConn.HandleCanBusData()
	}()

	// ---------- Lap Timing ----------
  if (appSettings.LapTiming) {
    wg.Add(1)
    go func() {
      defer wg.Done()
      wsConn.HandleGpsLapTiming()
    }()
  }

	// ---------- Warning Alerts ----------
	// TODO: Requires further performance testing
	if (appSettings.WarningAlerts) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wsConn.HandleWarningAlerts()
		}()
	}

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

package main

import (
    "fmt"
    "log"
		"flag"
    "net/http"
    
		"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func handleWs(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Error upgrading WebSocket: ", err)
        return
    }
    defer conn.Close()

    for {
        // Reading message
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println("Read error: ", err)
            return
        }
        fmt.Println("Received: ", string(p))

        // Writing message back
        //if err := conn.WriteMessage(websocket.TextMessage, p); err != nil {
				if err := conn.WriteMessage(messageType, p); err != nil {
            log.Println("Write error:", err)
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

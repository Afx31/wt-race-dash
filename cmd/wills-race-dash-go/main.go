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
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func handleBaseHttp(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		fmt.Println("test1")
		messageType, p, err := conn.ReadMessage()
		fmt.Println("test2")
		if err != nil {
			fmt.Println("test3")
			log.Println(err)
			return
		}
		fmt.Println("----")
		fmt.Println(messageType)
		fmt.Println(p)
		fmt.Println("----")
		if err := conn.WriteMessage(messageType, p); err != nil {
			fmt.Println("test4")
			log.Println(err)
			return
		}
		fmt.Println("test5")
	}
}

func main() {
	fmt.Println("--- Running ---")

	flag.Parse()
	// Setup the base HTTP connection first
	http.HandleFunc("/", handleBaseHttp)
	// Then upgrade to the WebSocket connection
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWs(w, r)
	})
	
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}
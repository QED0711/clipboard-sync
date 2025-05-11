package main

import (
	"clipboard-sync/utils"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	mode              = flag.String("mode", "server", "Launch in `server` or `client` mode")
	serverPort        = flag.Int64("port", 8000, "The port to expose the server on (only needed when mode = `server`)")
	serverUrl         = flag.String("url", "ws://localhost:8000", "The url of the server (only needed when mode = `client`)")
	clipboardInterval = flag.Int("interval", 1000, "The polling interval in MS that checks for local clipboard updates")
	debug             = flag.Bool("debug", false, "Show debugging messages")
)

func main() {
	flag.Parse()

	if *mode == "server" {
		log.Printf("Hosting server on ws://0.0.0.0:%d/ws", *serverPort)
		http.HandleFunc("/ws", utils.WebsocketReceiverHandler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *serverPort), nil))
	} else if *mode == "client" {
		log.Printf("Attempting to connect to  %s", *serverUrl)
		localClient := utils.CreateLocalClient(*serverUrl, time.Duration(*clipboardInterval), *debug)
		defer localClient.Conn.Close()
		localClient.HandleMessage()
	}
}

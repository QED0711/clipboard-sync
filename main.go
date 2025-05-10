package main

import (
	"clipboard-sync/utils"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	isServer   = flag.Bool("server", false, "Whether or not to launch in server mode")
	serverPort = flag.Int64("port", 8000, "The port to expose the server on")
)

func main() {
	flag.Parse()

	fmt.Println("Is Server: ", *isServer)
	fmt.Println("Server Port: ", *serverPort)

	if *isServer {
		http.HandleFunc("/ws", utils.WebsocketReceiverHandler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *serverPort), nil))
	}
}

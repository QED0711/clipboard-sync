package utils

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gorilla/websocket"
)

type LocalClient struct {
	mu        sync.Mutex
	Conn      *websocket.Conn
	Clipboard string
	debug     bool
}

func CreateLocalClient(url string, interval time.Duration, debug bool) *LocalClient {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Dial to %s failed: %v\n", url, err)
	}

	var currentClip string
	currentClip, err = clipboard.ReadAll()
	if err != nil {
		currentClip = ""
	}

	client := &LocalClient{
		mu:        sync.Mutex{},
		Conn:      conn,
		Clipboard: currentClip,
		debug:     debug,
	}

	client.CheckClipboard(interval)

	return client
}

func (lc *LocalClient) CheckClipboard(interval time.Duration) {
	if interval == 0 {
		interval = 1000
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {

			clip, err := clipboard.ReadAll()
			if err != nil {
				log.Println("Failed to read clipboard: ", err)
				time.Sleep(interval * time.Millisecond)
				continue
			}

			lc.mu.Lock()
			if clip != lc.Clipboard {
				// fmt.Println("Clipboard changed: ", clip)
				lc.Clipboard = clip

				// Format server request
				postReq := Request{
					Type: MessageType.POST,
					Message: RegistryMessage{
						Name: "default",
						Data: clip,
					},
				}
				jsonReq, err := json.Marshal(postReq)
				if err != nil {
					log.Println("Failed to format POST request", err)
					lc.mu.Unlock()
					time.Sleep(interval * time.Millisecond)
					continue
				}

				// Send message to server with new clipboard contents
				lc.Conn.WriteMessage(websocket.TextMessage, jsonReq)
			}

			lc.mu.Unlock()

			time.Sleep(interval * time.Millisecond)
		}
	}()
}

func (lc *LocalClient) HandleMessage() {
	lc.mu.Lock()
	getReq, _ := json.Marshal(Request{
		Type:    MessageType.GET,
		Message: RegistryMessage{Name: "default"},
	})
	lc.Conn.WriteMessage(websocket.TextMessage, getReq)
	lc.mu.Unlock()

	for {
		_, message, err := lc.Conn.ReadMessage()
		if err != nil {
			log.Println("Read Error: ", err)
			return
		}
		err = clipboard.WriteAll(string(message))
		if err != nil {
			log.Println("Failed to write to clipboard: ", err)
			continue
		}
		if lc.debug {
			log.Printf("Updated clipboard to: %s\n", message)
		}
	}
}

package utils

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gorilla/websocket"
)

var (
	lastModTime time.Time
	clip        string
)

type LocalClient struct {
	mu                sync.Mutex
	Conn              *websocket.Conn
	Clipboard         string
	ClipboardFilePath string
	debug             bool
}

func CreateLocalClient(url string, interval time.Duration, clipboardFile string, debug bool) *LocalClient {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Dial to %s failed: %v\n", url, err)
	}

	currentClip := ""

	client := &LocalClient{
		mu:                sync.Mutex{},
		Conn:              conn,
		Clipboard:         currentClip,
		ClipboardFilePath: clipboardFile,
		debug:             debug,
	}

	if clipboardFile != "" {
		fileContents, _, _ := client.CheckClipboardFile()
		currentClip = fileContents
	} else {
		currentClip, err = clipboard.ReadAll()
		if err != nil {
			currentClip = ""
		}
	}

	client.Clipboard = currentClip
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
			clipboardChanged := false
			// Determine which method is to be used to read clipboard
			if lc.ClipboardFilePath != "" {
				clip, changed, err := lc.CheckClipboardFile()
				if err != nil {
					log.Println("Failed to read clipboard file: ", err)
					time.Sleep(interval * time.Millisecond)
					continue
				}
				if changed && clip != lc.Clipboard {
					lc.Clipboard = clip
					clipboardChanged = true
				}
			} else {
				clip, err := clipboard.ReadAll()
				if err != nil {
					log.Println("Failed to read clipboard: ", err)
					time.Sleep(interval * time.Millisecond)
					continue
				}
				if clip != lc.Clipboard {
					lc.Clipboard = clip
					clipboardChanged = true
				}

			}

			lc.mu.Lock()
			if clipboardChanged {
				// Format server request
				postReq := Request{
					Type: MessageType.POST,
					Message: RegistryMessage{
						Name: "default",
						Data: lc.Clipboard,
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

func (lc *LocalClient) CheckClipboardFile() (string, bool, error) {
	if lc.ClipboardFilePath == "" {
		return "", false, nil
	}
	fi, err := os.Stat(lc.ClipboardFilePath)
	if err != nil {
		return "", false, err
	}
	if fi.ModTime().After(lastModTime) {
		lastModTime = fi.ModTime()
		data, err := os.ReadFile(lc.ClipboardFilePath)
		if err != nil {
			return "", false, err
		}
		return string(data), true, nil
	}
	return "", false, nil
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

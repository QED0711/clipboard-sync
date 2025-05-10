package utils

import (
	"strings"
)

type RegistryMessage struct {
	Name string `json:"name"`
	Data string `json:"data",omitempty`
}

type Request struct {
	Type    string          `json:"type"`
	Message RegistryMessage `json:"message"`
}

func (req *Request) HandleRequest(client *Client, hub *WebSocketHub) {
	regCol := GetRegistryCollection()
	req.Type = strings.ToUpper(req.Type)

	if req.Type == MessageType.GET {
		cr, ok := regCol.GetRegistryByName(req.Message.Name)
		if ok {
			hub.RespondTo(client.ID, []byte(cr.Data))
		}
	} else if req.Type == MessageType.POST {
		cr, ok := regCol.SetRegistryData(req.Message.Name, req.Message.Data)
		if ok {
			hub.BroadcastFrom(client.ID, []byte(cr.Data))
		}
	}

}

// RequestTypes
type messageType struct {
	GET  string
	POST string
}

var MessageType = messageType{
	GET:  "GET",
	POST: "POST",
}

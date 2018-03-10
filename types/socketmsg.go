package types

////////////////////////////////////////////////////////////////////////////////

import "encoding/json"

////////////////////////////////////////////////////////////////////////////////

type SocketMessage struct {
	Type string      `json:"Type"`
	Data interface{} `json:"Data"`
}

func NewSocketMessage(t string, d interface{}) *SocketMessage {
	return &SocketMessage{
		Type: t,
		Data: d,
	}
}

func (sm *SocketMessage) Marshal() ([]byte, error) {
	return json.Marshal(sm)
}

func (sm *SocketMessage) Unmarshal(bs []byte) error {
	return json.Unmarshal(bs, sm)
}

////////////////////////////////////////////////////////////////////////////////

type JSONBroadcaster interface {
	BroadcastJSON(string, interface{}) error
}

////////////////////////////////////////////////////////////////////////////////

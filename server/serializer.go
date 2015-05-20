package server

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type Serializer interface {
	Serialize(Message) ([]byte, error)
	Deserialize([]byte) (Message, error)
}

type nopSerializer struct{}

func (s *nopSerializer) Serialize(m Message) []byte {
	return nil
}

func (s *nopSerializer) Deserialize(m []byte) Message {
	return nil
}

type JsonSerializer struct{}

func (s *JsonSerializer) Serialize(m Message) ([]byte, error) {
	data := map[string]interface{}{"type": m.MessageType(), "msg": m}

	return json.Marshal(&data)
}

func (s *JsonSerializer) Deserialize(m []byte) (Message, error) {
	payload := map[string]interface{}{}
	err := json.Unmarshal(m, &payload)

	if _, ok := payload["type"].(float64); !ok {
		return nil, fmt.Errorf("Type is not float64", payload["type"])
	}

	mt := messageType(int(payload["type"].(float64)))
	msg := mt.New()
	err = mapstructure.Decode(payload["msg"], msg)

	return msg, err
}

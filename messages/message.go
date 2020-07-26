package messages

import (
	"encoding/json"
	"errors"
)

const (
	// DELIM ...
	DELIM = 0x04
)

// MessageType ...
type MessageType int

const (
	TYPE_TEXT MessageType = 0x01
)

func (t MessageType) String() string {
	switch t {
	case TYPE_TEXT:
		return "TEXT"
	}
	return ""
}

// Message ...
type Message interface {
	Type() MessageType
}

type wrapper struct {
	Type    MessageType     `json:"type"`
	Content json.RawMessage `json:"content"`
}

// Decode ...
func Decode(bytes []byte) (Message, error) {
	bytes = bytes[:len(bytes)-1]

	var wrapper wrapper
	err := json.Unmarshal(bytes, &wrapper)
	if err != nil {
		return nil, err
	}

	var message Message
	switch wrapper.Type {
	case TYPE_TEXT:
		message = &Text{}
	default:
		return nil, errors.New("Unexpected message type " + string(wrapper.Type))
	}

	err = json.Unmarshal(wrapper.Content, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// Encode ...
func Encode(message Message) ([]byte, error) {
	bytesMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	bytesWrapper, err := json.Marshal(wrapper{
		Type:    message.Type(),
		Content: bytesMessage,
	})
	if err != nil {
		return nil, err
	}

	bytesWrapper = append(bytesWrapper, DELIM)

	return bytesWrapper, nil
}

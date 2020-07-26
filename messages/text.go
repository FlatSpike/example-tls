package messages

import (
	"fmt"
	"time"
)

// Text ...
type Text struct {
	Text string    `json:"text"`
	Name string    `json:"name"`
	Time time.Time `json:"time"`
}

// Type ...
func (text Text) Type() MessageType {
	return TYPE_TEXT
}

func (text Text) String() string {
	return fmt.Sprintf(
		"Type %s | Time %s | From %s | %s",
		text.Type().String(),
		text.Time.Format("2006.01.02 15.04.05"),
		text.Name,
		text.Text,
	)
}

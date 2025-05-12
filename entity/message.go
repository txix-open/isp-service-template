package entity

import (
	"database/sql/driver"

	"github.com/txix-open/isp-kit/json"
)

type Message struct {
	Id      int64
	Version int64
	Data    MessageData
}

type MessageData struct {
	Text string
}

// nolint
func (m *MessageData) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), m)
}

func (m *MessageData) Value() (driver.Value, error) {
	bytes, err := json.Marshal(m)
	return driver.Value(bytes), err
}

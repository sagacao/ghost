package utils

import "time"

type LogMsg struct {
	Database string
	Msg      map[string]interface{}
	When     time.Time
}

type RawMsg struct {
	Database string
	Msg      []byte
	When     time.Time
}

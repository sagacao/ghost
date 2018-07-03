package models

type Engine interface {
	Init() error

	Write(database string, json map[string]interface{})
	WriteRaw(database string, json map[string]interface{})
	Destroy()
	Flush()
}

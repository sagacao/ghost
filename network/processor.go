package network

type Processor interface {
	// must goroutine safe
	Route(cmd uint16, msg interface{}, agent Agent) error
	// must goroutine safe
	Unmarshal(data []byte) (uint16, interface{}, error)
	// must goroutine safe
	Marshal(cmd uint16, msg interface{}) ([][]byte, error)
}

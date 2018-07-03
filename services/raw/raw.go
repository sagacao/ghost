package raw

import (
	"encoding/binary"
	"errors"
	"fmt"

	. "ghost/global"
	"ghost/network"
)

// --------------------------
// | cmd | raw message |
// --------------------------
type MsgHandler struct {
	cmd     uint16
	handler func(network.Agent, interface{})
}

type RawProcessor struct {
	littleEndian bool
	handlers     map[uint16]*MsgHandler
}

func NewProcessor() *RawProcessor {
	p := new(RawProcessor)
	p.littleEndian = false
	p.handlers = make(map[uint16]*MsgHandler)
	return p
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *RawProcessor) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *RawProcessor) Register(cmd uint16, f func(network.Agent, interface{})) {
	if _, ok := p.handlers[cmd]; ok {
		Log.Critical("message %s is already registered", cmd)
	}

	p.handlers[cmd] = &MsgHandler{cmd: cmd, handler: f}
}

func (p *RawProcessor) Route(cmd uint16, msg interface{}, agent network.Agent) error {
	imsg := p.handlers[cmd]
	if imsg.handler != nil {
		imsg.handler(agent, msg)
	}
	return nil
}

func (p *RawProcessor) Unmarshal(data []byte) (uint16, interface{}, error) {
	if len(data) < 2 {
		return 0, nil, errors.New("raw data too short")
	}

	var cmd uint16
	if p.littleEndian {
		cmd = binary.LittleEndian.Uint16(data)
	} else {
		cmd = binary.BigEndian.Uint16(data)
	}
	//Console.Info("Unmarshal cmd:[%v]", cmd)
	imsg := p.handlers[cmd]
	if imsg == nil {
		return cmd, nil, fmt.Errorf("message %v not registered", cmd)
	}
	return cmd, data[2:], nil
}

func (p *RawProcessor) Marshal(cmd uint16, msg interface{}) ([][]byte, error) {
	_cmd := make([]byte, 2)
	if p.littleEndian {
		binary.LittleEndian.PutUint16(_cmd, cmd)
	} else {
		binary.BigEndian.PutUint16(_cmd, cmd)
	}

	var err error
	// data
	data, ok := msg.([]byte)
	if !ok {
		err = fmt.Errorf("message [%v] Marshal data error", cmd)
	}
	return [][]byte{_cmd, data}, err
}

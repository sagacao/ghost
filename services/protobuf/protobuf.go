package protobuf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"

	. "ghost/global"
	"ghost/network"

	"github.com/golang/protobuf/proto"
)

// --------------------------
// | cmd | protobuf message |
// --------------------------
type MsgHandler struct {
	cmd     uint16
	proto   reflect.Type
	handler func(network.Agent, interface{})
}

type ProtobufProcessor struct {
	littleEndian bool
	handlers     map[uint16]*MsgHandler
	ids          map[reflect.Type]uint16
}

func NewProcessor() *ProtobufProcessor {
	p := new(ProtobufProcessor)
	p.littleEndian = false
	p.ids = make(map[reflect.Type]uint16)
	p.handlers = make(map[uint16]*MsgHandler)
	return p
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *ProtobufProcessor) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *ProtobufProcessor) Register(cmd uint16, msg proto.Message, f func(network.Agent, interface{})) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		Log.Critical("protobuf message pointer required")
	}
	if _, ok := p.ids[msgType]; ok {
		Log.Critical("message %s is already registered", msgType)
	}
	if len(p.handlers) >= math.MaxUint16 {
		Log.Critical("too many protobuf messages (max = %v)", math.MaxUint16)
	}

	p.handlers[cmd] = &MsgHandler{cmd: cmd, proto: msgType, handler: f}
	p.ids[msgType] = cmd
}

func (p *ProtobufProcessor) Route(cmd uint16, msg interface{}, agent network.Agent) error {
	// protobuf
	msgType := reflect.TypeOf(msg)
	cmd, ok := p.ids[msgType]
	if !ok {
		return fmt.Errorf("message %s not registered", msgType)
	}
	imsg := p.handlers[cmd]
	if imsg.handler != nil {
		imsg.handler(agent, msg)
	}
	return nil
}

func (p *ProtobufProcessor) Unmarshal(data []byte) (uint16, interface{}, error) {
	if len(data) < 2 {
		return 0, nil, errors.New("protobuf data too short")
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
	msg := reflect.New(imsg.proto.Elem()).Interface()
	return cmd, msg, proto.UnmarshalMerge(data[2:], msg.(proto.Message))
}

func (p *ProtobufProcessor) Marshal(cmd uint16, msg interface{}) ([][]byte, error) {
	msgType := reflect.TypeOf(msg)
	_cmd, ok := p.ids[msgType]
	if !ok {
		err := fmt.Errorf("message %s not registered", msgType)
		return nil, err
	}

	cmd := make([]byte, 2)
	if p.littleEndian {
		binary.LittleEndian.PutUint16(cmd, _cmd)
	} else {
		binary.BigEndian.PutUint16(cmd, _cmd)
	}
	// data
	data, err := proto.Marshal(msg.(proto.Message))
	return [][]byte{cmd, data}, err
}

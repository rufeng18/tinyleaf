package json2

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"reflect"

	"github.com/rufeng18/tinyleaf/chanrpc"
	"github.com/rufeng18/tinyleaf/log"
)

type Processor struct {
	msgInfo map[string]*MsgInfo
	msgIdTypes map[reflect.Type]string // type 与 msgId对应, 使用此处理器,  协议里cmd和消息结构名称必须是独一且对应 不能不同的cmd对应同一个消息结构
}

type MsgInfo struct {
	msgType       reflect.Type
	msgRouter     *chanrpc.Server
	msgHandler    MsgHandler
}

type MsgHandler func([]interface{})

func NewProcessor() *Processor {
	p := new(Processor)
	p.msgInfo = make(map[string]*MsgInfo)
	p.msgIdTypes = make(map[reflect.Type]string)
	return p
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) Register(msgID string, msg interface{}) string {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatal("json message pointer required")
	}
	if _, ok := p.msgInfo[msgID]; ok {
		log.Fatal("message %v is already registered", msgID)
	}

	i := new(MsgInfo)
	i.msgType = msgType
	p.msgInfo[msgID] = i
	p.msgIdTypes[msgType] = msgID
	return msgID
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) SetRouter(msg interface{}, msgRouter *chanrpc.Server) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatal("json message pointer required")
	}
	msgID, ok := p.msgIdTypes[msgType]
	if !ok {
		log.Fatal("message %v not registered", msgType)
	}
	i, ok := p.msgInfo[msgID]
	if !ok {
		log.Fatal("message %v not registered", msgID)
	}

	i.msgRouter = msgRouter
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) SetHandler(msg interface{}, msgHandler MsgHandler) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatal("json message pointer required")
	}
	msgID, ok := p.msgIdTypes[msgType]
	if !ok {
		log.Fatal("message %v not registered", msgType)
	}
	i, ok := p.msgInfo[msgID]
	if !ok {
		log.Fatal("message %v not registered", msgID)
	}

	i.msgHandler = msgHandler
}

// goroutine safe
func (p *Processor) Route(msg interface{}, userData interface{}) error {

	// json
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		return errors.New("json message pointer required")
	}
	msgID, ok1 := p.msgIdTypes[msgType]
	if !ok1 {
		log.Fatal("message %v not registered", msgType)
	}

	i, ok := p.msgInfo[msgID]
	if !ok {
		return fmt.Errorf("message %v not registered", msgID)
	}
	if i.msgHandler != nil {
		i.msgHandler([]interface{}{msg, userData})
	}
	if i.msgRouter != nil {
		i.msgRouter.Go(msgType, msg, userData)
	}
	return nil
}

// goroutine safe
func (p *Processor) Unmarshal(data []byte) (interface{}, error) {
	msgID := jsoniter.Get(data, "cmd").ToString()
	i, ok := p.msgInfo[msgID]
	if !ok {
		return nil, fmt.Errorf("message %v not registered", msgID)
	}
	msg := reflect.New(i.msgType.Elem()).Interface()
	return msg, jsoniter.Unmarshal(data, msg)
}

// goroutine safe
func (p *Processor) Marshal(msg interface{}) ([][]byte, error) {
	//msgType := reflect.TypeOf(msg)
	//if msgType == nil || msgType.Kind() != reflect.Ptr {
	//	return nil, errors.New("json message pointer required")
	//}
	//msgID, ok1 := p.msgIdTypes[msgType]
	//if !ok1 {
	//	return nil, fmt.Errorf("message %v not registered", msgType)
	//}
	//
	//if _, ok := p.msgInfo[msgID]; !ok {
	//	return nil, fmt.Errorf("message %v not registered", msgID)
	//}

	// data
	//var json = jsoniter.ConfigCompatibleWithStandardLibrary
	data, err := jsoniter.Marshal(msg)
	return [][]byte{data}, err
}

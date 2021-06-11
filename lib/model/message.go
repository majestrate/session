package model

import (
	"errors"
	"fmt"
	"github.com/majestrate/session/lib/protobuf"
	"google.golang.org/protobuf/proto"
	"strings"
	"unicode"
)

type Message struct {
	Raw       string
	Hash      string
	Timestamp string
}

func (msg *Message) IRCLine() string {
	return strings.TrimFunc(msg.Raw, func(r rune) bool {
		if r == '\x01' || unicode.IsPrint(r) {
			return false
		}
		return r == '\r' || r == '\n'
	})
}

func (msg *Message) From() string {
	return "anonymous"
}

func (msg *Message) Data() []byte {
	return []byte(msg.Raw)
}

func (msg *Message) decodeEnvelope() (*protobuf.Envelope, error) {
	env := &protobuf.Envelope{}
	err := proto.Unmarshal([]byte(msg.Raw), env)
	if err != nil {
		return nil, err
	}
	return env, err
}

func (msg *Message) Decode() ([]byte, error) {
	env, err := msg.decodeEnvelope()
	if err != nil {
		return nil, err
	}
	if env.Source == nil {
		return nil, errors.New("no source in envelope")
	}
	req := &protobuf.WebSocketRequestMessage{}
	err = proto.Unmarshal([]byte(*env.Source), req)
	if err != nil {
		fmt.Printf("bad websocket message: %s\n", err.Error())
		return nil, err
	}
	if req == nil {
		return nil, errors.New("no request in envelope")
	}
	m := &Message{Raw: string(req.Body)}
	env, err = m.decodeEnvelope()
	if err != nil {
		fmt.Printf("failed to decode inner envelope: %s\n", err.Error())
		return nil, err
	}
	return env.Content, nil
}

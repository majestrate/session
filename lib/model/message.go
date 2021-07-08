package model

import (
	"errors"
	"fmt"
	"github.com/majestrate/session/lib/protobuf"
	"github.com/majestrate/session/lib/cryptography"
	"google.golang.org/protobuf/proto"
	"strings"
	"unicode"
	"encoding/hex"
	"bytes"
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

func (msg *Message) decodeRaw() ([]byte, error) {
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

type PlainMessage struct {
	Message *protobuf.DataMessage
	From string
}

func (msg *Message) Decrypt(keys *cryptography.KeyPair) (*PlainMessage, error) {
	raw, err :=  msg.decodeRaw()
	if err != nil {
		return nil, err
	}
	data, from, err := keys.DecryptAndVerify(raw)
	if err != nil {
		return nil, err
	}
	// kill padding
	idx := bytes.LastIndexByte(data, 0x80)
	data = data[:idx]
	plain := new(PlainMessage)
	var content protobuf.Content
	plain.From = fmt.Sprintf("05%s", hex.EncodeToString(from))
	err = proto.Unmarshal(data, &content)
	if err != nil {
		return nil, err
	}
	plain.Message = content.GetDataMessage()
	return plain, nil
}

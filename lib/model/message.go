package model

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/majestrate/session/lib/cryptography"
	"github.com/majestrate/session/lib/protobuf"
	"google.golang.org/protobuf/proto"
	"strings"
	"time"
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
		return nil, err
	}
	if req == nil {
		return nil, errors.New("no request in envelope")
	}
	m := &Message{Raw: string(req.Body)}
	env, err = m.decodeEnvelope()
	if err != nil {
		return nil, err
	}
	return env.Content, nil
}

type PlainMessage struct {
	Message *protobuf.DataMessage
	From    string
}

func (plain *PlainMessage) Body() string {
	if plain.Message == nil || plain.Message.Body == nil {
		return ""
	}
	return *plain.Message.Body
}

func (plain *PlainMessage) When() time.Time {
	t := int64(0)
	if plain.Message != nil && plain.Message.Timestamp != nil {
		t = int64(*plain.Message.Timestamp) / 1000
	}
	return time.Unix(t, 0)
}

func (plain *PlainMessage) ReplyTag() []byte {
	return nil
}

var wsVerb = "PUT"
var wsPath = "/api/v1/message"
var wsID = uint64(0)
var envSource = ""

var envType = protobuf.Envelope_UNIDENTIFIED_SENDER.Enum()
var outerEnvType = protobuf.Envelope_Type(1)

func (msg *PlainMessage) Encrypt(keys *cryptography.KeyPair, to string) ([]byte, error) {
	now := uint64(time.Now().Unix() * 1000)
	msg.Message.Timestamp = &now
	content := protobuf.Content{
		DataMessage: msg.Message,
	}
	data, err := proto.Marshal(&content)
	if err != nil {
		return nil, err
	}
	to_bytes, err := hex.DecodeString(to[2:])
	if err != nil {
		return nil, err
	}
	raw, err := keys.SignAndEncrypt(to_bytes, data)
	if err != nil {
		return nil, err
	}
	innerEnv := &protobuf.Envelope{
		Type:      envType,
		Content:   raw,
		Timestamp: &now,
	}

	envRaw, err := proto.Marshal(innerEnv)

	if err != nil {
		return nil, err
	}

	req := &protobuf.WebSocketRequestMessage{
		Body: envRaw,
		Verb: &wsVerb,
		Path: &wsPath,
		Id:   &wsID,
	}

	reqData, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqDataStr := string(reqData)
	env := &protobuf.Envelope{
		Type:      &outerEnvType, // because shit af protocol
		Source:    &reqDataStr,
		Timestamp: &now,
	}
	return proto.Marshal(env)
}

func MakePlain(data string) *PlainMessage {
	return &PlainMessage{
		Message: &protobuf.DataMessage{
			Body: &data,
		},
	}
}

func (msg *Message) Decrypt(keys *cryptography.KeyPair) (*PlainMessage, error) {
	raw, err := msg.decodeRaw()
	if err != nil {
		return nil, fmt.Errorf("decode outer envelope failed: %s", err.Error())
	}
	data, from, err := keys.DecryptAndVerify(raw)
	if err != nil {
		return nil, fmt.Errorf("decrypt and verify failed: %s", err.Error())
	}
	plain := new(PlainMessage)
	content := &protobuf.Content{}
	plain.From = fmt.Sprintf("05%s", hex.EncodeToString(from))
	err = proto.Unmarshal(data, content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode inner content: %s", err.Error())
	}
	plain.Message = content.GetDataMessage()
	return plain, nil
}

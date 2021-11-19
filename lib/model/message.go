package model

import (
	"encoding/hex"

	"fmt"
	"github.com/majestrate/ubw/lib/cryptography"
	"github.com/majestrate/ubw/lib/protobuf"
	"google.golang.org/protobuf/proto"
	"time"
)

type Message struct {
	Raw       []byte
	Hash      string
	Timestamp string
}

func (msg *Message) decodeRaw() ([]byte, error) {
	req := &protobuf.WebSocketRequestMessage{}
	err := proto.Unmarshal([]byte(msg.Raw), req)
	if err != nil {
		return nil, err
	}
	env := &protobuf.Envelope{}
	err = proto.Unmarshal(req.Body, env)
	if err != nil {
		return nil, err
	}
	return env.Content, nil
}

type PlainMessage struct {
	Message *protobuf.DataMessage
	From    string
}

func (plain *PlainMessage) Body() *string {
	if plain.Message == nil || plain.Message.Body == nil {
		return nil
	}
	return plain.Message.Body
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

var innerEnvType = protobuf.Envelope_UNIDENTIFIED_SENDER.Enum()

func (msg *PlainMessage) Encrypt(keys *cryptography.KeyPair, to string) ([]byte, error) {
	now := uint64(time.Now().UnixNano() / 1000000)
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
		Type:      innerEnvType,
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

	return proto.Marshal(req)
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

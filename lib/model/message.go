package model

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/majestrate/session/lib/cryptography"
	"github.com/majestrate/session/lib/protobuf"
	"google.golang.org/protobuf/proto"
	"math"
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
	fmt.Printf("recv %q\n", env)
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
	From    string
}

func (plain *PlainMessage) ReplyTag() []byte {
	if plain.Message == nil {
		return nil
	}
	return plain.Message.ProfileKey
}

const partSize = 160
const padDelim = 0x80
const padByte = 0x00

func getPaddedMessageLength(originalLen int) int {
	originalLen += 1
	numParts := int(math.Floor(float64(originalLen) / partSize))
	if numParts%partSize != 0 {
		numParts += 1
	}
	return numParts * partSize
}

func addPadding(data *[]byte) {
	dlen := len(*data)
	msglen := getPaddedMessageLength(len(*data)+1) - 1
	padlen := msglen - dlen
	*data = append(*data, padDelim)
	for padlen > 0 {
		*data = append(*data, padByte)
		padlen--
	}
}

var wsVerb = "PUT"
var wsPath = "/api/v1/message"
var wsID = uint64(0)
var envSource = ""

var envType = protobuf.Envelope_UNIDENTIFIED_SENDER.Enum()

func (msg *PlainMessage) Encrypt(keys *cryptography.KeyPair) ([]byte, error) {
	now := uint64(time.Now().Unix() * 1000)

	content := protobuf.Content{
		DataMessage: msg.Message,
	}
	data, err := proto.Marshal(&content)
	if err != nil {
		return nil, err
	}
	addPadding(&data)
	from, err := hex.DecodeString(msg.From[2:])
	if err != nil {
		return nil,err
	}
	raw, err := keys.SignAndEncrypt(from, data)
	if err != nil {
		return nil, err
	}
	rawStr := string(raw)
	innerEnv := &protobuf.Envelope{
		Type:      envType,
		Source:    &rawStr,
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
		Type:      envType,
		Source:    &reqDataStr,
		Timestamp: &now,
	}
	fmt.Printf("%q\n", env)
	return proto.Marshal(env)
}

func MakePlain(from, data string, tag []byte) *PlainMessage {
	now := uint64(time.Now().Unix() * 1000)
	return &PlainMessage{
		From: from,
		Message: &protobuf.DataMessage{
			Body:       &data,
			Timestamp:  &now,
			ProfileKey: tag,
			SyncTarget: &from,
			Profile: &protobuf.DataMessage_LokiProfile{
				DisplayName: &from,
			},
		},
	}
}

func (msg *Message) Decrypt(keys *cryptography.KeyPair) (*PlainMessage, error) {
	raw, err := msg.decodeRaw()
	if err != nil {
		return nil, err
	}
	data, from, err := keys.DecryptAndVerify(raw)
	if err != nil {
		return nil, err
	}
	// kill padding
	idx := bytes.LastIndexByte(data, padDelim)
	data = data[:idx]
	plain := new(PlainMessage)
	var content protobuf.Content
	plain.From = fmt.Sprintf("05%s", hex.EncodeToString(from))
	err = proto.Unmarshal(data, &content)
	if err != nil {
		return nil, err
	}
	fmt.Printf("recv %s\n", content.String())
	plain.Message = content.GetDataMessage()
	return plain, nil
}

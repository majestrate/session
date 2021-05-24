package model

type Message struct {
	Raw string
	Hash string
	Timestamp string
}

func (msg *Message) Data() []byte {
	return []byte(msg.Raw)
}

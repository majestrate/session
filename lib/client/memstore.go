package client

import "github.com/majestrate/ubw/lib/model"

import "strconv"

type memStore struct {
	lastTimestamp int64
	lastHash      string
	msgs          map[string]model.Message
}

func (m *memStore) HasMessage(hash string) bool {
	_, ok := m.msgs[hash]
	return ok
}

func (m *memStore) Put(msg model.Message) error {
	t, _ := strconv.ParseInt(msg.Timestamp, 10, 64)
	m.msgs[msg.Hash] = msg
	if m.lastTimestamp < t {
		m.lastHash = msg.Hash
		m.lastTimestamp = t
	}
	return nil
}

func (m *memStore) LastHash() string {
	return m.lastHash
}

func (m *memStore) Close() error {
	m.lastHash = ""
	m.lastTimestamp = 0
	m.msgs = make(map[string]model.Message)
	return nil
}

func MemoryStore() MessageStore {
	return &memStore{msgs: make(map[string]model.Message)}
}

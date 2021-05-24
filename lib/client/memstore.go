package client

import "github.com/majestrate/session2/lib/model"

type memStore struct {
	lastTimestamp string
	lastHash string
	msgs map[string]model.Message
}

func (m *memStore) HasMessage(hash string) bool {
	_, ok := m.msgs[hash]
	return ok
}

func (m *memStore) Put(msg model.Message) error {
	m.msgs[msg.Hash] = msg
	if m.lastTimestamp > msg.Timestamp {
		m.lastHash =  msg.Hash
	}
	return nil
}

func (m *memStore) LastHash() string {
	return m.lastHash
}


func MemoryStore() MessageStore {
	return &memStore{msgs:make(map[string]model.Message)}
}

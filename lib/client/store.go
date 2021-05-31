package client

import (
	"github.com/majestrate/session/lib/model"
)

type MessageStore interface {
	HasMessage(hash string) bool
	Put(msg model.Message) error
	LastHash() string
}

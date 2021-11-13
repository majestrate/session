package client

import (
	"github.com/majestrate/ubw/lib/model"
	"io"
)

type MessageStore interface {
	HasMessage(hash string) bool
	Put(msg model.Message) error
	LastHash() string
	io.Closer
}

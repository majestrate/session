package client

import (
	_ "errors"
	"fmt"
	"github.com/majestrate/session/lib/cryptography"
	"github.com/majestrate/session/lib/model"
	"github.com/majestrate/session/lib/swarm"
	"time"
)

type Client struct {
	keys     *cryptography.KeyPair
	snodes   SnodeMap
	store    MessageStore
	ourSwarm *swarm.ServiceNode
}

func (cl *Client) SessionID() string {
	return cl.keys.SessionID()
}

func NewClient(keys *cryptography.KeyPair) *Client {
	return &Client{
		keys: keys,
		snodes: SnodeMap{
			snodeMap:     make(map[string]swarm.ServiceNode),
			nextUpdateAt: time.Now(),
		},
		store: MemoryStore(),
	}
}

func (cl *Client) Update() {
	if cl.snodes.Empty() {
		swarm.WithSeedNodes(func(node swarm.ServiceNode) {
			err := cl.snodes.Update(node)
			if err != nil {
				fmt.Printf("Failed to fetch from seed node: %s\n", err.Error())
			}
		})
	}
}
func (cl *Client) withRandomSNode(visit func(swarm.ServiceNode)) {
	visit(cl.snodes.Random())
}

func (cl *Client) FetchNewMessages() ([]model.Message, error) {
	return cl.recvFrom(cl.SessionID())
}

func (cl *Client) RecvFromHash(src string) ([]model.Message, error) {
	src = "05" + cryptography.B2SumHex(src)
	fmt.Printf("recv from %s\n", src)
	return cl.recvFrom(src)
}

func (cl *Client) recvFrom(src string) (found []model.Message, err error) {
	node := cl.snodes.Random()
	msgs, err2 := node.FetchMessages(src, cl.store.LastHash())
	err = err2
	if err == nil {
		for _, msg := range msgs {
			if cl.store.HasMessage(msg.Hash) {
				continue
			}
			err = cl.store.Put(msg)
			if err != nil {
				return
			}
			found = append(found, msg)
		}
	}
	fmt.Printf("got %d new messages\n", len(found))
	return
}

func (cl *Client) DecryptMessage(msg model.Message) ([]byte, error) {
	data, err := msg.Decode()
	if err != nil {
		return nil, err
	}
	return cl.keys.DecryptSessionMessage(data)
	// return []byte(msg.Raw), nil
}

/// SendT sends a message msg to destination dest (some string)
func (cl *Client) SendToHash(dest, msg string) {
	dest = "05" + cryptography.B2SumHex(dest)
	fmt.Printf("send to %s\n", dest)
	node := cl.snodes.Random()
	node.StoreMessage(dest, model.Message{Raw: msg})
}

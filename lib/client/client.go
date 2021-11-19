package client

import (
	_ "errors"
	"fmt"
	"github.com/majestrate/ubw/lib/cryptography"
	"github.com/majestrate/ubw/lib/model"
	"github.com/majestrate/ubw/lib/swarm"
	"time"
)

type Client struct {
	keys     *cryptography.KeyPair
	snodes   SnodeMap
	store    MessageStore
	ourSwarm *swarm.ServiceNode
}

func (cl *Client) Store() MessageStore {
	return cl.store
}

func (cl *Client) SessionID() string {
	return cl.keys.SessionID()
}

func NewClient(keys *cryptography.KeyPair, store MessageStore) *Client {
	if store == nil {
		store = MemoryStore()
	}
	return &Client{
		keys: keys,
		snodes: SnodeMap{
			snodeMap:     make(map[string]swarm.ServiceNode),
			nextUpdateAt: time.Now(),
		},
		store: store,
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
	return
}

func (cl *Client) DecryptMessage(msg model.Message) (*model.PlainMessage, error) {
	return msg.Decrypt(cl.keys)
}

func (cl *Client) makePlain(data string) *model.PlainMessage {
	msg := model.MakePlain(data)
	return msg
}

func (cl *Client) SendTo(dst, body string) error {
	msg := cl.makePlain(body)
	raw, err := msg.Encrypt(cl.keys, dst)
	if err != nil {
		return err
	}
	cl.snodes.VisitSwarmFor(dst, 1, func(node swarm.ServiceNode) {
		node.StoreMessage(dst, model.Message{Raw: raw})
	})
	return nil
}

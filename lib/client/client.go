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
		cl.snodes.VisitSwarmFor(cl.SessionID(), func(node swarm.ServiceNode) {
			if cl.ourSwarm == nil {
				cl.ourSwarm, _ = node.StoreMessage(cl.SessionID(), model.Message{Raw: "bemis"})
			}
		})
	} else if cl.snodes.ShouldUpdate() {
		cl.snodes.VisitSwarmFor(cl.SessionID(), func(node swarm.ServiceNode) {
			err := cl.snodes.Update(node)
			if err != nil {
				fmt.Printf("Failed to fetch from %s: %s\n", node.SNodeAddr(), err.Error())
				return
			}
			if cl.ourSwarm == nil {
				cl.ourSwarm, err = node.StoreMessage(cl.SessionID(), model.Message{Raw: "bemis"})
			}
		})
	}
}
func (cl *Client) withRandomSNode(visit func(swarm.ServiceNode)) {
	visit(cl.snodes.Random())
}

func (cl *Client) FetchNewMessages() (found []model.Message, err error) {
	fmt.Printf("fetching new messages...\n")
	if cl.ourSwarm != nil {
		msgs, err2 := cl.ourSwarm.FetchMessages(cl.SessionID(), cl.store.LastHash())
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
	}
	return
}

func (cl *Client) DecryptMessage(msg model.Message) ([]byte, error) {
	return msg.Decode()
}

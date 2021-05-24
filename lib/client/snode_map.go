package client

import (
	"fmt"
	"time"
	"math/rand"
	"github.com/majestrate/session2/lib/constants"
	"github.com/majestrate/session2/lib/swarm"
)

type SnodeMap struct {
	snodeMap map[string]swarm.ServiceNode
	nextUpdateAt time.Time
}

func (s *SnodeMap) VisitSwarmFor(id string, visit func(swarm.ServiceNode)) {
	visit(s.Random())
}

func (s *SnodeMap) Random() (node swarm.ServiceNode) {
	if s.Empty() {
		return
	}
	idx := rand.Int() % len(s.snodeMap)
	for _, info := range s.snodeMap {
		if idx == 0 {
			node = info
		}
		idx --
	}
	return
}

func (s *SnodeMap) Empty() bool {
	return len(s.snodeMap) == 0
}

func (s *SnodeMap) ShouldUpdate() bool {
	return s.nextUpdateAt.After(time.Now())
}

func (s *SnodeMap) Update(node swarm.ServiceNode) error {
		peers, err := node.GetSNodeList()
		if err != nil {
			return err
		}
		fmt.Printf("got %d nodes from %s\n", len(peers), node.SNodeAddr())
		s.snodeMap = make(map[string]swarm.ServiceNode)
		for _, peer := range peers {
			s.snodeMap[peer.IdentityKey] = peer
		}
	s.nextUpdateAt = time.Now().Add(constants.SNodeMapUpdateInterval * time.Second)
	return nil
}

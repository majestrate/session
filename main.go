package main

import (
	"github.com/majestrate/session2/lib/config"
	"github.com/majestrate/session2/lib/cryptography"
	"github.com/majestrate/session2/lib/swarm"
	_ "github.com/majestrate/session2/lib/fetcher"
	"fmt"
	"os"
)

func main() {
	fmt.Println("session2 starting up")
	_, err := config.Load()
	if err != nil {
		fmt.Printf("error loading config: %s\n", err.Error())
		return
	}

	sessionID := ""
	
	if len(os.Args) <= 1 {
		keys := cryptography.Keygen()
		sessionID = keys.SessionID()
	} else {
		sessionID = os.Args[1]
	}

	fmt.Printf("we are %s\n", sessionID)
	
	snodeMap := make(map[string]swarm.ServiceNode)
	
	swarm.WithSeedNodes(func(node swarm.ServiceNode) {
		peers, err := node.GetSNodeList()
		if err != nil {
			fmt.Printf("error fetching node list from %s: %s\n", node.SNodeAddr(), err.Error())
			return
		}
		fmt.Printf("got %d nodes from %s\n", len(peers), node.SNodeAddr())
		for _, peer := range peers {
			snodeMap[peer.IdentityKey] = peer
		}
	})
	
	for pk, snode := range snodeMap {
		fmt.Printf("store at %s\n", pk)
		where, err := snode.StoreMessage(sessionID, "bemis")
		if err != nil {
			fmt.Printf("error storing: %s\n", err.Error())
			continue
		}
		msgs, err := where.FetchMessages(sessionID)
		if err != nil {
			fmt.Printf("error fetching: %s\n", err.Error())
			return
		}
		fmt.Printf("got %d messages\n", len(msgs))
		for idx, msg := range msgs {
			fmt.Printf("%d: %q\n", idx, msg)
		}
	}
}

package main

import (
	"github.com/majestrate/session2/lib/config"
	"github.com/majestrate/session2/lib/client"
	_ "github.com/majestrate/session2/lib/fetcher"
	"fmt"
	"time"
)

func main() {
	fmt.Println("session2 starting up")
	_, err := config.Load()
	if err != nil {
		fmt.Printf("error loading config: %s\n", err.Error())
		return
	}

	me := client.NewClient(nil)
	
	fmt.Printf("we are %s\n", me.SessionID())

	for {

		me.Update()

		msgs, err := me.FetchNewMessages()
		if err != nil {
			fmt.Printf("failed to get new messages: %s\n", err.Error())
		}
		for idx, msg := range msgs {
			fmt.Printf("%d: %q\n", idx, msg)
		}
		time.Sleep(5 * time.Second)
	}
}

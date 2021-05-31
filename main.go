package main

import (
	"fmt"
	"github.com/majestrate/session/lib/client"
	"github.com/majestrate/session/lib/config"
	"github.com/majestrate/session/lib/cryptography"
	_ "github.com/majestrate/session/lib/fetcher"
	"os"
	"time"
)

const keyfile = "seed.dat"

func main() {
	fmt.Println("session starting up")
	_, err := config.Load()
	if err != nil {
		fmt.Printf("error loading config: %s\n", err.Error())
		return
	}

	keys := new(cryptography.KeyPair)

	if _, err := os.Stat(keyfile); os.IsNotExist(err) {
		keys.Regen()
		keys.SaveFile(keyfile)
	}
	err = keys.LoadFile(keyfile)
	if err != nil {
		fmt.Printf("could not load %s: %s\n", keyfile, err.Error())
		return
	}

	me := client.NewClient(keys)

	fmt.Printf("we are %s\n", me.SessionID())

	for {

		me.Update()

		msgs, err := me.FetchNewMessages()
		if err != nil {
			fmt.Printf("failed to get new messages: %s\n", err.Error())
		}
		for idx, msg := range msgs {
			m, err := me.DecryptMessage(msg)
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
				continue
			}
			fmt.Printf("%d: %q\n", idx, m)
		}
		time.Sleep(5 * time.Second)
	}
}

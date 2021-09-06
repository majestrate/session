package main

import (
	"fmt"
	"github.com/majestrate/session/lib/client"
	"github.com/majestrate/session/lib/config"
	"github.com/majestrate/session/lib/cryptography"
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
	baseDelay := 5 * time.Second
	delay := baseDelay
	for {
		me.Update()
		msgs, err := me.FetchNewMessages()
		if err != nil {
			fmt.Printf("fetch failed: %s\n", err.Error())
			delay += 5 * time.Second
			time.Sleep(delay)
			continue
		}
		if len(msgs) > 0 {
			fmt.Printf("got %d new messages\n", len(msgs))
		}
		for _, msg := range msgs {
			plain, err := me.DecryptMessage(msg)
			if err != nil {
				fmt.Printf("decrypt failed: %s\n", err.Error())
				continue
			}
			fmt.Printf("%q\n", plain.Message)
			body := plain.Message.GetBody()
			fmt.Printf("<%s> %s\n", plain.From, body)
			err = me.SendTo(plain.From, "penis "+body, plain.ReplyTag())
			if err != nil {
				fmt.Printf("sendto failed: %s\n", err.Error())
			}
		}

		delay = baseDelay
		time.Sleep(delay)
	}
}

package main

import (
	"database/sql"
	"fmt"
	"github.com/majestrate/ubw/lib/client"
	"github.com/majestrate/ubw/lib/cryptography"
	"github.com/majestrate/ubw/lib/model"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"os/exec"
	"time"
)

const keyfile = "seed.dat"

func main() {
	fmt.Println("session starting up")

	keys := new(cryptography.KeyPair)

	if _, err := os.Stat(keyfile); os.IsNotExist(err) {
		keys.Regen()
		keys.SaveFile(keyfile)
	}
	err := keys.LoadFile(keyfile)
	if err != nil {
		fmt.Printf("could not load %s: %s\n", keyfile, err.Error())
		return
	}
	c, err := sql.Open("sqlite3", "messages.db")
	if err != nil {
		fmt.Printf("could not open database: %s", err.Error())
		return
	}
	store := client.SQLStore(c)
	defer store.Close()

	makeReply := func(msg *model.PlainMessage) *string {
		return msg.Body()
	}
	if len(os.Args) >= 2 {
		exe := os.Args[1]
		args := os.Args[2:]
		makeReply = func(msg *model.PlainMessage) *string {
			var ret string
			cmd := exec.Command(exe, args...)
			cmd.Env = append(os.Environ(), fmt.Sprintf("SESSION_ID=%s", msg.From), fmt.Sprintf("SESSION_MESSAGE=%s", msg.Body()))
			data, err := cmd.Output()
			if err == nil {
				ret = string(data)
			} else {
				ret = err.Error()
			}
			return &ret
		}
	}

	me := client.NewClient(keys, store)
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
			if plain.Body() == nil {
				continue
			}
			reply := makeReply(plain)
			if reply == nil {
				continue
			}
			err = me.SendTo(plain.From, *reply)
			if err != nil {
				fmt.Printf("sendto failed: %s\n", err.Error())
			}
		}
		delay = baseDelay
		time.Sleep(delay)
	}
}

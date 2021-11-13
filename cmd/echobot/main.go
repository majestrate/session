package main

import (
	"database/sql"
	"fmt"
	"github.com/majestrate/session/lib/client"
	"github.com/majestrate/session/lib/cryptography"
	"github.com/majestrate/session/lib/model"
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

	makeReply := func(msg *model.PlainMessage) string {
		return msg.Body()
	}
	if len(os.Args) >= 2 {
		exe := os.Args[1]
		args := os.Args[2:]
		makeReply = func(msg *model.PlainMessage) string {
			cmd := exec.Command(exe, args...)
			cmd.Env = append(os.Environ(), fmt.Sprintf("SESSION_ID=%s", msg.From), fmt.Sprintf("SESSION_MESSAGE=%s", msg.Body()))
			data, err := cmd.Output()
			if err != nil {
				return err.Error()
			}
			return string(data)
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
			body := plain.Body()
			if len(body) == 0 {
				continue
			}
			fmt.Printf("%s | <%s> %s\n", plain.When(), plain.From, body)
			err = me.SendTo(plain.From, makeReply(plain))
			if err != nil {
				fmt.Printf("sendto failed: %s\n", err.Error())
			}
		}
		delay = baseDelay
		time.Sleep(delay)
	}
}

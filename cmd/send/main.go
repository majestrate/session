package main

import (
	"fmt"
	"github.com/majestrate/session/lib/client"
	"github.com/majestrate/session/lib/config"
	"github.com/majestrate/session/lib/cryptography"
	"os"
)

const keyfile = "seed.dat"

func main() {
	if len(os.Args) <= 2 {
		fmt.Printf("usage: %s session_id message goes here\n", os.Args[0])
		return
	}
	fmt.Println("session starting up")
	_, err := config.Load()
	if err != nil {
		fmt.Printf("error loading config: %s\n", err.Error())
		return
	}

	keys := new(cryptography.KeyPair)
	keys.Regen()

	me := client.NewClient(keys)

	to := os.Args[1]
	var msg string
	for _, arg := range os.Args[2:] {
		msg += fmt.Sprintf("%s ", arg)
	}

	me.Update()
	me.SendTo(to, msg, keys.PubKey())

}

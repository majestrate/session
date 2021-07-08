package main

import (
	"fmt"
	"github.com/majestrate/session/lib/client"
	"github.com/majestrate/session/lib/config"
	"github.com/majestrate/session/lib/cryptography"
	"github.com/majestrate/session/lib/irc"
	"net"
	"os"
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
	port := "6667"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	addr := net.JoinHostPort("127.0.0.1", port)
	fmt.Printf("starting irc daemon at %s\n", addr)
	sock, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("could not set up irc: %s\n", err.Error())
		return
	}
	defer sock.Close()
	server := irc.CreateServer(me)
	go server.Run()
	fmt.Printf("running\n")
	server.Serve(sock)

}

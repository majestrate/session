# Unlimited Bot Works

This repo is for archer class session bots.

note: The Archer class really is made up of archers.

dont use this repo it may brick your system at this time.

## requirements

* go >= 1.16
* protoc
* protoc-gen-go

## building

check source code out:

    $ git clone https://github.com/majestrate/ubw
    $ cd ubw
    
building the bot:

    $ go generate ./...
    $ go build ./cmd/archer

## running

echobot:

    $ ./archer

custom message handler:

    $ ./archer ./example/reply.sh

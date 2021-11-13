# session


session bot api

## requirements

* go >= 1.16
* protoc
* protoc-gen-go


## building

check source code out:

    $ git clone https://github.com/majestrate/session
    $ cd session
echo bot:

    $ go generate ./...
    $ go build -a -v ./cmd/echobot

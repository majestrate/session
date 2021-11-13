package irc

import (
	"fmt"
	"github.com/majestrate/ubw/lib/client"
	irc "gopkg.in/sorcix/irc.v2"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	client     *client.Client
	channels   sync.Map
	conns      sync.Map
	sendMsg    chan [3]string
	recvMsg    chan [3]string
	fetchTimer *time.Ticker
	prefix     irc.Prefix
}

func (s *Server) onPrivMsg(to, nick, msg string) {
	ichnl, ok := s.channels.Load(to)
	if !ok {
		fmt.Printf("!!! %s %s %s\n", to, nick, msg)
		return
	}
	chnl := ichnl.(*ircChannel)
	chnl.onPrivMsg(nick, msg)
}

type ircChannel struct {
	conns sync.Map
	name  string
}

func (c *ircChannel) hasNick(name string) (ok bool) {
	_, ok = c.conns.Load(name)
	return
}

func (c *ircChannel) joinConn(conn *ircConn) {
	c.conns.Store(conn.prefix.Name, conn)
	conn.sendJoin(c.name)
}

func (c *ircChannel) partConn(conn *ircConn) {
	c.conns.Delete(conn.prefix.Name)
	conn.sendPart(c.name)
}

func (c *ircChannel) onPrivMsg(from, msg string) {
	fmt.Printf("on privmsg %s %s\n", from, msg)
	c.conns.Range(func(_, iconn interface{}) bool {
		conn := iconn.(*ircConn)
		fmt.Printf("%s\n", msg)
		conn.sendPrivmsg(c.name, from, msg)
		return true
	})
}

type ircConn struct {
	conn   *irc.Conn
	server *Server
	prefix irc.Prefix
}

func (c *ircConn) sendJoin(name string) {
	c.conn.Encode(&irc.Message{
		Prefix:  &c.prefix,
		Command: "JOIN",
		Params:  []string{name},
	})

}

func (c *ircConn) sendPrivmsg(dst, from, msg string) {
	c.conn.Encode(&irc.Message{
		Prefix: &irc.Prefix{
			Name: from,
		},
		Command: "PRIVMSG",
		Params:  []string{dst, msg},
	})

}

func (c *ircConn) sendPart(name string) {
	c.conn.Encode(&irc.Message{
		Prefix:  &c.server.prefix,
		Command: "PART",
		Params:  []string{name},
	})
}

func (s *Server) Run() {
	for {
		select {
		case msgParts := <-s.sendMsg:
			s.client.SendToHash(msgParts[1], msgParts[2])
		case <-s.fetchTimer.C:
			s.client.Update()
			go s.fetchAll()
		case msgParts := <-s.recvMsg:
			fmt.Printf("got new message %q\n", msgParts)
			s.onPrivMsg(msgParts[0], msgParts[1], msgParts[2])
		}
	}
}

func (s *Server) fetchAll() {
	var channels []string
	s.channels.Range(func(key, _ interface{}) bool {
		channels = append(channels, key.(string))
		return true
	})
	for _, chnl := range channels {
		msgs, _ := s.client.RecvFromHash(chnl)
		for _, msg := range msgs {
			s.recvMsg <- [3]string{chnl, msg.From(), msg.IRCLine()}
		}
	}
	msgs, _ := s.client.FetchNewMessages()
	for _, msg := range msgs {
		data, err := s.client.DecryptMessage(msg)
		if err == nil {
			fmt.Printf("new msg %q\n", data)
		}
	}
}

func (s *Server) queueSendFrom(src, dst, msg string) {
	s.sendMsg <- [3]string{src, dst, msg}
}

func (s *Server) joinRoom(dst string, conn *ircConn) {
	ichnl, loaded := s.channels.LoadOrStore(dst, new(ircChannel))
	chnl := ichnl.(*ircChannel)
	if !loaded {
		// init channel
		fmt.Printf("new channel: %s\n", dst)
		chnl.name = dst
	}
	chnl.joinConn(conn)
}

func (s *Server) partRoom(dst string, conn *ircConn) {
	ichnl, ok := s.channels.Load(dst)
	if !ok {
		return
	}
	chnl := ichnl.(*ircChannel)
	chnl.partConn(conn)
}

func CreateServer(cl *client.Client) *Server {
	return &Server{
		client:     cl,
		sendMsg:    make(chan [3]string),
		recvMsg:    make(chan [3]string),
		fetchTimer: time.NewTicker(time.Second),
		prefix: irc.Prefix{
			Name: "anonserv",
		},
	}
}

func (s *Server) serveConn(c *ircConn) {
	for {
		msg, err := c.conn.Decode()
		if err != nil {
			c.conn.Close()
			return
		}
		fmt.Printf("%q\n", msg)
		switch msg.Command {
		case irc.PING:
			{
				c.conn.Encode(&irc.Message{
					Prefix:  &s.prefix,
					Command: irc.PONG,
					Params:  msg.Params,
				})
			}
		case irc.PRIVMSG:
			{
				s.queueSendFrom(c.prefix.String(), msg.Param(0), msg.Trailing())
			}
		case irc.USER:
			{
				if c.prefix.Name == "" {
					c.prefix.Name = msg.Param(0)
					c.prefix.User = msg.Param(3)
					c.greet()
				}
			}
		case irc.JOIN:
			{
				for _, room := range strings.Split(msg.Param(0), ",") {
					s.joinRoom(room, c)
				}
			}

		case irc.PART:
			{
				for _, room := range strings.Split(msg.Param(0), ",") {
					s.partRoom(room, c)
				}
			}
		}
	}
}

func (c *ircConn) greet() {

	parts := [][]string{
		[]string{irc.RPL_WELCOME, "suh"},
		[]string{irc.RPL_YOURHOST, "you are here"},
		[]string{irc.RPL_CREATED, "server was made a while ago"},
		[]string{irc.RPL_MYINFO, "my info goes here"},
		[]string{irc.RPL_ISUPPORT, "CHANTYPES=#"},
		[]string{irc.RPL_MOTDSTART, "bruh", "bruh"},
		[]string{irc.RPL_MOTD, "do a /join #bruh"},
		[]string{irc.RPL_ENDOFMOTD, "done"},
	}

	for _, part := range parts {
		var params []string
		params = append(params, c.prefix.Name)
		params = append(params, part[1:]...)
		c.conn.Encode(&irc.Message{
			Prefix:  &c.server.prefix,
			Command: part[0],
			Params:  params,
		})
	}
}

func (s *Server) Serve(l net.Listener) error {
	for {
		sock, err := l.Accept()
		if err != nil {
			return err
		}
		host, _, _ := net.SplitHostPort(sock.RemoteAddr().String())
		conn := &ircConn{conn: irc.NewConn(sock), server: s, prefix: irc.Prefix{Host: host}}
		go s.serveConn(conn)
	}
}

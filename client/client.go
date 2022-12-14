// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package client handles all network communication to and from a player.
package client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/mailbox"
	"code.wolfmud.org/WolfMUD.git/term"
	"code.wolfmud.org/WolfMUD.git/text"
)

const inputBufferSize = 80

type pkgConfig struct {
	logClient       bool
	accountMin      int
	passwordMin     int
	saltLength      int
	frontendTimeout time.Duration
	ingameTimeout   time.Duration
	debugPanic      bool
	greeting        string
	playerPath      string
}

// cfg setup by Config and should be treated as immutable and not changed.
var cfg pkgConfig

// Config sets up package configuration for settings that can't be constants.
// It should be called by main, only once, before anything else starts. Once
// the configuration is set it should be treated as immutable an not changed.
func Config(c config.Config) {
	cfg = pkgConfig{
		logClient:       c.Server.LogClient,
		accountMin:      c.Login.AccountLength,
		passwordMin:     c.Login.PasswordLength,
		saltLength:      c.Login.SaltLength,
		frontendTimeout: c.Login.Timeout,
		ingameTimeout:   c.Server.IdleTimeout,
		debugPanic:      c.Debug.Panic,
		greeting:        c.Greeting + "\n",
		playerPath:      filepath.Join(c.Server.DataPath, "players"),
	}
}

type client struct {
	*core.Thing
	*net.TCPConn
	input  []byte
	err    chan error
	queue  <-chan string
	quit   chan struct{}
	uid    string // Can't touch c.As[core.UID] when not under BWL
	width  int    // Width of player's terminal
	height int    // Height of player's terminal
	rseq   []byte // Escape sequence for resetting terminal
	oseq   []byte // Escape sequence for updating the output terminal area
	iseq   []byte // Escape sequence for updating the input terminal area
}

func New(conn *net.TCPConn) *client {
	c := &client{
		Thing:   core.NewThing(),
		TCPConn: conn,
		input:   make([]byte, inputBufferSize),
		err:     make(chan error, 1),
		quit:    make(chan struct{}, 1),
	}

	c.err <- nil

	c.width, c.height = term.GetSize(conn)
	c.Write(term.Setup(c.width, c.height))
	c.rseq = term.Reset(c.height)
	c.oseq = term.Output(c.height)
	c.iseq = term.Input(c.height)
	c.eat()

	c.SetKeepAlive(true)
	c.SetLinger(10)
	c.SetNoDelay(false)
	c.SetWriteBuffer(80 * 24)
	c.SetReadBuffer(inputBufferSize)

	c.queue = mailbox.Add(c.As[core.UID])
	c.uid = c.As[core.UID]

	if cfg.logClient {
		c.Log("connection from: %s", c.RemoteAddr())
	}

	return c
}

func (c *client) Play() {
	go c.messenger()
	if c.frontend() {
		c.enterWorld()
		c.receive()
	}
	c.cleanup()
}

// Log takes the same parameters as fmt.Printf and writes the resulting
// message to the log. The message will automatically be appended with the UID
// uniquely identifying the connection with the current log, for example:
//
//	[#UID-6] connection from: 127.0.0.1:35848
//
// If the server configuration value Server.LogClient is set to false then an
// attempt is made to rewrite the connecting IP address as "???". For example:
//
//	[#UID-6] connection from: ???:35848
//	[#UID-6] client error: read tcp 127.0.0.1:4001->???:35848: i/o timeout
//	[#UID-6] disconnect from: ???:35848
func (c *client) Log(f string, a ...interface{}) {
	f = fmt.Sprintf("[%s] %s", c.uid, f)

	if cfg.logClient {
		log.Printf(f, a...)
		return
	}

	f = fmt.Sprintf(f, a...)
	saddr := c.RemoteAddr().String()
	if _, port, err := net.SplitHostPort(saddr); err == nil {
		f = strings.ReplaceAll(f, saddr, "???:"+port)
	}
	log.Print(f)
}

func (c *client) cleanup() {
	if err := c.error(); err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			mailbox.Send(c.uid, true,
				text.Bad+"\nIdle connection terminated by server.\n"+text.Reset,
			)
		}
		c.Log("client error: %s", err)
	}

	if cfg.logClient {
		c.Log("disconnect from: %s", c.RemoteAddr())
	}
	mailbox.Send(c.uid, true, text.Good+"\nBye bye!\n\n")

	mailbox.Delete(c.uid)
	<-c.quit

	c.Write(c.rseq)

	// Grab the BRL before player clean-up as player has been in the world
	core.BWL.Lock()
	defer core.BWL.Unlock()

	if c.As[core.Account] != "" {
		accountsMux.Lock()
		defer accountsMux.Unlock()
		delete(accounts, c.As[core.Account])
	}

	c.Free()
	c.Close()
}

func (c *client) receive() {

	s := core.NewState(c.Thing)

	// If a client panics we don't want to bring the whole server down...
	if !cfg.debugPanic {
		defer func() {
			if err := recover(); err != nil {
				c.setError(errors.New("client panicked"))
				c.Log("client panicked: %s\n%s", err, debug.Stack())
				s.Script("$QUIT")
			}
		}()
	}

	cmd := s.Script("$POOF")

	var err error
	r := bufio.NewReaderSize(c, inputBufferSize)
	for cmd != "QUIT" && c.error() == nil {
		c.input = c.input[:0]
		c.SetReadDeadline(time.Now().Add(cfg.ingameTimeout))
		if c.input, err = r.ReadSlice('\n'); err != nil {
			if errors.Is(err, bufio.ErrBufferFull) {
				for ; errors.Is(err, bufio.ErrBufferFull); _, err = r.ReadSlice('\n') {
				}
				mailbox.Send(c.uid, true, text.Bad+"You type too much!")
				continue
			}
			c.setError(err)
			break
		}
		if len(c.queue) > 10 {
			continue
		}
		cmd = s.Parse(clean(c.input))
		runtime.Gosched()
	}

	if cmd != "QUIT" {
		cmd = s.Script("$QUIT")
	}
}

func (c *client) messenger() {
	var buf []byte

	for {
		select {
		case msg, ok := <-c.queue:
			if !ok {
				c.quit <- struct{}{}
				return
			}

			buf = buf[:0]
			if len(msg) > 0 {
				buf = append(buf, c.oseq...)
				buf = append(buf, text.Reset...)
				buf = append(buf, msg...)
				buf = append(buf, c.iseq...)
			}

			c.SetWriteDeadline(time.Now().Add(10 * time.Second))
			c.Write(text.Fold(buf, c.width-2))
		}
	}
}

// error returns the first error raised or nil if there have been no errors.
func (c *client) error() error {
	e := <-c.err
	c.err <- e
	return e
}

// setError records the first error raised only, which can be retrieved by
// calling error.
func (c *client) setError(err error) {
	e := <-c.err
	if e == nil {
		e = err
	}
	c.err <- e
}

// clean incoming data. Invalid runes or C0 and C1 control codes are dropped.
// An exception in the C0 control code is backspace ('\b', ASCII 0x08) which
// will erase the previous rune. This can occur when the player's Telnet client
// does not support line editing.
func clean(in []byte) string {

	o := make([]rune, len(in)) // oversize due to len = bytes
	i := 0
	for _, v := range bytes.Runes(in) {
		switch {
		case v == '\uFFFD':
			// drop invalid runes
		case v == '\b' && i > 0:
			i--
		case v <= 0x1F:
			// drop C0 control codes
		case 0x80 <= v && v <= 0x9F:
			// drop C1 control codes
		default:
			o[i] = v
			i++
		}
	}

	return string(o[:i])
}

// eat consumes any pending incoming client data.
func (c *client) eat() {
	b := []byte{10: 0}
	c.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	for n, err := c.Read(b); err == nil && n > 0; n, err = c.Read(b) {
		c.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	}
}

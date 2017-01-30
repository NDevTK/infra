package main

import (
	irc "github.com/thoj/go-ircevent"
)

var (
	makeIRC = func(cfg *botConfig) ircInterface {
		conn := irc.IRC(cfg.Nickname, cfg.Nickname)
		conn.UseTLS = cfg.UseSSL
		return conn
	}
)

type ircInterface interface {
	Connect(string) error
	AddCallback(string, func(e *irc.Event)) string
	Join(string)
	Privmsg(string, string)
}

type ircImpl struct {
	*irc.Connection
}

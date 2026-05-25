// SPDX-FileCopyrightText: Copyright (c) The go-mail Authors
//
// SPDX-License-Identifier: MIT

// Package log implements a logger interface that can be used within the go-mail package
package log

const (
	DirServerToClient Direction = iota // Server to Client communication
	DirClientToServer                  // Client to Server communication
)

const (
	// LevelError is the Level for only ERROR log messages
	LevelError Level = iota
	// LevelWarn is the Level for WARN and higher log messages
	LevelWarn
	// LevelInfo is the Level for INFO and higher log messages
	LevelInfo
	// LevelDebug is the Level for DEBUG and higher log messages
	LevelDebug
)

const (
	// DirString is a constant used for the structured logger
	DirString = "direction"
	// DirFromString is a constant used for the structured logger
	DirFromString = "from"
	// DirToString is a constant used for the structured logger
	DirToString = "to"
)

// Direction is a type wrapper for the direction a debug log message goes
type Direction int

// Level is a type wrapper for an int
type Level int

// Log represents a log message type that holds a log Direction, a Format string
// and a slice of Messages
type Log struct {
	Direction Direction
	Format    string
	Messages  []interface{}
}

// Logger is the log interface for go-mail
type Logger interface {
	Debugf(Log)
	Infof(Log)
	Warnf(Log)
	Errorf(Log)
}

// directionPrefix will return a prefix string depending on the Direction.
func (l Log) directionPrefix() string {
	p := "C <-- S:"
	if l.Direction == DirClientToServer {
		p = "C --> S:"
	}
	return p
}

// directionFrom will return a from direction string depending on the Direction.
func (l Log) directionFrom() string {
	p := "server"
	if l.Direction == DirClientToServer {
		p = "client"
	}
	return p
}

// directionTo will return a to direction string depending on the Direction.
func (l Log) directionTo() string {
	p := "client"
	if l.Direction == DirClientToServer {
		p = "server"
	}
	return p
}

package main

import "log"

var DebugLevel = 3

func DebugOutput(level int, v ...interface{}) {
	if level <= DebugLevel {
		log.Println(v)
	}
}

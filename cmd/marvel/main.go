package main

import (
	"github.com/hauxe/xendit_pratice/server"
)

func main() {
	// start server
	s, err := server.NewServer()
	if err != nil {
		panic(err)
	}
	s.Start()
}

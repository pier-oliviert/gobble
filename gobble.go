package main

import (
	"github.com/stianeikeland/go-rpio"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var mutex sync.Mutex

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := rpio.Open(); err != nil {
		log.Fatal("Couldn't initialize GPIO pins")
	}
}

func main() {

	path := "/tmp/gobble.sock"
	if syscall.Getuid() != 0 {
		log.Fatal("Root privilege required to handle GPIO")
	}

	defer rpio.Close()

	InitializePins(gpios)

	addr, err := net.ResolveUnixAddr("unix", path)
	handleFatalErr(err)

	ln, err := net.ListenUnix("unix", addr)
	handleFatalErr(err)

	file, err := ln.File()
	handleFatalErr(err)
	defer closeSocket(ln)
	go exit(ln)

	info, err := file.Stat()
	handleFatalErr(err)

	err = os.Chmod(path, info.Mode()|0777)
	handleFatalErr(err)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Connection Error: ", err)
		}
		AddClient(conn)
	}
}

func handleFatalErr(err error) {
	if err != nil {
		log.Fatalf("*** Fatal Error: %s", err)
	}
}

func closeSocket(ln *net.UnixListener) {
	ln.Close()
}

func exit(ln *net.UnixListener) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		ln.Close()
		os.Exit(0)
	}(sigc)
}

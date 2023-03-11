package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	defer midi.CloseDriver()
	in, err := midi.FindInPort(os.Args[1])
	if err != nil {
		return fmt.Errorf("failed finding midi device: %w", err)
	}

	m := make(chan []byte)
	stop, err := midi.ListenTo(in, func(msg midi.Message, timestamps int32) {
		j, err := Encode(msg)
		if err != nil {
			log.Println(err)
			return
		}
		m <- j
	}, midi.UseSysEx())
	if err != nil {
		return fmt.Errorf("failed listening midi events: %w", err)
	}
	defer stop()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return fmt.Errorf("failed creating tcp server: %w", err)
	}
	defer ln.Close()

	clients := map[string]net.Conn{}

	c := make(chan string)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			go func(conn net.Conn) {
				addr := conn.RemoteAddr().String()
				defer func() {
					log.Println("closing", addr)
					conn.Close()
					c <- addr
				}()
				clients[addr] = conn
				_, err := io.Copy(io.Discard, conn)
				if err != nil {
					log.Println(err)
					return
				}
			}(conn)
		}
	}()

	for {
		select {
		case addr := <-c:
			log.Println("removing", addr)
			delete(clients, addr)
		case msg := <-m:
			for addr, conn := range clients {
				log.Println("sending to", addr)
				_, err := conn.Write([]byte(fmt.Sprintln(string(msg))))
				if err != nil {
					return err
				}
			}
		}
	}
}

type ControlChange struct {
	Channel    uint8
	Controller uint8
	Value      uint8
}

func Encode(m midi.Message) ([]byte, error) {
	var channel, val1, val2 uint8
	switch {
	case m.GetControlChange(&channel, &val1, &val2):
		return json.Marshal(ControlChange{
			Channel:    channel,
			Controller: val1,
			Value:      val2,
		})
	default:
		return nil, errors.New("unimplemented type")
	}
}

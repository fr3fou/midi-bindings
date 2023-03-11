package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"

	"github.com/neovim/go-client/nvim/plugin"
)

type ControlChange struct {
	Channel    uint8
	Controller uint8
	Value      uint8
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	// establish a connection to the server
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		return err
	}
	defer conn.Close()

	var pl *plugin.Plugin
	go func() {
		for {
			if pl == nil {
				continue
			}

			msg, err := bufio.NewReader(conn).ReadBytes('\n')
			if err != nil {
				log.Println(err)
				continue
			}
			m := ControlChange{}
			if err := json.Unmarshal(msg, &m); err != nil {
				log.Println(err)
				continue
			}

			input := "i"
			if m.Value == 127 {
				input = "<Esc>"
			}

			_, err = pl.Nvim.Input(input)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}()

	plugin.Main(func(p *plugin.Plugin) error {
		pl = p
		return nil
	})

	return nil
}

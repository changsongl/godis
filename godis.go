package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var server = &Server{}

const AppVersion = "Godis 0.1"

type Server struct {
	Pid            int
	Port           int
	Clients        int64
	Commands       map[string]Command
	PubSubChannels *map[string]*List
	Db             *Database
}

type Client struct {
	Cmd   Command
	Argv  []*Object
	Argc  int
	Query string
	Buff  []byte
}

func InitServer() {
	server.Pid = os.Getpid()
	server.Port = 6379
	server.Db = NewDatabase()
	server.populateCommandTable()
}

func RunServer() {
	addr := fmt.Sprintf("127.0.0.1:%d", server.Port)
	netListen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listen tcp failed: %v", err)
	}

	for {
		conn, err := netListen.Accept()
		if err != nil {
			log.Printf("netListen accept error: %v\n", err)
			continue
		}

		go Handle(conn)
	}

	defer netListen.Close()
}

func Handle(conn net.Conn) {
	client := NewClient()
	for {
		err := client.ReadQuery(conn)
		if err != nil {
			return
		}

		// Process query
		err = client.ProcessInput()
		if err != nil {
			log.Println("conn.Read error:", err)
			continue
		}
		// Process command and Response
		client.HandleCmd()
	}
}

func (s *Server) populateCommandTable() {
	s.Commands = map[string]Command{
		"get": GetCommand,
	}
}

func (s *Server) getCommand(cmd string) (Command, bool) {
	c, exists := s.Commands[cmd]
	return c, exists
}

func NewClient() *Client {
	return &Client{
		Buff: make([]byte, 512),
	}
}

func (cl *Client) ReadQuery(conn net.Conn) error {
	n, err := conn.Read(cl.Buff)
	if err != nil {
		log.Println("conn.Read error:", err)
		_ = conn.Close()
		return err
	}

	cl.Query = string(cl.Buff[:n])
	return nil
}

func (cl *Client) ProcessInput() error {
	inputs := strings.Split(cl.Query, " ")

	for idx, input := range inputs {

		if idx == 0 {
			if cmd, exists := server.getCommand(input); exists {
				cl.Cmd = cmd
			} else {
				return errors.New("invalid command")
			}
		} else {
			cl.Argv[idx-1] = CreateObject(ObjectTypeString, input)
		}
	}

	cl.Argc = len(inputs)

	return nil
}

func (cl *Client) HandleCmd() {
	//cl.Cmd(cl, server)
}

func SmoothExit() {
	fmt.Println("Handle finish")
	fmt.Println("bye")
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	maxClients   = 1000
	defaultPort  = "7711"
	welcomeMsg   = "Welcome to Simple Chat! Use /nick <nickname> to set your nick name.\n"
	unsupportedCmd = "Unsupported command\n"
)

var (
	port = flag.String("port", defaultPort, "Port for the chat server to listen on")

	clients    = make(map[net.Conn]*client)
	clientsMtx sync.Mutex
)

type client struct {
	conn net.Conn
	nick string
}

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		fmt.Println("Error listening on port", *port, ":", err)
		return
	}
	defer listener.Close()

	fmt.Println("Chat server started on port", *port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client connection:", err)
			continue
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	client := &client{
		conn: conn,
		nick: "user:" + conn.RemoteAddr().String(),
	}

	clientsMtx.Lock()
	clients[conn] = client
	clientsMtx.Unlock()
    fmt.Println("Connected client", conn.RemoteAddr().String()," ", client.nick)

	conn.Write([]byte(welcomeMsg))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		if strings.HasPrefix(msg, "/") {
			handleCommand(client, msg)
		} else {
			broadcast(client, msg)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from client", client.nick, ":", err)
	}

	clientsMtx.Lock()
	delete(clients, conn)
	clientsMtx.Unlock()

	fmt.Println("Disconnected client", client.nick)
}

func handleCommand(client *client, cmd string) {
	cmdAndArgs := strings.SplitN(strings.TrimSpace(cmd), " ", 2)
	if len(cmdAndArgs) == 0 {
		return
	}

	switch cmdAndArgs[0] {
	case "/nick":
		if len(cmdAndArgs) == 2 {
			client.nick = cmdAndArgs[1]
		} else {
			client.conn.Write([]byte(unsupportedCmd))
		}
	default:
		client.conn.Write([]byte(unsupportedCmd))
	}
}

func broadcast(sender *client, msg string) {
	formattedMsg := fmt.Sprintf("%s> %s\n", sender.nick, msg)
	fmt.Print(formattedMsg)

	clientsMtx.Lock()
	defer clientsMtx.Unlock()

	for _, client := range clients {
		if client != sender {
			client.conn.Write([]byte(formattedMsg))
		}
	}
}
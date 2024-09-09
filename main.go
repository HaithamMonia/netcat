package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// Message structure to represent a message sent from a client
type Message struct {
	from      string
	payload   string
	timestamp string
}

// Client struct to represent a connected client
type Client struct {
	conn     net.Conn
	username string
}

// Server struct to represent the chat server
type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan Message
	joinch     chan Client
	leavech    chan Client
	clients    map[net.Conn]string
	history    []Message
	mu         sync.Mutex
}

// NewServer initializes a new chat server
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message, 10),
		joinch:     make(chan Client),
		leavech:    make(chan Client),
		clients:    make(map[net.Conn]string),
		history:    make([]Message, 0),
	}
}

// start initiates the server to listen for incoming connections
func (s *Server) start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.ln.Close()
		s.mu.Unlock()
	}()

	go s.acceptLoop()
	<-s.quitch
	close(s.msgch)
	close(s.joinch)
	close(s.leavech)
	return nil
}

// handleConnections manages join, leave, and message broadcasting
func (s *Server) handleConnections() {
	for {
		select {
		case client := <-s.joinch:
			s.clients[client.conn] = client.username
			joinMsg := Message{
				from:      "Server",
				payload:   fmt.Sprintf("%s has joined our chat...", client.username),
				timestamp: time.Now().Format("2006-01-02 15:04:05"),
			}
			s.history = append(s.history, joinMsg)
			s.broadcast(joinMsg, client.conn) // Exclude the joining client

			// Send chat history
			s.sendHistory(client.conn)
		case client := <-s.leavech:
			leaveMsg := Message{
				from:      "Server",
				payload:   fmt.Sprintf("%s has left our chat...", client.username),
				timestamp: time.Now().Format("2006-01-02 15:04:05"),
			}
			s.history = append(s.history, leaveMsg)
			s.broadcast(leaveMsg, client.conn) // Exclude the leaving client
			delete(s.clients, client.conn)
		case msg := <-s.msgch:
			s.history = append(s.history, msg)
			s.broadcast(msg, nil) // Broadcast to all clients
		}
	}
}


// acceptLoop handles incoming client connections
func (s *Server) acceptLoop() {
	for {
		s.mu.Lock()
		conn, err := s.ln.Accept()
		s.mu.Unlock()

		if err != nil {
			fmt.Println("Accept Error: ", err)
			continue
		}
		fmt.Println("New connection to the server:", conn.RemoteAddr())

		go s.handleNewClient(conn)
	}
}

// handleNewClient handles the process of sending ASCII art, asking for a username, and joining the client to the server
func (s *Server) handleNewClient(conn net.Conn) {
	// Send the welcome message with ASCII art
	s.sendAsciiArt(conn)
	conn.Write([]byte("[ENTER YOUR NAME]: "))

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Username Read Error:", err)
		conn.Close()
		return
	}
	username := string(buf[:n-1]) // removing the newline character
	if len(username) == 0 {
		conn.Write([]byte("Username cannot be empty. Connection will be closed.\n"))
		conn.Close()
		return
	}
	client := Client{conn: conn, username: username}
	s.joinch <- client
	go s.readLoop(client)
}

func (s *Server) readLoop(client Client) {
	defer func() {
		s.leavech <- client
		client.conn.Close()
	}()

	buf := make([]byte, 2048)
	for {
		n, err := client.conn.Read(buf)
		if err != nil {
			fmt.Println("Read Error:", err)
			return
		}
		msg := Message{
			from:      client.username,
			payload:   string(buf[:n-1]), // removing the newline character
			timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		s.msgch <- msg

		// Remove this part to avoid sending empty messages:
		/*
		s.msgch <- Message{
			from:      client.username,
			payload:   "",
			timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		*/
	}
}


// broadcast sends a message to all connected clients except the sender
func (s *Server) broadcast(msg Message, senderConn net.Conn) {
	var formattedMessage string
	if msg.from == "Server" {
		formattedMessage = fmt.Sprintf("%s\n", msg.payload)
	} else {
		formattedMessage = fmt.Sprintf("[%s][%s]: %s\n", msg.timestamp, msg.from, msg.payload)
	}

	for conn := range s.clients {
		if conn != senderConn {
			_, err := conn.Write([]byte(formattedMessage))
			if err != nil {
				fmt.Println("Broadcast Error:", err)
			}
		}
	}
}


// sendHistory sends all previous messages to a newly connected client
func (s *Server) sendHistory(conn net.Conn) {
	for _, msg := range s.history {
		var formattedMessage string
		if msg.from == "Server" {
			formattedMessage = fmt.Sprintf("%s\n", msg.payload)
		} else {
			formattedMessage = fmt.Sprintf("[%s][%s]: %s\n", msg.timestamp, msg.from, msg.payload)
		}

		_, err := conn.Write([]byte(formattedMessage))
		if err != nil {
			fmt.Println("History Send Error:", err)
		}
	}
}

// sendAsciiArt reads the ASCII art from a file and sends it to a client
func (s *Server) sendAsciiArt(conn net.Conn) {
	content, err := ioutil.ReadFile("linuxLogo")
	if err != nil {
		fmt.Println("Failed to read the ASCII art file:", err)
		return
	}

	conn.Write([]byte("Welcome to TCP-Chat!\n"))
	_, err = conn.Write(content)
	if err != nil {
		fmt.Println("Failed to send the ASCII art to the client:", err)
	}
	conn.Write([]byte("\n"))
}

// main function starts the chat server
func main() {
	var port string
	if len(os.Args) == 1 {
		port = ":8989"
	} else {
		port = os.Args[1]
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			fmt.Println("[USAGE]: ./TCPChat $port")
			return
		}
		port = ":" + port
	}
	fmt.Println("Listening on the port ", port)

	server := NewServer(port)
	go server.handleConnections()
	log.Fatal(server.start())
}

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
			s.broadcast(fmt.Sprintf("User %s has joined the chat", client.username), "Server")

			// Send the ASCII art to the newly connected client
			s.sendAsciiArt(client.conn)

			// Send chat history
			s.sendHistory(client.conn)
		case client := <-s.leavech:
			delete(s.clients, client.conn)
			s.broadcast(fmt.Sprintf("User %s has left the chat", client.username), "Server")
		case msg := <-s.msgch:
			s.history = append(s.history, msg)
			s.broadcast(fmt.Sprintf("[%s][%s]: %s", msg.timestamp, msg.from, msg.payload), msg.from)
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

// handleNewClient handles the process of asking for a username and joining the client to the server
func (s *Server) handleNewClient(conn net.Conn) {
	conn.Write([]byte("Enter your username: "))
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Username Read Error:", err)
		conn.Close()
		return
	}
	username := string(buf[:n-1]) // removing the newline character
	client := Client{conn: conn, username: username}
	s.joinch <- client
	go s.readLoop(client)
}

// readLoop reads messages from a connected client and forwards them to the msgch channel
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
	}
}

// broadcast sends a message to all connected clients
func (s *Server) broadcast(message string, from string) {
	for conn := range s.clients {
		_, err := conn.Write([]byte(fmt.Sprintf("%s\n", message)))
		if err != nil {
			fmt.Println("Broadcast Error:", err)
		}
	}
}

// sendHistory sends all previous messages to a newly connected client
func (s *Server) sendHistory(conn net.Conn) {
	for _, msg := range s.history {
		_, err := conn.Write([]byte(fmt.Sprintf("[%s][%s]: %s\n", msg.timestamp, msg.from, msg.payload)))
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

	_, err = conn.Write(content)
	if err != nil {
		fmt.Println("Failed to send the ASCII art to the client:", err)
	}
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

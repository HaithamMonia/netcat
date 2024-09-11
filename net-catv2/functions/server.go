package netcat

import (
	"net"
	"sync"
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
	listenAddr  string
	ln          net.Listener
	quitch      chan struct{}
	msgch       chan Message
	joinch      chan Client
	leavech     chan Client
	clients     map[net.Conn]string
	history     []Message
	mu          sync.Mutex
	clientCount int
}

const maxClients = 10

// NewServer initializes a new chat server
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message),
		joinch:     make(chan Client),
		leavech:    make(chan Client),
		clients:    make(map[net.Conn]string),
		history:    make([]Message, 0),
	}
}

package netcat

import (
	"fmt"
	"net"
	"time"
)

func (s *Server) HandleNewClient(conn net.Conn) {
	s.SendAsciiArt(conn)
	conn.Write([]byte("[ENTER YOUR NAME]: "))

	buf := make([]byte, 2048) // Reading username
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Username Read Error:", err)
		conn.Close()
		return
	}
	username := string(buf[:n-1]) // Removing the newline character
	if s.GetConnFromUsername(username) != nil {
		conn.Write([]byte("The Username aready exists. Connection will be closed.\n"))
		conn.Close()
		return
	}
	if len(username) == 0 {
		conn.Write([]byte("Username cannot be empty. Connection will be closed.\n"))
		conn.Close()
		return
	}
	client := Client{conn: conn, username: username}

	s.mu.Lock()
	s.clients[client.conn] = client.username
	s.clientCount++
	s.mu.Unlock()

	s.SendHistory(client.conn)

	joinMsg := Message{
		from:      "Server",
		payload:   fmt.Sprintf("%s has joined our chat...", client.username),
		timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	s.history = append(s.history, joinMsg)
	s.Broadcast(joinMsg, client.conn)

	go s.ReadLoop(client)
}

package netcat

import (
	"fmt"
	"time"
)

func (s *Server) AcceptLoop() {
	for {
		s.mu.Lock()
		if s.clientCount >= maxClients {
			s.mu.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}
		s.mu.Unlock()

		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("Accept Error: ", err)
			continue
		}
		fmt.Println("New connection to the server:", conn.RemoteAddr())

		go s.HandleNewClient(conn) // renamed to exported function
	}
}

package netcat

import (
	"net"
)

func (s *Server) Start() error {
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

	go s.AcceptLoop() // Start accepting clients in a goroutine
	<-s.quitch        // Block until the server receives a signal to shut down
	close(s.msgch)
	close(s.joinch)
	close(s.leavech)
	return nil
}

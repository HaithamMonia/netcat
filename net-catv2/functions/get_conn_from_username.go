package netcat

import "net"

func (s *Server) GetConnFromUsername(from string) net.Conn {
	for conn, username := range s.clients {
		if username == from {
			return conn
		}
	}
	return nil
}

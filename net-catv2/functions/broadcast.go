package netcat

import (
	"fmt"
	"net"
)

func (s *Server) Broadcast(msg Message, excludeConn net.Conn) {
	var formattedMessage string
	if msg.from == "Server" {
		formattedMessage = fmt.Sprintf("%s\n", msg.payload)
	} else {
		formattedMessage = fmt.Sprintf("[%s][%s]: %s\n", msg.timestamp, msg.from, msg.payload)
	}

	for conn := range s.clients {
		if conn == excludeConn {
			continue
		}
		_, err := conn.Write([]byte(formattedMessage))
		if err != nil {
			fmt.Println("Broadcast Error:", err)
		}
	}
}

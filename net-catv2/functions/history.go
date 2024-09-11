package netcat

import (
	"fmt"
	"net"
)

func (s *Server) SendHistory(conn net.Conn) {
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

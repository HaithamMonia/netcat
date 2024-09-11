package netcat

import (
	"fmt"
	"time"
)

func (s *Server) ReadLoop(client Client) {
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

		if IsEmpty(string(buf[:n-1])) {
			continue
		}

		msg := Message{
			from:      client.username,
			payload:   string(buf[:n-1]), // remove newline character
			timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		s.msgch <- msg
	}
}

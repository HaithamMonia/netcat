package netcat

import (
	"fmt"
	"io/ioutil"
	"net"
)

func (s *Server) SendAsciiArt(conn net.Conn) {
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

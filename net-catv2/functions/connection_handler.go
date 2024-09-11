package netcat

import "time"

func (s *Server) HandleConnections() {
	for {
		select {
		case client := <-s.joinch:
			s.clients[client.conn] = client.username
			joinMsg := Message{
				from:      "Server",
				payload:   client.username + " has joined our chat...",
				timestamp: time.Now().Format("2006-01-02 15:04:05"),
			}
			s.history = append(s.history, joinMsg)
			s.Broadcast(joinMsg, client.conn)

			s.SendHistory(client.conn)

		case client := <-s.leavech:
			leaveMsg := Message{
				from:      "Server",
				payload:   client.username + " has left our chat...",
				timestamp: time.Now().Format("2006-01-02 15:04:05"),
			}
			s.history = append(s.history, leaveMsg)
			s.Broadcast(leaveMsg, client.conn)
			s.mu.Lock()
			delete(s.clients, client.conn)
			s.clientCount--
			s.mu.Unlock()

		case msg := <-s.msgch:
			s.history = append(s.history, msg)
			s.Broadcast(msg, s.GetConnFromUsername(msg.from))
		}
	}
}

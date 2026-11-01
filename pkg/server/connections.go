package server

import "github.com/coder/websocket"

func (s *Server) trackConn(conn *websocket.Conn) {
	s.connsMu.Lock()
	s.conns[conn] = struct{}{}
	s.connsMu.Unlock()
}

func (s *Server) untrackConn(conn *websocket.Conn) {
	s.connsMu.Lock()
	delete(s.conns, conn)
	s.connsMu.Unlock()
}

func (s *Server) closeAllConns(reason string) {
	s.connsMu.Lock()
	conns := make([]*websocket.Conn, 0, len(s.conns))
	for c := range s.conns {
		conns = append(conns, c)
	}
	s.connsMu.Unlock()

	for _, c := range conns {
		c.Close(websocket.StatusGoingAway, reason)
	}
}

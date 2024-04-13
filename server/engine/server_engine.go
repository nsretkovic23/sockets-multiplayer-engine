package engine

import (
	"encoding/json"
	"fmt"
	"net"
	"socket_server/helpers"
)

type Lobby struct {
	Id    int
	Conns []net.Conn
}

// Matchmakes players, creates and returns the lobby
// You can provide a message to be sent to all players upon connecting, if you don't want to send a message, pass nil
func MakeLobby(listener net.Listener, maxConn int, roomId int, message []byte) (*Lobby, error) {
	lobby := &Lobby{roomId, []net.Conn{}}
	// Matchmake players
	conns, err := MatchMake(listener, maxConn, lobby.Id, message)

	if err != nil {
		fmt.Println("Error while trying to matchmake players", err)
		return lobby, nil
	}
	lobby.Conns = *conns

	return lobby, nil
}

// Accepts maxConn number of connections and returns them in a slice
// If message is not nil, it sends it to all connections upon accepting the connection
func MatchMake(listener net.Listener, maxConn int, roomId int, message []byte) (*[]net.Conn, error) {
	conns := &[]net.Conn{}

	for len(*conns) < maxConn {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		*conns = append(*conns, conn)

		if message != nil {
			_, err = conn.Write([]byte(message))
			if err != nil {
				fmt.Println("Error sending message to connected player:", err)
				continue
			}
		}

		helpers.PrintGreen(fmt.Sprintf("[%d/%d] Player in room %d connected: %s", len(*conns), maxConn, roomId, conn.RemoteAddr().String()))
	}

	return conns, nil
}

// Formats the message of any type into a byte slice
// Returns nil as value if there was an error while formatting the message
func FormatMessage[V any](message V) []byte {
	msgJSON, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error while trying to format message", err)
		return nil
	}

	return msgJSON
}

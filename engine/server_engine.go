package engine

import (
	"encoding/json"
	"fmt"
	"net"
	"sockets-multiplayer/helpers"
)

type Lobby struct {
	Id    int
	Conns []net.Conn
}

/*
	TODO: Add timer for making a lobby
	TODO: Add timer for matchmaking
*/

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
// If message is not nil, it sends message to all connections upon accepting the connection
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
			SendUnicastMessage(&conn, message)
		}

		helpers.PrintGreen(fmt.Sprintf("[%d/%d] Player in room %d connected: %s", len(*conns), maxConn, roomId, conn.RemoteAddr().String()))
	}

	return conns, nil
}

func SendUnicastMessage(conn *net.Conn, message []byte) (int, error) {
	if message == nil {
		return 0, fmt.Errorf("provided message is nil")
	}

	nBytes, err := (*conn).Write(message)
	if err != nil {
		return 0, fmt.Errorf("error sending message to connection: %v", err)
	}

	return nBytes, nil
}

func SendMulticastMessage(conns *[]net.Conn, message []byte) error {
	if message == nil {
		return fmt.Errorf("provided message is nil")
	}

	for _, conn := range *conns {
		_, err := conn.Write(message)
		if err != nil {
			return fmt.Errorf("error sending message to connection: %v", err)
		}
	}

	return nil
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

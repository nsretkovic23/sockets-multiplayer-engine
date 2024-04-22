package engine

import (
	"encoding/json"
	"fmt"
	"net"
	"sockets-multiplayer/helpers"
	"time"
)

const (
	HEARTBEAT_TIMEOUT = 1
)

// Conns is a slice (array) for now, but it can be changed to a map if needed
// I kept it as a slice because I don't know if conn's address will always be unique
type Lobby struct {
	Id    int
	Conns []net.Conn
}

/*
	TODO: Add timer for making a lobby
	TODO: Add timer for matchmaking
*/

// Accepts connections, creates and returns the lobby
// You can provide a message to be sent to all players upon connecting, if you don't want to send a message, pass nil
func MakeLobby(listener net.Listener, maxConn int, lobbyId int, message []byte) (*Lobby, error) {
	lobby := &Lobby{lobbyId, []net.Conn{}}
	// Matchmake players
	conns, err := AcceptClients(listener, maxConn, lobby.Id, message)

	if err != nil {
		fmt.Println("Error while trying to matchmake players", err)
		return lobby, nil
	}
	lobby.Conns = *conns

	return lobby, nil
}

// Accepts maxConn number of connections and returns them in a slice
// If message is not nil, it sends message to all connections upon accepting the connection
func AcceptClients(listener net.Listener, maxConn int, lobbyId int, message []byte) (*[]net.Conn, error) {
	conns := &[]net.Conn{}
	stopLobbyHeartbeat := false

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

		helpers.PrintGreen(fmt.Sprintf("Lobby %d [%d/%d] Player connected: %s", lobbyId, len(*conns), maxConn, conn.RemoteAddr().String()))

		// Handling possible client disconnect during the matchmaking/lobby creating process, while lobby is not fully complete
		disconnected := make(chan *net.Conn)
		// Start the heartbeat for the connection
		go LobbyHeartbeat(lobbyId, &conn, disconnected, &stopLobbyHeartbeat)

		go func() {
			dcClient := <-disconnected
			// Condition to exit the goroutine
			// nil is sent to the channel when lobby is full, this ultimately means that the client didn't disconnect during the matchmaking/lobby making process
			if dcClient == nil {
				return
			}

			err := RemoveConn(dcClient, conns)
			if err == nil {
				helpers.PrintRed(fmt.Sprintf("Lobby %d: Client disconnected and removed from the channel: %v", lobbyId, (*dcClient).RemoteAddr().String()))
			}
		}()

	}

	// Since this variable is passed by reference to the goroutines that heartbeat the connections, we can stop the every heartbeat goroutine by setting it to true here
	stopLobbyHeartbeat = true
	time.Sleep((HEARTBEAT_TIMEOUT + 1) * time.Second)
	return conns, nil
}

// Checks every HEARTBEAT_TIMEOUT seconds (number should be as low as possible (preferrably 1)) if the connection is still alive while clients are currently in a lobby that is not yet full
func LobbyHeartbeat(lobbyId int, conn *net.Conn, disconnected chan *net.Conn, stop *bool) {
	helpers.PrintInfo(fmt.Sprintf("Lobby %d : Starting Lobby heartbeat for the connection %v", lobbyId, (*conn).RemoteAddr().String()))
	// Make a minimal buffer that will just serve for the read call
	buff := make([]byte, 1)

	// Set timeout to one second so that the Read call acts as a heartbeat
	for !*stop {
		// Timeout needs to be set every iteration
		(*conn).SetReadDeadline(time.Now().Add(HEARTBEAT_TIMEOUT * time.Second))
		// Try to read from the connection
		_, err := (*conn).Read(buff)
		// Error can be a timeout, or a closed connection
		if err != nil {
			// Check if the connection is closed by the client
			if netErr, ok := err.(net.Error); ok && !netErr.Timeout() {
				disconnected <- conn
				break
			}
			// If deadline/timeout exceeded, continue checking the stop condition
		}
	}

	disconnected <- nil
	helpers.PrintYellow(fmt.Sprintf("Lobby %d : Stopping lobby heartbeat for: %v", lobbyId, (*conn).RemoteAddr()))
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

func IsTimeoutError(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}

func RemoveConn(conn *net.Conn, arr *[]net.Conn) error {
	for i, c := range *arr {
		if *conn == c {
			*arr = append((*arr)[:i], (*arr)[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("connection not found in the lobby")
}

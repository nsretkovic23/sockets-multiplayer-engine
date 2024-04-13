package main

import (
	"fmt"
	"net"
	"socket_server/engine"
	"socket_server/helpers"
)

// TODO: These should be in a config file, environment variables or command line arguments
const (
	SMILEY      = "\U0001F604"
	MAX_CONN    = 2
	MAX_LOBBIES = 1
	PORT        = 8080
	SECRET_WORD = "MY WORD"
)

type Message struct {
	Text string
}

func main() {
	helpers.PrintInfo("Starting server...")

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", PORT))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	helpers.PrintInfo(fmt.Sprintf("Server started, listening on port %d", PORT))

	var lobbies []engine.Lobby
	for len(lobbies) < MAX_LOBBIES {

		lobby, err := engine.MakeLobby(listener, MAX_CONN, len(lobbies),
			engine.FormatMessage[*Message](&Message{"Welcome to the game!"}))
		if err != nil {
			fmt.Println("Error while trying to make a lobby", err)
			continue
		}

		fmt.Println("Lobby created", lobby.Id)

		lobbies = append(lobbies, *lobby)
		go runGame(lobby)
	}

	for len(lobbies) > 0 {

	}
}

func runGame(lobby *engine.Lobby) {
	fmt.Println("Running game")
	if lobby.Conns == nil || len(lobby.Conns) <= 1 {
		helpers.PrintRed("Not enough players in the lobby")
		return
	}

	for i, conn := range lobby.Conns {
		msg := &Message{fmt.Sprintf("You are player %d", i)}
		_, err := engine.SendUnicastMessage(&conn, engine.FormatMessage[*Message](msg))
		if err != nil {
			fmt.Println("Error sending message to player", err)
		}
	}

	for len(lobby.Conns) > 0 {
		// playerMsg := make([]byte, 1024)
		// _, err := lobby.conns[0].Read(playerMsg)
		// if err != nil {
		// 	fmt.Println("Error reading player message", err)
		// }
	}
}

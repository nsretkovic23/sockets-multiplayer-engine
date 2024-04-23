package main

import (
	"fmt"
	"net"
	"sockets-multiplayer/engine"
	"sockets-multiplayer/helpers"
)

// TODO: These should be in a config file, environment variables or command line arguments
const (
	SMILEY      = "\U0001F604"
	MAX_CONN    = 2
	MAX_LOBBIES = 2
	PORT        = 8080
	SECRET_WORD = "MY WORD"
	MIN_CONN    = 2
	TIMEOUT     = 15
)

type ServerMessage struct {
	Type            string
	Text            string
	Tag             int
	Turn            int
	PreviousGuesses []string
	GuessState      string
	WinnerTag       int
	WonByDisconnect bool
}

type ClientMessage struct {
	Sender int
	Guess  string
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

	var lobbies []WordGuessLobby
	for len(lobbies) < MAX_LOBBIES {

		lobby, err := engine.MakeLobby(
			listener,
			MAX_CONN,
			len(lobbies),
			engine.FormatMessage(&ServerMessage{"welcome", "Welcome to the game!", -1, -1, []string{}, "", -1, false}),
		)
		wordGuessLobby := WordGuessLobby(*lobby)

		if err != nil {
			fmt.Println("Error while trying to make a lobby", err)
			continue
		}

		fmt.Println("Lobby created with Id:", lobby.Id)

		lobbies = append(lobbies, wordGuessLobby)
		go wordGuessLobby.RunGame()
	}

	for len(lobbies) > 0 {

	}
}

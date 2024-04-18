package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sockets-multiplayer/engine"
	"sockets-multiplayer/helpers"
	"sockets-multiplayer/samples/word_guess_game/server/word"
	"time"
)

// TODO: These should be in a config file, environment variables or command line arguments
const (
	SMILEY      = "\U0001F604"
	MAX_CONN    = 2
	MAX_LOBBIES = 1
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
}

type ClientMessage struct {
	Sender int
	Guess  string
}

type GameState struct {
	Turn            int
	PreviousGuesses []string
	GuessedLetters  map[string]bool
	Word            string
	IsWordGuessed   bool
	Message         string
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
			engine.FormatMessage[*ServerMessage](&ServerMessage{
				"welcome",
				"Welcome to the game!",
				-1,
				-1,
				[]string{},
				"",
			}))
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

	// TODO: Refactor
	state := &GameState{
		0,
		[]string{},
		word.InitializeLetterMap(),
		"NIKOLA SRETKOVIC",
		false,
		"",
	}

	// TODO: Refactor
	// WELCOME MESSAGE
	for i, conn := range lobby.Conns {
		msg := &ServerMessage{
			"tag_assignment",
			fmt.Sprintf("You are player %d", i),
			i,
			0,
			[]string{},
			word.FormatSentenceGuessState(state.Word, state.GuessedLetters),
		}
		_, err := engine.SendUnicastMessage(&conn, engine.FormatMessage[*ServerMessage](msg))
		if err != nil {
			fmt.Println("Error sending message to player", err)
		}
	}

	// TODO: Refactor
	// GAME LOOP
	for {
		// Send message to all players about the next turn
		msg := state.getNextTurnMessage()
		engine.SendMulticastMessage(&lobby.Conns, msg)

		msgRaw := make([]byte, 2048)
		lobby.Conns[state.Turn].SetReadDeadline(time.Now().Add(TIMEOUT * time.Second))
		n, err := lobby.Conns[state.Turn].Read(msgRaw)
		if err != nil {

			if engine.IsTimeoutError(err) {
				helpers.PrintRed("Client input timeout, SEND MESSAGE TO PLAYERS and continue")
			} else {
				// TODO: See if you can handle the error better here, this else is pretty general and maybe doesn't always mean that the player disconnected

				helpers.PrintRed("Player " + /*fmt.Sprintf("%d", state.Turn)*/ lobby.Conns[state.Turn].RemoteAddr().String() + " disconnected")
				lobby.Conns = append(lobby.Conns[:state.Turn], lobby.Conns[state.Turn+1:]...)

				if len(lobby.Conns) < MIN_CONN {
					helpers.PrintRed("Not enough players in the lobby, game over!")
					helpers.PrintRed("If there is a player in the lobby, message him that he won!")
					break
				}

			}

			state.Turn = (state.Turn + 1) % len(lobby.Conns)
			continue
		}

		var clientMsg ClientMessage

		err = json.Unmarshal(msgRaw[:n], &clientMsg)
		if err != nil {
			helpers.PrintRed("Error unmarshalling message: " + err.Error())
			continue
		}

		if clientMsg.Sender != state.Turn {
			helpers.PrintRed("Player " + fmt.Sprintf("%d", state.Turn) + " tried to guess out of turn")
			continue
		}

		word.ProcessInput(state.Word, clientMsg.Guess, state.GuessedLetters, &state.IsWordGuessed, &state.Message)

		if state.IsWordGuessed {
			helpers.PrintInfo("Lobby " + fmt.Sprint(lobby.Id) + "::: Player " + fmt.Sprintf("%d", state.Turn) + " guessed the word/sentence correctly, game over!")
			// Since turn is not yet incremented, the player who guessed the word is the current player, so we can send state.Turn
			msg := engine.FormatMessage(&ServerMessage{"game_over", state.Message, -1, state.Turn, state.PreviousGuesses, state.Word})
			engine.SendMulticastMessage(&lobby.Conns, msg)
			break
		}

		state.PreviousGuesses = append(state.PreviousGuesses, clientMsg.Guess)
		helpers.PrintInfo("Lobby " + fmt.Sprint(lobby.Id) + "::: Player " + fmt.Sprintf("%d", state.Turn) + " guessed: " + clientMsg.Guess)
		state.Turn = (state.Turn + 1) % len(lobby.Conns)
	}

}

func (state *GameState) getNextTurnMessage() []byte {
	return engine.FormatMessage(&ServerMessage{
		"turn",
		fmt.Sprintf("Player %d's turn", state.Turn),
		-1,
		state.Turn,
		state.PreviousGuesses,
		word.FormatSentenceGuessState(state.Word, state.GuessedLetters),
	})
}

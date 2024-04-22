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

		lobby, err := engine.MakeLobby(
			listener,
			MAX_CONN,
			len(lobbies),
			engine.FormatMessage(&ServerMessage{"welcome", "Welcome to the game!", -1, -1, []string{}, "", -1, false}),
		)

		if err != nil {
			fmt.Println("Error while trying to make a lobby", err)
			continue
		}

		fmt.Println("Lobby created with Id:", lobby.Id)

		lobbies = append(lobbies, *lobby)
		go runGame(lobby)
	}

	for len(lobbies) > 0 {

	}
}

func runGame(lobby *engine.Lobby) {
	fmt.Println(lobbyString(lobby), ": Running game")

	state := createInitialState()
	sendTagAssignments(lobby, state)

	// TODO: Refactor
	// GAME LOOP
	for {
		// Send message to all players about the next turn
		msg := state.getNextTurnMessage()
		engine.SendMulticastMessage(&lobby.Conns, msg)
		// Listen for guess from the player whose turn it is
		msgRaw := make([]byte, 4096)
		n, err := state.listenForNextTurn(lobby, &msgRaw)

		if err != nil {
			if isGameOverOnError(lobby, state, err) {
				break
			} else {
				state.Turn = (state.Turn + 1) % len(lobby.Conns)
				continue
			}
			// if engine.IsTimeoutError(err) {
			// 	helpers.PrintRed(lobbyString(lobby) + ": Client input timeout, SEND MESSAGE TO PLAYERS and continue")
			// } else {
			// 	helpers.PrintRed(lobbyString(lobby) + ": Player " + lobby.Conns[state.Turn].RemoteAddr().String() + " disconnected")
			// 	engine.RemoveConn(&lobby.Conns[state.Turn], &lobby.Conns)

			// 	if len(lobby.Conns) < MIN_CONN {
			// 		finishGameOnNotEnoughPlayers(lobby, state)
			// 		break
			// 	}
			// }
			// state.Turn = (state.Turn + 1) % len(lobby.Conns)
			// continue
		}

		var clientMsg ClientMessage

		err = json.Unmarshal(msgRaw[:n], &clientMsg)
		if err != nil {
			helpers.PrintRed("Error unmarshalling message: " + err.Error())
			continue
		}

		if clientMsg.Sender != state.Turn {
			helpers.PrintRed(lobbyString(lobby) + "Player " + fmt.Sprintf("%d", state.Turn) + " tried to guess out of turn")
			continue
		}

		word.ProcessInput(state.Word, clientMsg.Guess, state.GuessedLetters, &state.IsWordGuessed, &state.Message)

		if state.IsWordGuessed {
			helpers.PrintGreen(lobbyString(lobby) + ": Player " + fmt.Sprintf("%d", state.Turn) + " guessed the word/sentence correctly, game over!")
			// Since turn is not yet incremented, the player who guessed the word is the current player, so we can send state.Turn
			msg := engine.FormatMessage(&ServerMessage{"game_over", state.Message, -1, state.Turn, state.PreviousGuesses, state.Word, state.Turn, false})
			engine.SendMulticastMessage(&lobby.Conns, msg)
			break
		}

		state.PreviousGuesses = append(state.PreviousGuesses, clientMsg.Guess)
		helpers.PrintInfo(lobbyString(lobby) + ": Player " + fmt.Sprintf("%d", state.Turn) + " guessed: " + clientMsg.Guess)
		state.Turn = (state.Turn + 1) % len(lobby.Conns)
	}

}

func isGameOverOnError(lobby *engine.Lobby, state *GameState, err error) bool {
	if engine.IsTimeoutError(err) {
		helpers.PrintRed(lobbyString(lobby) + ": Client input timeout, SEND MESSAGE TO PLAYERS and continue")
		return false
	} else {
		helpers.PrintRed(lobbyString(lobby) + ": Player " + lobby.Conns[state.Turn].RemoteAddr().String() + " disconnected")
		engine.RemoveConn(&lobby.Conns[state.Turn], &lobby.Conns)

		if len(lobby.Conns) < MIN_CONN {
			finishGameOnNotEnoughPlayers(lobby, state)
			return true
		}
	}

	return false
}

func (state *GameState) listenForNextTurn(lobby *engine.Lobby, msgRaw *[]byte) (int, error) {
	lobby.Conns[state.Turn].SetReadDeadline(time.Now().Add(TIMEOUT * time.Second))
	return lobby.Conns[state.Turn].Read(*msgRaw)
}

func createInitialState() *GameState {
	return &GameState{
		0,
		[]string{},
		word.InitializeLetterMap(),
		"NIKOLA SRETKOVIC",
		false,
		"",
	}
}

func sendTagAssignments(lobby *engine.Lobby, state *GameState) {
	for i, conn := range lobby.Conns {
		msg := &ServerMessage{
			"tag_assignment",
			fmt.Sprintf("You are player %d", i),
			i,
			0,
			[]string{},
			word.FormatSentenceGuessState(state.Word, state.GuessedLetters),
			-1,
			false,
		}
		_, err := engine.SendUnicastMessage(&conn, engine.FormatMessage[*ServerMessage](msg))
		if err != nil {
			fmt.Println("Error sending message to player", err)
		}
	}
}

func finishGameOnNotEnoughPlayers(lobby *engine.Lobby, state *GameState) {
	helpers.PrintRed(fmt.Sprintf("Lobby %d : Not enough players in the lobby, game over", lobby.Id))
	msg := engine.FormatMessage(&ServerMessage{
		"game_over",
		"Not enough players in the lobby to continue the game. You Won!",
		-1,
		state.Turn,
		state.PreviousGuesses,
		state.Word,
		-1,
		true,
	})

	engine.SendMulticastMessage(&lobby.Conns, msg)
}

func (state *GameState) getNextTurnMessage() []byte {
	return engine.FormatMessage(&ServerMessage{
		"turn",
		fmt.Sprintf("Player %d's turn", state.Turn),
		-1,
		state.Turn,
		state.PreviousGuesses,
		word.FormatSentenceGuessState(state.Word, state.GuessedLetters),
		-1,
		false,
	})
}

func lobbyString(lobby *engine.Lobby) string {
	return fmt.Sprintf("Lobby %d", lobby.Id)
}

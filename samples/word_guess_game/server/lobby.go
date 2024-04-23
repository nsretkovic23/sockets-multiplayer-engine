package main

import (
	"encoding/json"
	"fmt"
	"sockets-multiplayer/engine"
	"sockets-multiplayer/helpers"
	"sockets-multiplayer/samples/word_guess_game/server/word"
)

type WordGuessLobby engine.Lobby

func (lobby *WordGuessLobby) String() string {
	return fmt.Sprintf("Lobby %d", lobby.Id)
}

func (lobby *WordGuessLobby) RunGame() {
	fmt.Println(lobby.String(), ": Running game")

	var state GameState
	state.Initialize()

	lobby.SendTagAssignments()

	// GAME LOOP
	for {
		// Send message to all players about the next turn
		msg := state.GetNextTurnMessage()
		engine.SendMulticastMessage(&lobby.Conns, msg)
		// Listen for guess from the player whose turn it is
		rawMsg := make([]byte, 4096)
		n, err := state.ListenForNextTurn(lobby, &rawMsg)

		if err != nil {
			if lobby.IsGameOverOnError(&state, err) {
				break
			} else {
				state.Turn = (state.Turn + 1) % len(lobby.Conns)
				continue
			}
		}

		var clientMsg ClientMessage

		err = json.Unmarshal(rawMsg[:n], &clientMsg)
		if err != nil {
			helpers.PrintRed("Error unmarshalling message: " + err.Error())
			continue
		}

		if clientMsg.Sender != state.Turn {
			helpers.PrintRed(lobby.String() + "Player " + fmt.Sprintf("%d", state.Turn) + " tried to guess out of turn")
			continue
		}

		word.ProcessInput(state.Word, clientMsg.Guess, state.GuessedLetters, &state.IsWordGuessed, &state.Message)

		if state.IsWordGuessed {
			helpers.PrintGreen(lobby.String() + ": Player " + fmt.Sprintf("%d", state.Turn) + " guessed the word/sentence correctly, game over!")
			// Since turn is not yet incremented, the player who guessed the word is the current player, so we can send state.Turn
			msg := engine.FormatMessage(&ServerMessage{"game_over", state.Message, -1, state.Turn, state.PreviousGuesses, state.Word, state.Turn, false})
			engine.SendMulticastMessage(&lobby.Conns, msg)
			break
		}

		state.PreviousGuesses = append(state.PreviousGuesses, clientMsg.Guess)
		helpers.PrintInfo(lobby.String() + ": Player " + fmt.Sprintf("%d", state.Turn) + " guessed: " + clientMsg.Guess)
		state.Turn = (state.Turn + 1) % len(lobby.Conns)
	}

}

func (lobby *WordGuessLobby) SendTagAssignments() {
	for i, conn := range lobby.Conns {
		msg := &ServerMessage{
			"tag_assignment",
			fmt.Sprintf("You are player %d", i),
			i,
			0,
			[]string{},
			"",
			-1,
			false,
		}
		_, err := engine.SendUnicastMessage(&conn, engine.FormatMessage[*ServerMessage](msg))
		if err != nil {
			fmt.Println("Error sending message to player", err)
		}
	}
}

func (lobby *WordGuessLobby) finishGameOnNotEnoughPlayers(state *GameState) {
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

func (lobby *WordGuessLobby) IsGameOverOnError(state *GameState, err error) bool {
	if engine.IsTimeoutError(err) {
		helpers.PrintRed(lobby.String() + ": Client input timeout, SEND MESSAGE TO PLAYERS and continue")
		return false
	} else {
		helpers.PrintRed(lobby.String() + ": Player " + lobby.Conns[state.Turn].RemoteAddr().String() + " disconnected")
		engine.RemoveConn(&lobby.Conns[state.Turn], &lobby.Conns)

		if len(lobby.Conns) < MIN_CONN {
			lobby.finishGameOnNotEnoughPlayers(state)
			return true
		}
	}

	return false
}

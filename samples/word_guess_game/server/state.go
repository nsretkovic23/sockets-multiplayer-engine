package main

import (
	"fmt"
	"sockets-multiplayer/engine"
	"sockets-multiplayer/samples/word_guess_game/server/word"
	"time"
)

type GameState struct {
	Turn            int
	PreviousGuesses []string
	GuessedLetters  map[string]bool
	Word            string
	IsWordGuessed   bool
	Message         string
}

func (state *GameState) ListenForNextTurn(lobby *WordGuessLobby, msgRaw *[]byte) (int, error) {
	lobby.Conns[state.Turn].SetReadDeadline(time.Now().Add(TIMEOUT * time.Second))
	return lobby.Conns[state.Turn].Read(*msgRaw)
}

func CreateInitialState() *GameState {
	return &GameState{
		0,
		[]string{},
		word.InitializeLetterMap(),
		"NIKOLA SRETKOVIC",
		false,
		"",
	}
}

func (state *GameState) Initialize() {
	state.Turn = 0
	state.PreviousGuesses = []string{}
	state.GuessedLetters = word.InitializeLetterMap()
	state.Word = "NIKOLA SRETKOVIC"
	state.IsWordGuessed = false
	state.Message = ""
}

func (state *GameState) GetNextTurnMessage() []byte {
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

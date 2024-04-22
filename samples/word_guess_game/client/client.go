package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sockets-multiplayer/engine"
	"sockets-multiplayer/helpers"
	"strings"
	"time"
)

const (
	TIMEOUT = 15
	PORT    = 8080
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
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", PORT))
	if err != nil {
		helpers.PrintRed("Error connecting to server: " + err.Error())
		return
	}
	defer conn.Close()
	helpers.PrintInfo("Connected to server")

	var message ServerMessage
	var myTag int

	for {
		msgRaw := make([]byte, 2048)
		n, err := conn.Read(msgRaw)
		if err != nil {
			helpers.PrintRed("Error: " + err.Error())
			return
		}

		err = json.Unmarshal(msgRaw[:n], &message)
		if err != nil {
			helpers.PrintRed("Error unmarshalling message: " + err.Error())
			continue
		}

		if message.Type == "welcome" {
			handleWelcomeMessage(&message)
		} else if message.Type == "tag_assignment" {
			handleTagAssignmentMessage(&message, &myTag)
		} else if message.Type == "turn" {
			handleTurn(&message, myTag, &conn)
		} else if message.Type == "game_over" {
			handleGameOver(&message, myTag)
			break
		}
	}

}

func handleGameOver(msg *ServerMessage, myTag int) {
	if msg.WinnerTag == myTag || msg.WonByDisconnect {
		helpers.PrintGreen(fmt.Sprintf("Game over: %s The word/sentence was: %s", msg.Text, msg.GuessState))
	} else {
		helpers.PrintRed(fmt.Sprintf("Game over. Player %d guessed the word/sentence correctly. The word/sentence was: %s", msg.WinnerTag, msg.GuessState))
	}
}

func handleTurn(msg *ServerMessage, myTag int, serverConn *net.Conn) {
	if msg.Turn == myTag {
		ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT*time.Second)
		defer cancel()

		helpers.PrintGreen(fmt.Sprintf("It's my turn! Guesses: [ %v ]", strings.Join(msg.PreviousGuesses, ", ")))
		helpers.PrintYellow(msg.GuessState)

		guessChan := make(chan string, 1)
		scanner := bufio.NewScanner(os.Stdin)

		// TODO: LEAKING GOROUTINE PROBLEM
		go func() {
			if scanner.Scan() {
				guessChan <- scanner.Text()
			}
		}()

		var guess string
		select {
		case guess = <-guessChan:
			// Received a guess from the user
		case <-ctx.Done():
			// Timeout
			helpers.PrintRed("Timeout. No guess was entered.")
			return
		}

		msg := engine.FormatMessage(&ClientMessage{myTag, guess})
		// Send the guess to the server
		engine.SendUnicastMessage(serverConn, msg)
	} else {
		helpers.PrintInfo(fmt.Sprintf("It's %d's turn... Guesses: [ %v ]", msg.Turn, strings.Join(msg.PreviousGuesses, ", ")))
		helpers.PrintYellow(msg.GuessState)
	}
}

func handleWelcomeMessage(msg *ServerMessage) {
	fmt.Println("Connected to the server, server message:", msg.Text)
}

func handleTagAssignmentMessage(msg *ServerMessage, myTag *int) {
	fmt.Println("My tag is:", msg.Tag)
	*myTag = msg.Tag
}

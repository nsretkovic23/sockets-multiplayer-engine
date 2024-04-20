package main

import (
	"bufio"
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

	if msg.Turn == myTag {
		helpers.PrintGreen("Congratulations! You correctly guessed the word/sentence: !" + msg.GuessState)
	} else {
		helpers.PrintRed("Game over! Player " + fmt.Sprint(msg.Turn) + " guessed the sentence/word." + "The word/sentence was: " + msg.GuessState)
	}
}

func handleTurn(msg *ServerMessage, myTag int, serverConn *net.Conn) {
	if msg.Turn == myTag {
		helpers.PrintGreen(fmt.Sprintf("It's my turn! Previous guesses: [ %v ]", strings.Join(msg.PreviousGuesses, ", ")))
		helpers.PrintYellow(msg.GuessState + "\n")

		guessChan := make(chan string)
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				guessChan <- scanner.Text()
			}
		}()

		var guess string
		select {
		case guess = <-guessChan:
			// Received a guess from the user
		case <-time.After(TIMEOUT * time.Second):
			// Timeout
			helpers.PrintRed("Timeout. No guess was entered.")
			return
		}

		// Send the guess to the server
		msg := engine.FormatMessage(&ClientMessage{myTag, guess})
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

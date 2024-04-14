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
)

const (
	PORT = 8080
)

type ServerMessage struct {
	Type            string
	Text            string
	Tag             int
	Turn            int
	PreviousGuesses []string
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

	var lastMessage ServerMessage
	var myTag int

	for {
		msgRaw := make([]byte, 2048)
		n, err := conn.Read(msgRaw)
		if err != nil {
			helpers.PrintRed("Error: " + err.Error())
			return
		}

		err = json.Unmarshal(msgRaw[:n], &lastMessage)
		if err != nil {
			helpers.PrintRed("Error unmarshalling message: " + err.Error())
			continue
		}

		if lastMessage.Type == "welcome" {
			handleWelcomeMessage(&lastMessage)
		} else if lastMessage.Type == "tag_assignment" {
			handleTagAssignmentMessage(&lastMessage, &myTag)
		} else if lastMessage.Type == "turn" {
			handleTurn(&lastMessage, myTag, &conn)
		} else if lastMessage.Type == "game_over" {
			break
		}
	}

}

func handleTurn(msg *ServerMessage, myTag int, serverConn *net.Conn) {
	if msg.Turn == myTag {
		helpers.PrintGreen(fmt.Sprintf("It's my turn! Previous guesses: [ %v ]", strings.Join(msg.PreviousGuesses, ", ")))

		scanner := bufio.NewScanner(os.Stdin)
		guess := ""

		for scanner.Scan() {
			guess = scanner.Text()
			break
		}

		msg := engine.FormatMessage(&ClientMessage{myTag, guess})
		engine.SendUnicastMessage(serverConn, msg)
	} else {
		helpers.PrintInfo(fmt.Sprintf("It's %d's turn... Guesses: [ %v ]", msg.Turn, strings.Join(msg.PreviousGuesses, ", ")))
	}
}

func handleWelcomeMessage(msg *ServerMessage) {
	fmt.Println("Connected to the server, server message:", msg.Text)
}

func handleTagAssignmentMessage(msg *ServerMessage, myTag *int) {
	fmt.Println("My tag is:", msg.Tag)
	*myTag = msg.Tag
}

package main

import (
	"encoding/json"
	"fmt"
	"net"
)

const (
	PORT = 8080
)

type Message struct {
	Text string
}

func main() {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", PORT))
	if err != nil {
		fmt.Println("Error connecting to server:", err.Error())
		return
	}
	defer conn.Close()
	fmt.Println("Connected to server")

	for {
		msgRaw := make([]byte, 2048)
		n, err := conn.Read(msgRaw)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var msg Message
		err = json.Unmarshal(msgRaw[:n], &msg)
		if err != nil {
			fmt.Println("Error unmarshalling message:", err)
			continue
		}

		fmt.Println("Received message from server:")
		fmt.Println(msg.Text)
	}

	// conn.Write([]byte("Hello I'm here!"))
	// for {
	// 	msg := make([]byte, 1024)
	// 	_, err := conn.Read(msg)
	// 	if err != nil {
	// 		fmt.Println("Error:", err)
	// 		return
	// 	}
	// }

}

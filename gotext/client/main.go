package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type ChatRequest struct {
	Action   string
	Username string
	Message  string
}

func main() {
	if len(os.Args) < 5 {
		log.Fatal("Usage: ", os.Args[0], " port serverAddr username peername")
	}
	port := fmt.Sprintf(":%s", os.Args[1])
	serverAddr := os.Args[2]
	username := os.Args[3]
	peer := os.Args[4]
	buf := make([]byte, 2048)

	// Prepare to register user to server.
	saddr, err := net.ResolveUDPAddr("udp4", serverAddr)
	if err != nil {
		log.Print("Resolve server address failed.")
		log.Fatal(err)
	}

	// Prepare for local listening.
	addr, err := net.ResolveUDPAddr("udp4", port)
	if err != nil {
		log.Print("Resolve local address failed.")
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Print("Listen UDP failed.")
		log.Fatal(err)
	}

	// Send registration information to server.
	initChatRequest := ChatRequest{
		"New",
		username,
		"",
	}
	jsonRequest, err := json.Marshal(initChatRequest)
	if err != nil {
		log.Print("Marshal Register information failed.")
		log.Fatal(err)
	}
	_, err = conn.WriteToUDP(jsonRequest, saddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Waiting for server response...")
	_, _, err = conn.ReadFromUDP(buf)
	if err != nil {
		log.Print("Register to server failed.")
		log.Fatal(err)
	}

	// Send connect request to server
	connectChatRequest := ChatRequest{
		"Get",
		username,
		peer,
	}
	jsonRequest, err = json.Marshal(connectChatRequest)
	if err != nil {
		log.Print("Marshal connection information failed.")
		log.Fatal(err)
	}

	var serverResponse ChatRequest
	for i := 0; i < 3; i++ {
		conn.WriteToUDP(jsonRequest, saddr)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Print("Get peer address from server failed.")
			log.Fatal(err)
		}
		err = json.Unmarshal(buf[:n], &serverResponse)
		if err != nil {
			log.Print("Unmarshal server response failed.")
			log.Fatal(err)
		}
		if serverResponse.Message != "" {
			break
		}
		time.Sleep(10 * time.Second)
	}
	if serverResponse.Message == "" {
		log.Fatal("Cannot get peer's address")
	}
	log.Print("Peer address: ", serverResponse.Message)
	peerAddr, err := net.ResolveUDPAddr("udp4", serverResponse.Message)
	if err != nil {
		log.Print("Resolve peer address failed.")
		log.Fatal(err)
	}

	// Start chatting.
	write(conn, username, peerAddr)
	go listen(conn)

}

func write(conn *net.UDPConn, username string, peerAddr *net.UDPAddr) {
	for {
		fmt.Print("Input message: ")
		var message string
		fmt.Scanln(&message)
		if strings.Contains(message, "_exit0") {
			send(conn, username, "quitting...", peerAddr)
			os.Exit(0)
		}
		send(conn, username, message, peerAddr)
	}
}

func send(conn *net.UDPConn, username string, message string, peerAddr *net.UDPAddr) {
	messageRequest := ChatRequest{
		"Chat",
		username,
		message,
	}
	jsonRequest, err := json.Marshal(messageRequest)
	if err != nil {
		log.Print("Error: ", err)
	}
	_, _ = conn.WriteToUDP(jsonRequest, peerAddr)
}

func listen(conn *net.UDPConn) {
	for {
		buf := make([]byte, 4096)
		_, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Print(err)
			continue
		}
		// log.Print("Message from ", addr.IP)
		var message ChatRequest
		err = json.Unmarshal(buf, &message)
		if err != nil {
			log.Print(err)
			continue
		}
		if strings.Contains(message.Message, "quitting...") {
			fmt.Println("Exit command received")
			fmt.Println("Do you want to quit? (Y/N)")
			var confirmExit string
			_, _ = fmt.Scanln(&confirmExit)
			if confirmExit == "Y" {
				os.Exit(0)
			} else {
				continue
			}
		}
		fmt.Println(message.Username, ":", message.Message)
	}
}
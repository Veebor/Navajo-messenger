package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

var username string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) < 5 {
		log.Fatal("Usage: ", os.Args[0], " port serverAddr username peername")
	}
	portLoc := fmt.Sprintf(":%s", os.Args[1])
	serverAddr := os.Args[2]
	username = os.Args[3]
	peer := os.Args[4]

	conn := startServerConn(serverAddr, portLoc)

	SetupCloseHandler(conn)

	registerPeer(conn)

	peerAddress := getPeerAddr(conn, peer)

	// Start chatting.
	fmt.Println("Connected to", peer)

	go func() {
		for {
			fmt.Print("> ")
			reader := bufio.NewReader(os.Stdin)
			message, _ := reader.ReadString('\n')
			message = strings.TrimRight(message, "\r\n")
			Write(conn, message, peerAddress)
		}
	}()
	go func() {
		for {
			remUser, remMessage, remMessageOK := Receive(conn)
			if remMessageOK {
				fmt.Println(remUser, " said: ", remMessage)
			} else {
				fmt.Println("Bad message")
			}
			if remMessage == "_quitting...0" {
				cleanQuit(conn)
			}
		}
	}()
	select {}
}

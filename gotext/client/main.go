package main

import (
	"fmt"
	"log"
	"os"
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

	registerPeer(conn)

	SetupCloseHandler(conn)

	peerAddress := getPeerAddr(conn, peer)

	// Start chatting.
	fmt.Println("Connected to", peer)
	go func() {
		for {
			message := make([]byte, 4096)
			fmt.Print("> ")
			fmt.Scanln(&message)
			write(conn, string(message), peerAddress)
		}
	}()
	go func() {
		for {
			remUser, remMessage := listen(conn)
			fmt.Println(remUser, "said: ", remMessage)
			if remMessage == "_quitting...0" {
				cleanQuit(conn)
			}
			fmt.Print("> ")
		}
	}()
	select {}
}
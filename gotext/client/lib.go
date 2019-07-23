package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type ChatRequest struct {
	Action   string
	Username string
	Message  string
}

var saddr *net.UDPAddr

func HashStr(Txt string) string {
	h := sha1.New()
	h.Write([]byte(Txt))
	bs := h.Sum(nil)
	sh := string(fmt.Sprintf("%x\n", bs))
	return sh
}

func startServerConn(serverAddress string, localport string) *net.UDPConn {
	saddr, _ = net.ResolveUDPAddr("udp4", serverAddress)
	// Prepare for local listening.
	addr, err := net.ResolveUDPAddr("udp4", localport)
	if err != nil {
		log.Print("Resolve local address failed.")
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Print("Listen UDP failed.")
		log.Fatal(err)
	}
	return conn
}

func registerPeer(conn *net.UDPConn) {
	buf := make([]byte, 2048)
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
}

func getPeerAddr(conn *net.UDPConn, peer string) *net.UDPAddr {
	buf := make([]byte, 2048)
	connectChatRequest := ChatRequest{
		"Get",
		username,
		peer,
	}
	jsonRequest, err := json.Marshal(connectChatRequest)
	if err != nil {
		log.Print("Marshal connection information failed.")
		log.Fatal(err)
	}
	var serverResponse ChatRequest
	for {
		_, err = conn.WriteToUDP(jsonRequest, saddr)
		if err != nil {
			log.Println("WriteToServer Error", err)
		}
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
		if strings.Contains(serverResponse.Message, "IDINUSE") {
			fmt.Println("Peer ID already in use")
			os.Exit(0)
		} else if serverResponse.Message != "" {
			break
		}
		time.Sleep(2 * time.Second)
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
	return peerAddr
}

func SetupCloseHandler(conn *net.UDPConn) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		cleanQuit(conn)
	}()
}

func cleanQuit(conn *net.UDPConn) {
	fmt.Print("Quitting...")
	for i := 0; i < 3; i++ {
		send(conn, "RM", username, saddr)
		time.Sleep(1 * time.Second)
	}
	os.Exit(0)
}

func Write(conn *net.UDPConn, message string, peerAddr *net.UDPAddr) {
	message = strings.TrimSpace(message)
	if strings.Contains(message, "_~_") {
		log.Print("Invalid message. _~_ not allowed")
	}
	if strings.Contains(message, "_exit0") {
		send(conn, "Chat", "_quitting...0", peerAddr)
		cleanQuit(conn)
	} else {
		send(conn, "Chat", message+"_~_"+HashStr(message), peerAddr)
	}
}

func send(conn *net.UDPConn, action string, message string, peerAddr *net.UDPAddr) {
	messageRequest := ChatRequest{
		action,
		username,
		message,
	}
	jsonRequest, err := json.Marshal(messageRequest)
	if err != nil {
		log.Print("Error: ", err)
	}
	_, err = conn.WriteToUDP(jsonRequest, peerAddr)
	if err != nil {
		log.Print("Error: ", err)
	}
}

func listen(conn *net.UDPConn) (string, string) {
	buf := make([]byte, 4096)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.Print(err)
	}
	// log.Print("Message from ", addr.IP)
	var message ChatRequest
	err = json.Unmarshal(buf[:n], &message)
	if err != nil {
		log.Print(err)
	}
	return message.Username, message.Message
}

func Receive(conn *net.UDPConn) (string, string, bool) {
	var messageOK bool
	username, message := listen(conn)
	messageHash := strings.Split(message, "_~_")
	if HashStr(messageHash[0]) == messageHash[1] {
		messageOK = true
	} else {
		messageOK = false
	}
	return username, messageHash[0], messageOK
}

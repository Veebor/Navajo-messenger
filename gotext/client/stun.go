package main

import (
	"fmt"
	"log"
	"net"

	"gortc.io/stun"
)

var (
	myPublicAddress string
	STUNserver      = "stun.l.google.com:19302"
	stunAddress     string
)

const (
	udp           = "udp4"
	timeoutMillis = 500
)

func sendBindingRequest(conn *net.UDPConn, addr *net.UDPAddr) error {
	m := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	_, err := conn.WriteToUDP(m.Raw, addr)
	if err != nil {
		return fmt.Errorf("binding: %v", err)
	}

	return nil
}

func getAddrStun(conn *net.UDPConn) string {
	srvAddr, err := net.ResolveUDPAddr(udp, STUNserver)

	if err != nil {
		log.Fatal("ERR:", err)
	}

	//keepalive := time.Tick(timeoutMillis * time.Millisecond)
	// Creating a "connection" to STUN server.
	err = sendBindingRequest(conn, srvAddr)

	for {
		buf := make([]byte, 4096)
		n, _, _ := conn.ReadFromUDP(buf)
		buf = buf[:n]
		if stun.IsMessage(buf) {
			m := new(stun.Message)
			m.Raw = buf
			decErr := m.Decode()
			if decErr != nil {
				log.Println("decode:", decErr)
				break
			}
			var xorAddr stun.XORMappedAddress
			if getErr := xorAddr.GetFrom(m); getErr != nil {
				log.Println("getFrom:", getErr)
				break
			}
			stunAddress = xorAddr.String()
			break
		}
	}
	return stunAddress
}

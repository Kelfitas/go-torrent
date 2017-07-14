package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
)

func handleError(e error) {
	if e != nil {
		panic(e)
	}
}

func prettyPrint(dict interface{}) {
	b, err := json.MarshalIndent(dict, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))
}

func getMapString(m map[string]interface{}, k string) string {
	if v, ok := m[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func hashBytes(b []byte) []byte {
	hasher := sha1.New()
	hasher.Write(b)

	return hasher.Sum(nil)
}

func hashString(b string) []byte {
	return hashBytes([]byte(b))
}

const (
	minTCPPort = 0
	maxTCPPort = 65535
)

func isTCPPortAvailable(port int) bool {
	if port < minTCPPort || port > maxTCPPort {
		return false
	}

	conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	defer conn.Close()
	if err != nil {
		return false
	}

	return true
}

func getListenPort() (port int, err error) {
	start := 6881
	end := 6889

	for port = start; port <= end; port++ {
		if isTCPPortAvailable(port) {
			return
		}
	}

	err = errors.New("no listening port available")
	return
}

func getNetString() string {
	var netString string

	ifaces, err := net.Interfaces()
	handleError(err)

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		handleError(err)
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			netString += ip.String()
		}
	}

	return netString
}

func getPeerID() (peerID string) {
	hostName, err := os.Hostname()
	handleError(err)

	netString := getNetString()

	peerID = hostName + ":"
	peerID += netString + ":"

	peerID = hex.EncodeToString(hashString(peerID))
	peerID = peerID[:20]

	return
}

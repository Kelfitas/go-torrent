package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net"
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

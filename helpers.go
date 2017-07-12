package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
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

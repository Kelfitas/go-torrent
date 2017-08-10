package main

import (
	"bufio"
	"fmt"
	"net"
)

func (p *Peer) connect() (err error) {
	var addr string

	if p.Port != 0 {
		addr = fmt.Sprintf("%s:%d", p.IP, p.Port)
	} else {
		addr = p.IP
	}

	fmt.Printf("\nConnecting to: %s...\n", addr)

	p.conn, err = net.Dial("tcp", addr)
	if err != nil {
		return
	}

	caca, err := p.readHeader()
	handleError(err)
	fmt.Println(caca)

	fmt.Println("Reading...")
	// fmt.Fprintf(p.conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(p.conn).ReadString('\n')
	if err != nil {
		return
	}
	fmt.Println("Finished reading!")

	fmt.Printf("Status: %s\n", status)

	return
}

func (p *Peer) readHeader() (b []byte, err error) {
	header := make([]byte, 68)

	_, err = p.conn.Read(header[0:1])
	if err != nil {
		err = fmt.Errorf("Couldn't read 1st byte: %v", err)
		return
	}

	if header[0] != 19 {
		err = fmt.Errorf("First byte is not 19")
		return
	}

	_, err = p.conn.Read(header[1:20])
	if err != nil {
		err = fmt.Errorf("Couldn't read magic string: %v", err)
		return
	}

	if string(header[1:20]) != "BitTorrent protocol" {
		err = fmt.Errorf("Magic string is not correct: %v", string(header[1:20]))
		return
	}

	// Read rest of header
	_, err = p.conn.Read(header[20:])
	if err != nil {
		err = fmt.Errorf("Couldn't read rest of header")
		return
	}

	b = make([]byte, 48)
	copy(b, header[20:])

	return
}

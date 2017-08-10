package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
)

const userAgent = "uTorrentMac/1870(42417)"

const (
	peerChoke         = iota // 0 - choke
	peerUnchoke              // 1 - unchoke
	peerInterested           // 2 - interested
	peerNotInterested        // 3 - not interested
	peerHave                 // 4 - have
	peerBitfield             // 5 - bitfield
	peerRequest              // 6 - request
	peerPiece                // 7 - piece
	peerCancel               // 8 - cancel
)

func (t *Torrent) parseTorrentFile(filename string) (err error) {
	t.File = filename

	f, err := os.Open(filename)
	handleError(err)
	defer f.Close()

	reader := bufio.NewReader(f)
	info, err := bencode.Decode(reader)
	handleError(err)

	topMap, ok := info.(map[string]interface{})
	if !ok {
		err = errors.New("couldn't parse torrent file")
		return
	}

	infoMap, ok := topMap["info"]
	if !ok {
		err = errors.New("no info dict")
		return
	}

	var b bytes.Buffer
	if err = bencode.Marshal(&b, infoMap); err != nil {
		return
	}

	t.Meta.InfoHash = string(hashBytes(b.Bytes()))

	err = bencode.Unmarshal(&b, &t.Meta.Info)
	if err != nil {
		return
	}

	t.Meta.Announce = getMapString(topMap, "announce")

	return
}

func (t *Torrent) parseCompactResponse() {
	// The first 4 bytes contain the 32-bit ipv4 address. The remaining two bytes contain the port number.
	var ip string
	var port int
	for i, b := range t.Response.Body {
		byteNo := i % 6

		if byteNo == 0 && i > 0 {
			t.Response.Peers = append(t.Response.Peers, Peer{
				IP:   ip,
				Port: port,
			})

			ip = ""
			port = 0
		}

		intB := int(b)
		if byteNo < 4 {
			ip += strconv.Itoa(intB)

			if byteNo < 3 {
				ip += "."
			}
		} else if intB > 0 {
			if port == 0 {
				port = intB
			} else {
				port *= intB
			}
		}
	}
}

func (t *Torrent) getInfoHash() string {
	return url.QueryEscape(t.Meta.InfoHash)
}

func (t *Torrent) buildAnnounceURL(event string) (url *url.URL) {
	listenPort, err := getListenPort()
	handleError(err)

	peerID := getPeerID()

	u := t.Meta.Announce
	u += "?info_hash=" + t.getInfoHash()
	u += "&peer_id=" + peerID
	u += "&port=" + strconv.Itoa(listenPort)
	u += "&uploaded=" + strconv.Itoa(t.Stats.Uploaded)
	u += "&downloaded=" + strconv.Itoa(t.Stats.Downloaded)
	u += "&left=" + strconv.Itoa(t.Stats.Left)
	u += "&corrupt=" + strconv.Itoa(t.Stats.Corrupt)
	u += "&event=" + event

	if compactResponse {
		u += "&compact=1"
	} else {
		u += "&compact=0"
	}

	url, err = url.Parse(u)
	handleError(err)

	return
}

func (t *Torrent) announce(event string) (resp *http.Response) {
	url := t.buildAnnounceURL(event)

	client := &http.Client{}
	fmt.Printf("GET: %s\n", url.String())
	req, err := http.NewRequest("GET", url.String(), nil)
	handleError(err)

	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Connection", "close")

	resp, err = client.Do(req)
	handleError(err)

	return
}

// AnnounceStart announces the starting of the download phase
func (t *Torrent) AnnounceStart() {
	resp := t.announce(eventStarted)
	defer resp.Body.Close()

	var err error
	t.Response.Body, err = ioutil.ReadAll(resp.Body)
	handleError(err)

	if compactResponse {
		t.parseCompactResponse()
		return
	}

	var b bytes.Buffer
	b.Write(t.Response.Body)

	err = bencode.Unmarshal(&b, &t.Response)
	handleError(err)
}

// AnnounceStop announces the stop of the download phase
func (t *Torrent) AnnounceStop() {
	resp := t.announce(eventStopped)
	defer resp.Body.Close()

	var err error
	t.Response.Body, err = ioutil.ReadAll(resp.Body)
	handleError(err)

	prettyPrint(t.Response.Body)
}

// AnnounceComplete announces the completion of the download phase
func (t *Torrent) AnnounceComplete() {
	resp := t.announce(eventCompleted)
	defer resp.Body.Close()

	var err error
	t.Response.Body, err = ioutil.ReadAll(resp.Body)
	handleError(err)

	prettyPrint(t.Response.Body)
}

func (t *Torrent) Listen() {
	listenPort, err := getListenPort()
	handleError(err)

	fmt.Printf("Listening on: %d", listenPort)

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(listenPort))
	handleError(err)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Err: %s", err.Error())
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Printf("Connection IP: %s", conn.RemoteAddr().String())
	status, err := bufio.NewReader(conn).ReadString('\n')
	handleError(err)
	fmt.Printf("Status: %s", status)
	// fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
}

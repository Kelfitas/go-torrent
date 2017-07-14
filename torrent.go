package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
)

const userAgent = "uTorrentMac/1870(42417)"

type FileDict struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"path"`
	Md5sum string
}

type InfoDict struct {
	PieceLength int64 `bencode:"piece length"`
	Pieces      string
	Private     int64
	Name        string
	// Single File Mode
	Length int64
	Md5sum string
	// Multiple File mode
	Files        []FileDict
	FileDuration []int64
	FileMedia    []string
}

type MetaInfo struct {
	Info         InfoDict
	InfoHash     string
	Announce     string
	AnnounceList [][]string `bencode:"announce-list"`
	CreationDate int64      `bencode:"creation date"`
	Comment      string
	CreatedBy    string `bencode:"created by"`
	Encoding     string
}

type PeerDict struct {
	PeerID string `bencode:"peer id"`
	IP     string `bencode:"ip"`
	Port   int    `bencode:"port"`
}

type AnnounceResponse struct {
	FailureReason string     `bencode:"failure reason"`
	MinInterval   int64      `bencode:"min interval"`
	Interval      int64      `bencode:"interval"`
	Peers         []PeerDict `bencode:"peers"`
	Body          []byte
}

type Stats struct {
	Uploaded   int
	Downloaded int
	Left       int
	Corrupt    int
}

type Torrent struct {
	Meta     MetaInfo
	Response AnnounceResponse
	File     string
	Stats    Stats
}

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
			t.Response.Peers = append(t.Response.Peers, PeerDict{
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

	u := t.Meta.Announce
	u += "?info_hash=" + t.getInfoHash()
	u += "&peer_id=" + getPeerID()
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

func (t *Torrent) AnnounceStart() {
	url := t.buildAnnounceURL(eventStarted)

	client := &http.Client{}
	fmt.Printf("GET: %s\n", url.String())
	req, err := http.NewRequest("GET", url.String(), nil)
	handleError(err)

	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Connection", "close")

	resp, err := client.Do(req)
	handleError(err)

	defer resp.Body.Close()

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

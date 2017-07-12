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

	"encoding/hex"

	bencode "github.com/jackpal/bencode-go"
)

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

type AnnounceResponse struct {
	FailureReason string `bencode:"failure reason"`
	MinInterval   int64  `bencode:"min interval"`
	Interval      int64  `bencode:"interval"`
	Message       []byte
}

type Torrent struct {
	Meta     MetaInfo
	Response AnnounceResponse
	File     string
}

var torrent Torrent

func parseTorrentFile(filename string, t *Torrent) (err error) {
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

func getInfoHash(t *Torrent) string {
	return url.QueryEscape(t.Meta.InfoHash)
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

	err = errors.New("no Listening port available")
	return
}

func buildAnnounceURL(t *Torrent) (url *url.URL) {
	var u string

	// fmt.Printf("t.Meta.Info: %s\n", hex.EncodeToString([]byte(t.Meta.InfoHash)))
	// fmt.Println("uTorrent:    647dc25f88db008dec98c4d60ba68c71ea07162b")

	listenPort, err := getListenPort()
	handleError(err)

	u = t.Meta.Announce
	u += "?info_hash=" + getInfoHash(t)
	u += "&peer_id=" + getPeerID()
	u += "&port=" + strconv.Itoa(listenPort)
	u += "&uploaded=0"
	u += "&downloaded=0"
	u += "&left=0"
	u += "&corrupt=0"
	// u += "&key=02146AFD"
	u += "&event=started"
	// u += "&numwant=200"
	// u += "&no_peer_id=1"
	u += "&compact=1"
	// u += "&ipv6=fe80%3a%3ac9b%3a501d%3a61bb%3a4d4c"

	url, err = url.Parse(u)
	handleError(err)

	return
}

func Announce(t *Torrent) {
	url := buildAnnounceURL(t)

	client := &http.Client{}
	fmt.Printf("GET: %s\n", url.String())
	req, err := http.NewRequest("GET", url.String(), nil)
	handleError(err)

	req.Header.Add("User-Agent", "uTorrentMac/1870(42417)")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Connection", "close")

	resp, err := client.Do(req)
	handleError(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var b bytes.Buffer
	b.Write(body)

	err = bencode.Unmarshal(&b, &t.Response)
	if err != nil {
		t.Response.Message = body
	}
}

func main() {
	sampleTorrent := "/Users/cioatamihai/Downloads/Sex.Drive.Unrated.2008.1080p.BluRay.x264.AC3.RoSubbed-HDChina.torrent"
	err := parseTorrentFile(sampleTorrent, &torrent)
	handleError(err)

	Announce(&torrent)

	prettyPrint(torrent.Response)
}

// announce — the URL of the tracker
// info — this maps to a dictionary whose keys are dependent on whether one or more files are being shared:
//     files — a list of dictionaries each corresponding to a file (only when multiple files are being shared). Each dictionary has the following keys:
//     length — size of the file in bytes.
//     path — a list of strings corresponding to subdirectory names, the last of which is the actual file name
//     length — size of the file in bytes (only when one file is being shared)
//     name — suggested filename where the file is to be saved (if one file)/suggested directory name where the files are to be saved (if multiple files)
//     piece length — number of bytes per piece. This is commonly 28 KiB = 256 KiB = 262,144 B.
//     pieces — a hash list, i.e., a concatenation of each piece's SHA-1 hash. As SHA-1 returns a 160-bit hash, pieces will be a string whose length is a multiple of 160-bits. If the torrent contains multiple files, the pieces are formed by concatenating the files in the order they appear in the files dictionary (i.e. all pieces in the torrent are the full piece length except for the last piece, which may be shorter).

// trackers
//     Tracker GET requests have the following keys:

// info_hash
//     The 20 byte sha1 hash of the bencoded form of the info value from the metainfo file. This value will almost certainly have to be escaped.

// Note that this is a substring of the metainfo file. The info-hash must be the hash of the encoded form as found in the .torrent file, which is identical to bdecoding the metainfo file, extracting the info dictionary and encoding it if and only if the bdecoder fully validated the input (e.g. key ordering, absence of leading zeros). Conversely that means clients must either reject invalid metainfo files or extract the substring directly. They must not perform a decode-encode roundtrip on invalid data.

// peer_id
//     A string of length 20 which this downloader uses as its id. Each downloader generates its own id at random at the start of a new download. This value will also almost certainly have to be escaped.
// ip
//     An optional parameter giving the IP (or dns name) which this peer is at. Generally used for the origin if it's on the same machine as the tracker.
// port
//     The port number this peer is listening on. Common behavior is for a downloader to try to listen on port 6881 and if that port is taken try 6882, then 6883, etc. and give up after 6889.
// uploaded
//     The total amount uploaded so far, encoded in base ten ascii.
// downloaded
//     The total amount downloaded so far, encoded in base ten ascii.
// left
//     The number of bytes this peer still has to download, encoded in base ten ascii. Note that this can't be computed from downloaded and the file length since it might be a resume, and there's a chance that some of the downloaded data failed an integrity check and had to be re-downloaded.
// event
//     This is an optional key which maps to started, completed, or stopped (or empty, which is the same as not being present). If not present, this is one of the announcements done at regular intervals. An announcement using started is sent when a download first begins, and one using completed is sent when the download is complete. No completed is sent if the file was complete when started. Downloaders send an announcement using stopped when they cease downloading.

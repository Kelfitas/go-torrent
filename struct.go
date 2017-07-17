package main

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
	PeerID  string `bencode:"peer id"`
	IP      string `bencode:"ip"`
	Port    int    `bencode:"port"`
	message uint8
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

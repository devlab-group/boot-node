package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Peer struct {
	Addr      string `json:"addr,omitempty"`
	PubKey    string `json:"publicKey,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

var peersList map[string]Peer
var isStarted = false

func main() {
	peersList = make(map[string]Peer)

	http.HandleFunc("/peers", handlePeers)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlePeers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getPeersHandler(w, r)
	case "POST":
		addPeerHandler(w, r)
	}
}

func addPeerHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var peerData map[string]string
	err := decoder.Decode(&peerData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
		return
	}

	address := r.RemoteAddr
	pubKey := peerData["publicKey"]

	if len(pubKey) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
		return
	}

	peersList[pubKey] = Peer{
		Addr:      address,
		PubKey:    pubKey,
		Timestamp: time.Now().Unix(),
	}

	if isStarted {
		return
	}

	isStarted = true

	go func() {
		for {
			for k, peer := range peersList {
				if peer.Timestamp+15 < time.Now().Unix() {
					delete(peersList, k)
				}
			}

			time.Sleep(time.Second)
		}
	}()
}

func getPeersHandler(w http.ResponseWriter, r *http.Request) {
	var result []Peer

	for _, peer := range peersList {
		result = append(result, peer)

		if len(result) == 20 {
			break
		}
	}

	peersJson, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
		return
	}
	w.Write(peersJson)
}

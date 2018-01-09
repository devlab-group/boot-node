package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var networkIDs = map[string]int{
	"MAINNET": 1,
	"TESTNET": 2,
}

type Peer struct {
	Addr      string `json:"addr,omitempty"`
	PubKey    string `json:"publicKey,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

var peersList map[string]map[string]Peer
var isStarted = false

// One day
const peerExpireTime = 24 * 60 * 60

func main() {
	// Init empty peers maps for all possible networks
	peersList = make(map[string]map[string]Peer)
	for netId, _ := range networkIDs {
		peersList[netId] = make(map[string]Peer)
	}

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
	var peerData map[string]string
	decodeRequestBody(w, r, &peerData)

	netId := peerData["networkID"]
	address := peerData["address"]
	pubKey := peerData["publicKey"]

	if _, ok := networkIDs[netId]; !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unknown network ID\n"))
		return
	}

	if len(address) == 0 || len(pubKey) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
		return
	}

	peersList[netId][pubKey] = Peer{
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
			for _, peers := range peersList {
				for k, peer := range peers {
					if peer.Timestamp+peerExpireTime < time.Now().Unix() {
						delete(peers, k)
					}
				}
			}

			time.Sleep(time.Second)
		}
	}()
}

func getPeersHandler(w http.ResponseWriter, r *http.Request) {
	var peerData map[string]string
	decodeRequestBody(w, r, &peerData)

	netId := peerData["networkID"]
	if _, ok := networkIDs[netId]; !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unknown network ID\n"))
		return
	}

	var result []Peer
	for _, peer := range peersList[netId] {
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

func decodeRequestBody(w http.ResponseWriter, r *http.Request, output *map[string]string) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(output)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
		return
	}
	defer r.Body.Close()
}

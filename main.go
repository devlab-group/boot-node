package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
)

type arrayFlags []string

func (flags *arrayFlags) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

func (flags *arrayFlags) String() string {
	var buf bytes.Buffer
	for _, arg := range *flags {
		buf.WriteString(arg)
	}
	return buf.String()
}

type Peer struct {
	Addr      string `json:"addr,omitempty"`
	PubKey    string `json:"publicKey,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

var peersList map[string]map[string]Peer
var isStarted = false
var networks arrayFlags

// One day
const peerExpireTime = 24 * 60 * 60

func main() {
	peersList = make(map[string]map[string]Peer)

	flag.Var(&networks, "net", "Networks supported by this boot node")
	flag.Parse()

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

	peerNetwork := peerData["network"]
	address := peerData["address"]
	pubKey := peerData["publicKey"]

	if ok := isNetworkSupported(peerNetwork); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported network"))
		return
	}

	if len(address) == 0 || len(pubKey) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missed peer info"))
		return
	}

	if _, ok := peersList[peerNetwork]; !ok {
		peersList[peerNetwork] = make(map[string]Peer)
	}

	peersList[peerNetwork][pubKey] = Peer{
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
	netId := r.FormValue("network")

	if ok := isNetworkSupported(netId); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported network"))
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
		return
	}
	w.Write(peersJson)
}

func decodeRequestBody(w http.ResponseWriter, r *http.Request, output *map[string]string) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(output)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
}

func isNetworkSupported(network string) bool {
	if len(networks) == 0 {
		return true
	}

	for _, net := range networks {
		if network == net {
			return true
		}
	}
	return false
}

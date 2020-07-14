package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pborman/getopt"
)

var isActive bool
var isReplicating bool
var ts int64
var lastFetchedVersion int
var runElection bool

type replicationSet struct {
	ResultSet   string `json:"resultSet"`
	LastVersion string `json:"lastVersion"`
}

type electActive struct {
	Ts int64 `json:"ts"`
}

func replicationSetHandler(w http.ResponseWriter, req *http.Request) {
	version := req.URL.Query().Get("version")
	if version == "" {
		version = "0"
	}
	log.Println("Request from " + req.RemoteAddr +
		" for /replication-set with last requested version " + version)
	if req.Method != "GET" {
		http.Error(w, req.Method+" method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rSet := replicationSet{ResultSet: "", LastVersion: version}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rSet)
}

func electActiveHandler(w http.ResponseWriter, req *http.Request) {
	var elActive electActive
	err := json.NewDecoder(req.Body).Decode(&elActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if elActive.Ts > ts {
		// Peer has started later than this instance
		isActive = true
		isReplicating = false
		runElection = true
		log.Println("Server becomes active / remains active")
	} else {
		// Remain as standby, start replication
		isReplicating = true
		isActive = false
		log.Println("Server remains standby, starts replication from active")
	}

	w.WriteHeader(http.StatusNoContent)
}

func readyHandler(w http.ResponseWriter, req *http.Request) {
	if isActive == true {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Not active", http.StatusInternalServerError)
	}
}

func main() {
	isActive = false
	isReplicating = false
	ts = time.Now().UnixNano()
	lastFetchedVersion = 0
	runElection = true
	var (
		peerAddr = getopt.StringLong(
			"peerAddr", 'a', "", "IP address or FQDN of the peer instance. Mandatory")
		port = getopt.StringLong(
			"port", 'p', "8090", "Server Port. Peer instance must run on this port. Optional")
		help = getopt.BoolLong("help", 'h', "Help")
	)
	getopt.Parse()
	if *help {
		getopt.Usage()
		os.Exit(0)
	}

	if *peerAddr == "" {
		log.Println("A valid IP address or FQDN of peer instance is mandatory")
		os.Exit(0)
	} else {
		log.Println(
			"Starting server, running as standby (non-replicating). Peer instance address " +
				*peerAddr + ".")
		log.Println("ts value =", ts)
	}

	signalch := make(chan os.Signal, 1)
	signal.Notify(signalch, os.Interrupt, syscall.SIGTERM)
	tickerReplication := time.NewTicker(10000 * time.Millisecond)
	tickerElection := time.NewTicker(1000 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-tickerElection.C:
				if runElection == true {
					peerURL := "http://" + *peerAddr + ":" + *port + "/elect-active"
					elActive := electActive{Ts: ts}
					jsonValue, _ := json.Marshal(elActive)
					res, err := http.Post(peerURL, "application/json", bytes.NewBuffer(jsonValue))
					if err == nil {
						res.Body.Close()
						if isActive == false && isReplicating == false {
							// No communication from peer yet, continue to advertise candidature
						} else {
							runElection = false
						}
					} else {
						// Probably peer is not alive yet
					}
				}
			case <-tickerReplication.C:
				if isActive == false && isReplicating == true {
					peerURL := "http://" + *peerAddr + ":" + *port + "/replication-set?version=" +
						strconv.Itoa(lastFetchedVersion)
					client := http.Client{
						Timeout: 2 * time.Second,
					}
					res, err := client.Get(peerURL)
					if err != nil {
						// Peer is down or link failure, time to become active and start advertising candidature
						log.Println("Replicaion error: " + err.Error())
						isActive = true
						isReplicating = false
						runElection = true
						log.Println("Server transitions to active from standby")
					} else {
						// Do nothing with replication set
						res.Body.Close()
					}
					lastFetchedVersion += rand.Intn(100)
				}
			}
		}
	}()

	http.HandleFunc("/replication-set", replicationSetHandler)
	http.HandleFunc("/elect-active", electActiveHandler)
	// Following API serves as the readiness probe
	// https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
	// An active instance returns 2xx status, while a standby instance returns 5xx status
	http.HandleFunc("/ready", readyHandler)

	go func() {
		err := http.ListenAndServe(":"+*port, nil)
		if err != http.ErrServerClosed {
		}
	}()
	<-signalch
	log.Println("Stopping server")
	tickerReplication.Stop()
	tickerElection.Stop()
	done <- true
}

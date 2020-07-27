package main

import (
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
var ts int64
var lastFetchedVersion int

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

func readyHandler(w http.ResponseWriter, req *http.Request) {
	if isActive == true {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Not active", http.StatusInternalServerError)
	}
}

func main() {
	isActive = false
	ts = time.Now().UnixNano()
	lastFetchedVersion = 0
	var (
		peerAddr = getopt.StringLong(
			"peerAddr", 'a', "", "IP address or FQDN of the peer instance. Mandatory")
		port = getopt.StringLong(
			"port", 'p', "8090", "Server Port. Peer instance must run on this port. Optional")
		active = getopt.BoolLong(
			"active", 'c', "If set, the instance will be active instance in the cluster")
		help = getopt.BoolLong("help", 'h', "Help")
	)
	getopt.Parse()
	if *help {
		getopt.Usage()
		os.Exit(0)
	}

	if *active {
		isActive = true
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
	tickerReplication := time.NewTicker(1000 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-tickerReplication.C:
				if isActive == false {
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
	done <- true
}

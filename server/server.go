package main

import (
	"fmt"
	"github.com/pborman/getopt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func hello(w http.ResponseWriter, req *http.Request) {
	log.Println("Request from " + req.RemoteAddr + " for /hello")
	fmt.Fprintf(w, "hello world\n")
}

func main() {
	var (
		peerAddr = getopt.StringLong("peerAddr", 'a', "", "IP address or FQDN of the peer instance")
		peerPort = getopt.StringLong("peerPort", 'p', "", "Port of the peer instance")
		help     = getopt.BoolLong("help", 'h', "Help")
	)
	getopt.Parse()
	if *help {
		getopt.Usage()
		os.Exit(0)
	}

	log.Println("Starting server at " + time.Now().String())
	signalch := make(chan os.Signal, 1)
	signal.Notify(signalch, os.Interrupt, syscall.SIGTERM)
	ticker := time.NewTicker(10000 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				peerURL := "http://" + *peerAddr + ":" + *peerPort + "/hello"
				res, err := http.Get(peerURL)
				if err == nil {
					res.Body.Close()
				}
			}
		}
	}()
	http.HandleFunc("/hello", hello)
	go func() {
		err := http.ListenAndServe(":8090", nil)
		if err != http.ErrServerClosed {
		}
	}()
	<-signalch
	log.Println("Stopping server")
	ticker.Stop()
	done <- true
}

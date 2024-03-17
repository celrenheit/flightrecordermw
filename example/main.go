package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/celrenheit/flightrecordermw"
	"golang.org/x/exp/trace"
)

func main() {
	if err := main2(); err != nil {
		log.Fatal(err)
	}
}

func main2() error {
	fr := trace.NewFlightRecorder()
	fr.Start()

	frmw, err := flightrecordermw.New("/tmp", fr, func(r *http.Request, stats flightrecordermw.Stats) bool {
		return stats.Elapsed > 100*time.Millisecond
	})
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	mux.Handle("/traces/", http.StripPrefix("/traces/", frmw))

	mux.Handle("/hello", frmw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dur := rand.N(150 * time.Millisecond)
		fmt.Println("sleeping for", dur)

		time.Sleep(dur)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})))

	return http.ListenAndServe("127.0.0.1:3000", mux)
}

package flightrecordermw

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/exp/trace"
)

// Stats contains the statistics of a request.
type Stats struct {
	StatusCode int
	Elapsed    time.Duration
}

// ShouldRecord is a function that returns true if the request should be recorded.
type ShouldRecord func(r *http.Request, stats Stats) bool

// New returns a new FlightRecorder middleware and handler.
func New(dstDir string, fr *trace.FlightRecorder, shouldRecord ShouldRecord) (*frmw, error) {
	dir, err := os.MkdirTemp(dstDir, "flightrecorder-")
	if err != nil {
		return nil, err
	}

	return &frmw{
		dir:          dir,
		fr:           fr,
		fserver:      http.FileServerFS(os.DirFS(dir)),
		shouldRecord: shouldRecord,
	}, nil
}

type frmw struct {
	dir          string
	fr           *trace.FlightRecorder
	fserver      http.Handler
	shouldRecord ShouldRecord
}

// ServeHTTP implements http.Handler.
func (f *frmw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.fserver.ServeHTTP(w, r)
}

// Middleware returns a new http.Handler that records traces based on the provided condition.
func (f *frmw) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &rwWrapper{ResponseWriter: w}

		start := time.Now()
		h.ServeHTTP(w, r)

		stats := Stats{
			StatusCode: rec.statusCode,
			Elapsed:    time.Since(start),
		}

		// if elapsed > 500*time.Millisecond {
		if f.shouldRecord(r, stats) {
			var b bytes.Buffer
			_, err := f.fr.WriteTo(&b)
			if err != nil {
				log.Print(err)
				return
			}

			f, err := os.CreateTemp(f.dir, fmt.Sprintf("trace-%s-*.out", time.Now().UTC().Format(time.RFC3339)))
			if err != nil {
				return
			}

			defer f.Close()

			_, err = io.Copy(f, &b)
			if err != nil {
				return
			}
		}
	})
}

type rwWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rec *rwWrapper) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

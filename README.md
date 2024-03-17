# Flight Recorder middleware

This repo contains a middleware that uses the [FlightRecorder](https://pkg.go.dev/golang.org/x/exp/trace#FlightRecorder) from x/exp/trace to record trace data from a running Go application.

## Usage

To setup the middleware, you need to provide a directory where the trace data will be stored, a FlightRecorder instance and a function that will be called for each request to decide if the trace data should be recorded.

```go
fr := trace.NewFlightRecorder()
fr.Start()

frmw, err := flightrecordermw.New("/tmp", fr, func(r *http.Request, stats flightrecordermw.Stats) bool {
    return stats.Elapsed > 100*time.Millisecond
})
if err != nil {
    return err
}
```

To use the middleware, you need to wrap your handler with it.

```go
mux.Handle("/hello", frmw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
})))
```

That's it!

## Handler

The middleware also provides a handler that can be used to serve the trace data.

```go
mux.Handle("/traces/", http.StripPrefix("/traces/", frmw))
```

You will find a listing of all the available traces in the directory provided when creating the middleware.

## Example

You can find a complete example in the [example](./example) directory.

## License (Apache 2)

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details

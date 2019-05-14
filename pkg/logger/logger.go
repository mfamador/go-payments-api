package logger

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type HttpLogEntry struct {
	Severity string `json:"severity"`
	Method   string `json:"method"`
	Uri      string `json:"uri"`
	Time     int    `json:"ts"`
	Usec     int    `json:"usec"`
	Status   int    `json:"status"`
}

type GenericLogEntry struct {
	Severity string      `json:"severity"`
	Msg      string      `json:"msg"`
	Time     int         `json:"ts"`
	Data     interface{} `json:"data"`
}

func Info(msg string, data interface{}) {
	write(&GenericLogEntry{
		Severity: "info",
		Msg:      msg,
		Time:     millis(),
		Data:     data,
	})
}

func Error(err error) {
	write(&GenericLogEntry{
		Severity: "error",
		Msg:      err.Error(),
		Time:     millis(),
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {

	if w.status == 0 {
		w.status = http.StatusOK
	}

	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

func NewHttpLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			// wrap our response write, so that we are able
			// to intercept the status code and the content length
			wrapper := &statusWriter{
				ResponseWriter: w,
			}

			// Define our log function to be called in
			// deferred mode, so that this gets logged even
			// if a panic ocurrs
			defer func() {

				severity := "info"
				if wrapper.status >= 500 {
					severity = "error"
				}

				write(&HttpLogEntry{
					Severity: severity,
					Time:     millis(),
					Method:   r.Method,
					Uri:      r.RequestURI,
					Usec:     int(time.Now().Sub(start) / time.Microsecond),
					Status:   wrapper.status,

					// Add more fields, such as user agent
					// peer ip/port etc..
				})
			}()
			next.ServeHTTP(wrapper, r)
		})
	}
}

func write(e interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.Encode(e)
}

func millis() int {
	now := time.Now()
	return int(now.UnixNano() / 1000000)
}

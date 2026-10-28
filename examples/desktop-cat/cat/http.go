package cat

import (
	"encoding/json"
	"io"
	"net/http"
)

// Handler exposes the loopback agent API. Remote use needs a reverse
// proxy with auth or a shared secret checked before Submit.
func (i *Inbox) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /expressions", func(w http.ResponseWriter, r *http.Request) {
		reply(w, http.StatusOK, Expressions())
	})
	mux.HandleFunc("GET /state", func(w http.ResponseWriter, r *http.Request) {
		if !checkOrigin(r) {
			reply(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
		n, v := i.Latest()
		reply(w, http.StatusOK, struct {
			Notification
			Sequence uint64 `json:"sequence"`
		}{n, v})
	})
	mux.Handle("POST /notify", i.limitNotify(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkOrigin(r) {
			reply(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxNotifyBytes)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var n Notification
		if err := decoder.Decode(&n); err != nil {
			reply(w, http.StatusBadRequest, map[string]string{"error": "invalid notification"})
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			reply(w, http.StatusBadRequest, map[string]string{"error": "send one JSON object"})
			return
		}
		if err := i.Submit(n); err != nil {
			reply(w, http.StatusBadRequest, map[string]string{"error": "invalid notification"})
			return
		}
		reply(w, http.StatusAccepted, n)
	})))
	return withHeaders(withHostCheck(mux))
}

func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

package fixture

import (
	"net/http"
	"strconv"
)

func (s *Server) serveRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, PathOK, http.StatusFound)
}

// serveStatus answers with ?code=.
func (s *Server) serveStatus(w http.ResponseWriter, r *http.Request) {
	code, err := strconv.Atoi(r.URL.Query().Get("code"))
	if err != nil || code < 100 || code > 599 {
		code = http.StatusTeapot
	}

	w.WriteHeader(code)
}

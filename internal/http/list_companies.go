package http

import (
	"encoding/json"
	"net/http"
)

func (s *Server) listCompanies(w http.ResponseWriter, r *http.Request) {
	cs, err := s.rep.ListCompanies(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(cs)
}

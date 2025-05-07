package http

import (
	"encoding/json"
	"net/http"
)

func (s *Server) createCompany(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"company_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	c, err := s.rep.CreateCompany(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(c)
}

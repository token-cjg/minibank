package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/token-cjg/mable-backend-code-test/internal/repo"
)

type Transfer struct{ Repo *repo.Repo }

func NewTransfer(r *repo.Repo) *Transfer { return &Transfer{Repo: r} }

/* POST /transfer  (Content‑Type: text/csv) */
func (h *Transfer) Batch(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != "text/csv" && ct != "text/plain" {
		http.Error(w, "expect Content-Type: text/csv", http.StatusUnsupportedMediaType)
		return
	}

	csvr := csv.NewReader(r.Body)
	csvr.FieldsPerRecord = 3
	defer r.Body.Close()

	var txns []repo.TransferInput
	line := 1
	for {
		rec, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			msg := fmt.Sprintf("bad CSV on line %d: %v", line, err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		src, e1 := strconv.ParseInt(rec[0], 10, 64)
		dst, e2 := strconv.ParseInt(rec[1], 10, 64)
		amt, e3 := strconv.ParseFloat(rec[2], 64)
		if err := firstErr(e1, e2, e3); err != nil {
			msg := fmt.Sprintf("parse error on line %d: %v", line, err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		txns = append(txns, repo.TransferInput{Source: src, Target: dst, Amount: amt})
		line++
	}

	if berr := h.Repo.BatchTransfer(r.Context(), txns); berr != nil {
		status := http.StatusInternalServerError
		if berr.Err == repo.ErrInsufficient {
			status = http.StatusConflict
		}
		resp := map[string]any{"error": berr.Err.Error(), "row": berr.Row + 1}
		writeJSON(w, status, resp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

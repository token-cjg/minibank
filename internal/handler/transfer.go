package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/token-cjg/minibank/internal/repo"
)

type Transfer struct{ Repo *repo.Repo }

func NewTransfer(r *repo.Repo) *Transfer { return &Transfer{Repo: r} }

const FileUploadField = "file"

/* Batch is a handler for transferring money between accounts.
 * 	POST /transfer
 * 	Content-Type: multipart/form-data
 * 	Body: {"file": <file>}
 * 	OR
 * 	Content-Type: text/csv
 * 	Body: <csv data>
 * 	CSV format: <source_account_id>,<target_account_id>,<amount>
 * 	Example:
 * 		1,2,100.50
 * 		3,4,200.00
 * 		5,6,300.75
 * 	Returns:
 * 		201 Created
 * 		400 Bad Request if the CSV is malformed or the request is not multipart/form-data
 * 		500 Internal Server Error if the server encounters an error
 * 	Notes:
 * 		- The CSV file can be uploaded as a file part in a multipart/form-data request.
 * 		- The CSV file can also be sent as a text/csv request body.
 * 		- The CSV file must contain three columns: source_account_id, target_account_id, and amount.
 * 		- The source_account_id and target_account_id must be valid account IDs.
 * 		- The amount must be a valid float64 value.
 * 		- The transfer will be processed in a batch, and the response will indicate the status of the transfer.
 * 		- The transfer will be processed in the order they appear in the CSV file.
 */
func (h *Transfer) Batch(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")

	var csvReader *csv.Reader

	switch {
	case strings.HasPrefix(ct, "multipart/form-data"):
		// Parse up to 10 MB of file parts into memory before spilling to disk
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "bad multipart form: "+err.Error(), http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile(FileUploadField)
		if err != nil {
			http.Error(w, "missing file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		csvReader = csv.NewReader(file)

	case ct == "text/csv" || ct == "text/plain":
		csvReader = csv.NewReader(r.Body)
		defer r.Body.Close()

	default:
		http.Error(w, "expect multipart/form-data or text/csv", http.StatusUnsupportedMediaType)
		return
	}

	csvReader.FieldsPerRecord = 3
	defer r.Body.Close()

	var txns []repo.TransferInput
	line := 1
	for {
		rec, err := csvReader.Read()
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

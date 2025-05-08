package repo_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/token-cjg/minibank/internal/repo"
)

func TestNewRepo(t *testing.T) {
	// Use sqlmock as a stub db for testing New.
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	r := repo.New(db)
	if r == nil {
		t.Fatal("expected non-nil Repo")
	}

	// Verify the sentinel error.
	expectedMsg := "insufficient balance"
	if repo.ErrInsufficient.Error() != expectedMsg {
		t.Errorf("expected ErrInsufficient message to be %q, got %q", expectedMsg, repo.ErrInsufficient.Error())
	}
}

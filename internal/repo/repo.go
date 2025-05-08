// Package repo provides a repository for managing user accounts and transactions.
// It contains methods for creating, updating, and deleting accounts,
// as well as for transferring money between accounts.
package repo

import (
	"database/sql"
	"errors"
)

type Repo struct{ db *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db} }

// ErrInsufficient is a sentinel error for insufficient balance
var ErrInsufficient = errors.New("insufficient balance")

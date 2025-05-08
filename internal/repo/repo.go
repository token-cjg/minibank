package repo

import (
	"database/sql"
	"errors"
)

type Repo struct{ db *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db} }

// Sentinel error for insufficient balance
var ErrInsufficient = errors.New("insufficient balance")

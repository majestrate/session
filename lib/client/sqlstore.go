package client

import (
	"database/sql"
	"github.com/majestrate/ubw/lib/model"
)

type sqlStore struct {
	db *sql.DB
}

func (s *sqlStore) HasMessage(hash string) bool {
	row := s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE hash=?", hash)
	if row == nil {
		return false
	}
	var count int64
	row.Scan(&count)
	return count != 0
}

func (s *sqlStore) Put(msg model.Message) error {
	_, err := s.db.Exec("INSERT INTO messages(hash, contents) VALUES(?,?)", msg.Hash, msg.Raw)
	return err
}

func (s *sqlStore) LastHash() string {
	row := s.db.QueryRow("SELECT hash FROM messages ORDER BY timestamp DESC LIMIT 1")
	if row == nil {
		return ""
	}
	var h string
	row.Scan(&h)
	return h
}

func (s *sqlStore) Close() error {
	return s.db.Close()
}

func (s *sqlStore) migrate() error {
	_, err := s.db.Exec("CREATE TABLE IF NOT EXISTS messages(hash BLOB PRIMARY KEY, contents BLOB NOT NULL, timestamp DATETIME DEFAULT NOW )")
	return err
}

func SQLStore(c *sql.DB) MessageStore {
	s := &sqlStore{db: c}
	err := s.migrate()
	if err != nil {
		panic(err)
	}
	return s
}

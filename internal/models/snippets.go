package models

import (
	"database/sql"
	"errors"
	"time"
)

type SnippetModelService interface {
	Insert(title, content string, expires int) (int, error)
	Get(id int) (*Snippet, error)
	Latest() ([]*Snippet, error)
	AddTag(snippetID int, tag string) error //novo método para adicionar tags
}

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title, content string, expires int) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES (?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *SnippetModel) Get(id int) (*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`

	row := m.DB.QueryRow(stmt, id)
	s := &Snippet{}
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return s, nil
}

func (m *SnippetModel) Latest() ([]*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snippets := []*Snippet{}
	for rows.Next() {
		s := &Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}

// AddTag associa uma tag a um snippet.
func (m *SnippetModel) AddTag(snippetID int, tag string) error {
	// Insere a tag na tabela se ela ainda não existir
	stmt := `INSERT INTO tags (name) VALUES(?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)`
	result, err := m.DB.Exec(stmt, tag)
	if err != nil {
		return err
	}

	// pega o ID da tag inserida ou encontrada
	tagID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Insere o relacionamento na tabela snippet_tags
	stmt = `INSERT INTO snippet_tags (snippet_id, tag_id) VALUES(?, ?)`
	_, err = m.DB.Exec(stmt, snippetID, tagID)
	return err
}
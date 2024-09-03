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
	GetTags(snippetID int) ([]string, error) 
}

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
	Tags    []string  // campo para as tags
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

// GetTags retrieves the tags associated with a given snippet.
func (m *SnippetModel) GetTags(snippetID int) ([]string, error) {
	// SQL query to retrieve tag names for a specific snippet ID
	stmt := `
		SELECT t.name FROM tags t
		INNER JOIN snippet_tags st ON t.id = st.tag_id
		WHERE st.snippet_id = ?`

	// Execute the query
	rows, err := m.DB.Query(stmt, snippetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold the tag names
	var tags []string

	// Loop through the rows and append each tag name to the slice
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	// Check for any error encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Return the slice of tags
	return tags, nil
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
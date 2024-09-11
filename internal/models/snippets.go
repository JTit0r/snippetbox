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
	AddTag(snippetID int, tag string) error // novo método para adicionar tags
	GetTags(snippetID int) ([]string, error) // novo método para recuperar tags
	Search(query string) ([]*Snippet, error) // novo método para a pesquisa de snippets
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

// GetTags recupera as tags associadas a um snippet
func (m *SnippetModel) GetTags(snippetID int) ([]string, error) {
	// Query SQL para recuperar os nomes das tags para um snippet especifico pelo ID
	stmt := `
		SELECT t.name FROM tags t
		INNER JOIN snippet_tags st ON t.id = st.tag_id
		WHERE st.snippet_id = ?`

	// Executa a Query
	rows, err := m.DB.Query(stmt, snippetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// para armazenar os nomes das tags
	var tags []string

	// iterar pelas linhas e fazer append de cada nome 
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	// checar por erros
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// retorna os nomes das tags
	return tags, nil
}

// método pesquisa (recebe query, retorna lista de snippets)
func (m *SnippetModel) Search(query string) ([]*Snippet, error) {
	stmt := `														// query sql retorna snippets que satisfaçam título ou tags, que não tenham expirado
		SELECT s.id, s.title, s.content, s.created, s.expires 
		FROM snippets s
		LEFT JOIN snippet_tags st ON s.id = st.snippet_id
		LEFT JOIN tags t ON st.tag_id = t.id
		WHERE s.expires > UTC_TIMESTAMP() 
		AND (s.title LIKE ? OR t.name LIKE ?)
		GROUP BY s.id
		ORDER BY s.created DESC`

	rows, err := m.DB.Query(stmt, "%"+query+"%", "%"+query+"%")	// execução da query
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snippets := []*Snippet{}	// cria a lista de snippets iterando sobre o resultado da query
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

	return snippets, nil // retorna os snippets encontrados
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

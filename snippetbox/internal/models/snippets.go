package models

import (
	"database/sql"
	"errors"
	"time"
)

// Define a snippet type to hold data for an individual snippet. Notice how
// the fields of the struct corresponds to the fields in our MySQL snippets.
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// Define a SnippetModel which wraps sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

// Insert() a new snippet into database and return snippet id and error.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {

	// Write the SQL statement we want to execute. I've split it over two lines
	// for readability (which is why it's surrounded with backquotes instead of
	// normal double quotes).
	stmt := `INSERT INTO snippets (title, content, created, expires) 
			VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// Use the Exec() method on the embedded connection pool to execute the
	// statement. The first parameter is the SQL statement, followed by the
	// title, content and expiry values for the placeholder parameters. This
	// method return sql.Result type, which contains some basic information
	// about what happened when the statement was executed.
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// Use the LastInsertID() method on the result to get the ID of our newly
	// inserted record in the snippets table
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Get() returns a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (*Snippet, error) {
	// Write the SQL statement we want to execute.
	stmt := `SELECT id, title, content, created, expires FROM snippets
			WHERE expires > UTC_TIMESTAMP() AND id = ?`

	// Use the QueryRow() method on the connection pool to execute our SQL
	// statement, passing in the untrusted id variable as the value for the
	// placeholder parameter. This returns a pointer to sql.Row object which
	// holds the results from the database.

	row := m.DB.QueryRow(stmt, id)

	// Initialize a pointer to a new zeroed Snipped struct.
	s := &Snippet{}

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a
		// sql.ErrNoRows error. We use the error.Is() function check for that
		// error specifically, and return our own ErrNoRecord error instead
		// (we'll create this in a moment)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	// IF everything went OK then return the Snippet object.
	return s, nil
}

// Latest() returns the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	return nil, nil
}

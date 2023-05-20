# Database-Driven Responses

Day 4, May 19 2023

> ထိုက်တန်တဲ့ရင်းနှီးမှုတော့ ရှိရမယ်
မယုတ်မလွန်တဲ့ကြိုးစားမှုတော့ စိုက်ရမယ်
နတ်စီတဲ့ အိမ်မက်လှလှများ
လွယ်လွယ်နဲ့မရ

## Installing a database driver

```
$ cd ~/code/snippetbox
$ go get github.com/go-sql-driver/mysql@v1
```

> Refer page 90 for additional package management.

## Creating a database connection pool

```
// The sql.Open() function initializes a new sql.DB object, which is essentially
// a pool of database connections.
db, err := sql.Open("mysql", "web:frontiir@snippetbox?parseTime=true")
if err != nil {
    ...
}
```

* The first parameter to `sql.Open()` is the *driver* name and the second parameter is the *data source name* (somethims also called a *connection string* or *DSN*).

* The format of data source name will depend on which database and driver you're using. For the driver we're using, you can find that documentation [here](https://github.com/go-sql-driver/mysql#dsn-data-source-name).

* The `parseTime=true` part of the DSN above is *driver-specific* parameter which instructs our driver to convert SQL `Time` and `Date` fields to Go `time.Time` objects.

* The `sql.Open()` function returns [sql.DB](https://pkg.go.dev/database/sql/#DB) object. This isnt' a database connection -- it's a *pool of many connections*. Go manages connections in this pool as needed, automatically opening and closing connections to the database via driver.

* The connection pool is safe for concurrent access, so you can use it from web application handlers safely.

* The connection pool is intended to be long-lived. In a web application, it's normal to initialize the connection pool in your `main()` function and then pass the pool to your handlers. You shouldn't call `sql.Open()` in a short-lived handler itself -- it would be a waste of memory and network resources.

In `main.go`

```
package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")

    // Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "web:frontiri@tcp(172.16.251.171:3306)/snippetbox?parseTime=true",
		"MySQL data source name")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Close closes the database and prevents new queries from starting. Close
	// then waits for all queries that have started processing on the server to
	// finish. It is rare to Close a DB, as the DB handle is meant to be 
	// long-lived and shared between many goroutines.

	// Defer a call to db.Close(), so that the connection pool is closed before
	// the main() function exists
	defer db.Close()

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting http server on %s", *addr)
	err = srv.ListenAndServe()
	if err != nil {
		errorLog.Fatal(err)
	}

}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
```

* Notice how the import path for our driver is prefixed with an underscore? This
is because our `main.go` actually use anything in the `mysql` package. So if we
try to import it, normally the Go compiler will raise an error. However, we need
the driver's `init()` function to run so that it can register itself with the 
`database/sql` package. The trick to getting around this is to alias the package
name to the blank identifier. This is standard practice for most of Go's SQL
drivers.

* The `sql.Open()` function doesn’t actually create any connections, all it does
is initialize the pool for future use. Actual connections to the database are
established lazily,as and when needed for the first time. So to verify that
everything is set up correctly we need to use the `db.Ping()` method to create a
connection and check for any errors.

* At this moment in time, the call to defer `db.Close()` is a bit superfluous.
Our application is only ever terminated by a signal interrupt (i.e. `Ctrl+c` )
or by `errorLog.Fatal()` . In both of those cases, the program exits immediately
and deferred functions are never run. But including `db.Close()` is a good habit
to get into and it could be beneficial later in the future if you add a graceful
shutdown to your application.

## Using the SnippetModel

To use this model in our handlers we need to establish a new `SnippetModel` 
struct in our `main()` function and then inject it as a dependency via the
`application` struct.

```
package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"snippetbox.yehtet.net/snippetbox/internal/models"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	snippet  *models.SnippetModel
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "web:frontiir@tcp(172.16.251.171:9999)/snippetbox?parseTime=true",
		"MySQL data source name")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Close closes the database and prevents new queries from starting. Close
	// then waits for all queries that have started processing on the server to
	// finish. It is rare to Close a DB, as the DB handle is meant to be
	// long-lived and shared between many goroutines.

	// Defer a call to db.Close(), so that the connection pool is closed before
	// the main() function exists.
	defer db.Close()

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippet:  &models.SnippetModel{DB: db},
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting http server on %s", *addr)
	err = srv.ListenAndServe()
	if err != nil {
		errorLog.Fatal(err)
	}

}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
```

## Executing SQL statements

Let's update `SnippetsModel.Insert()` method.

```
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
```

Let's discuss the [sql.Result](https://pkg.go.dev/database/sql/#Result) type 
returned by `DB.Exec()`. This provides two methods:

* `LastInsertId()` -- which returns the integer (an `int64`) generated by the
database in response to a command. Typically thsi will be from an "auto increment"
column when inserting a new row, which is exactly what's happening in our case.

* `RowsAffect()` -- which returns the number of rows (as an `int64`) affected 
by the statement

> **Important:** Not all drivers and databases support these methods. For example,
`LastInsertId()` is [not supported](https://github.com/lib/pq/issues/24) by PostgreSQL.

Also, it is perfectly acceptable (and common) to ignore the `sql.Result` return
value if you don't need it. Like so:
```
_, err := m.DB.Exec("INSERT INTO ." ...)
```

## Using the model in our handlers

```
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	// Create some variables holding dummy data. We will remove these later on.
	title := "0 snail"
	content := "0 snaill\nClimb Mount fuji,\nBut slowly, slowly!\n\n- Kobayashi Issa"
	expires := 7

	// Pass the data to the SnippetModel.Insert() method, receiving the
	// Id of the new recored back.
	id, err := app.snippet.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Redirect the user to the relevant page for the snippet.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
}
```

```
$ curl -iL -X POST http://localhost:4000/snippet/create
HTTP/1.1 303 See Other
Location: /snippet/view?id=6
Date: Sat, 20 May 2023 07:12:46 GMT
Content-Length: 0

HTTP/1.1 200 OK
Date: Sat, 20 May 2023 07:12:46 GMT
Content-Length: 36
Content-Type: text/plain; charset=utf-8

Display a specific snippet with ID 4
```

We've just sent a HTTP request which triggered our `snippetCreate` handler, which
in turn called our `SnippetModel.Insert()` method. This inserted a new record in
the database and return the ID of this new record. Our handler then issued a
redirect to another URL with the ID as a query string paramter.

## Placeholder parameters `` ? ``

The reason for using placeholder parameters to construct our query (rather than
string interpolation) is to help avoid SQL injection attacks from any untrusted
user-provided input.

Behind the scenes, the `DB.Exec()` method works in three steps:

- It creates a new [prepared statement](https://en.wikipedia.org/wiki/Prepared_statement)
on the database using the provided SQL statement. The database parses and compiles
the statement, then stores it ready for execution.

- In a second separate stop, `Exec()` passes the parameter values to the database.
The database then executes the prepared statement using these parameters. Because
the parameters are transmitted later, after the statement has been compiled, the
database treats them as pure data. They can't change the *intent* of the statement.
 So long as the original statement is not derived from an untrusted data, injection
 cannnot occur.

 - It then closes (or deallocates) the prepared statement on the database.
 > placeholder parameter syntax differs depending on your database.

 ## Single-record SQL queries

 ```
 SELECT id, title, content, created, expires FROM snippets
WHERE expires > UTC_TIMESTAMP() AND id = ?
```
Because our `snippets` table uses the `id` column as its primay key, this query
will only ever return one database row (or none at all). The query also includes
a check on the expiry time so that we don't return any snippets that have expired.

```
// internal/models/snippets.go

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
```
Behind the scenes of `rows.Scan()` your driver will automatically convert the
raw output from the SQL database to the required native Go types. So long as
you're sensible with the types that you're mapping between SQL and Go. these
conversions should generally Just Work.

- CHAR , VARCHAR and TEXT map to string .
- BOOLEAN maps to bool .
- INT maps to int ; BIGINT maps to int64 .
- DECIMAL and NUMERIC map to float .
- TIME , DATE and TIMESTAMP map to time.Time .

> **Note:**  A quirk of our MySQL driver is that we need to use the `parseTime=true`
parameter in our DSN to force it to convert `TIME` and `DATE` fields to `time.Time`.
Otherwise it returns these as `[]byte` objects. This is one of the many [driver-specific 
parameters](https://github.com/go-sql-driver/mysql#parameters) that it offers.

```
// internal/models/errors.go

package models

import (
	"errors"
)

var ErrNoRecord = errors.New("models: no matching record found")
```

You might be wondering why we're returning the `ErrNoRecord` error from our
`SnippetModel.Get()` method, instead of `sql.ErrNoRows` directly. The reason is 
to help encapsulate the model completely, so that our application isn't concerned
with the underlying datastore or reliant on datastore-specific errors for its
behavior.

## Using the model in our handlers

```
// cmd/web/handlers.go

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	// Use the SnippetModel object's Get method to retrieve the data for a
	// specific record based on its ID. If no matching record is found,
	// return a 404 Not Found Response.
	snippet, err := app.snippet.Get(id)
	if err != nil {
		// that's why we imported model
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Write the snippet data as a plain-text HTTP response.
	fmt.Fprintf(w, "%v", snippet)

}
```

## Checking for specific errors

From Go 1.13 forward, it's now safer and best practice to use the `errors.Is()`
function instead of the equality operator `==` to perform the check.
```
if err == models.ErrNoRecord {
app.notFound(w)
} else {
app.serverError(w, err)
}

if errors.Is(err, models.ErrNoRecord) {
app.notFound(w)
} else {
app.serverError(w, err)
}
```

This is because Go 1.13 introduced the ability to add additional information to errors by
[wrapping them](https://go.dev/blog/go1.13-errors#wrapping-errors-with-w). If an error happens to get wrapped, a entirely new error value is created —
which in turn means that it’s not possible to check the value of the original underlying error
using the regular `==` equality operator.
In contrast, the `errors.Is()` function works by unwrapping errors as necessary before
checking for a match.
Basically, if you are running Go 1.13 or newer, prefer to use `errors.Is()` . It’s a sensible way
to future-proof your code and prevent bugs caused by you — or any packages that your code
imports — deciding to wrap errors in the future.
There is also another function, `errors.As()` which you can use to check if a (potentially
wrapped) error has a specific type

## Shorthand single-record queries

I’ve deliberately made the code in `SnippetModel.Get()` slightly long-winded to help clarify
and emphasize what is going on behind-the-scenes of your code. In practice, you can shorten the code slightly by leveraging the fact that errors from
`DB.QueryRow()` are deferred until `Scan()` is called. It makes no functional difference, but if
you want it’s perfectly OK to re-write the code to look something like this:

```
func (m *SnippetModel) Get(id int) (*Snippet, error) {
	s := &Snippet{}

	err := m.DB.QueryRow("SELECT ...", id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecordd
		} else {
			return nil, err
		}
	}

	return s, nil
}
```

## Multi-record SQL queries

```
// internal/models/snippets.go

func (m *SnippetModel) Latest() ([]*Snippet, error) {
	// Write the SQL statement we want to execute.
	stmt := `SELECT id, title, content, created, expires FROM snippets
			WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	// Use the Query() method on the connection pool to execute our SQL
	// statement. This returns a sql.Rows resultset containing the result
	// of our query.
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// We defer rows.Close() to ensure the sql.Rows resultset is always
	// properly closed before the Latest() method returns. This defer statement
	// should come *after* you check for an error from the Query() method.
	// Otherwise, if Query() returns an error, you'll get a panic trying to
	// close a nil resultset.
	defer rows.Close()

	// Initialize an empty slice to hold the Snippet structs.
	snippets := []*Snippet{}

	// Use rows.Next() to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by the
	// rows.Scan() method. IF iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed Snippet struct.
		s := &Snippet{}

		// Use rows.Scan() to copy the values from each field in the row to the
		// new Snippet object that we created. Again, the arguments to rows.Scan()
		// must be pointers to the place you want to copy the data into, and the
		// number of arguments must be exactly the same as the number of columns
		// returned by your statement.
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		// Append it to the slice of snippets.
		snippets = append(snippets, s)
	}

	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to call
	// this. Don't assume that a successful iteration was completed over the
	// whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// If everything went OK, then return the Snippets slice.
	return snippets, nil
}
```
> **Important:** Closing a resultset with defer `rows.Close()` is critical in the code above. As
long as a resultset is open it will keep the underlying database connection open… so if
something goes wrong in this method and the resultset isn’t closed, it can rapidly lead
to all the connections in your pool being used up.

```
// cmd/web/handlers.go

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	for _, snippet := range snippets {
		fmt.Fprintf(w, "%+v\n", snippet)
	}
}
```
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
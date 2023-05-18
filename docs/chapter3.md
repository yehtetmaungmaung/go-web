# Configuration and error handling

> tar to mhr khun arr, tar lat mhr asone aphyat, tar shay mhr sate dat

```
func main() {
	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// and some short help text explaining what the flag controls. The value of the
	// flag will be store in the addr variable at runtime.
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Importantly, we use teh flag.Parse() function to parse the command-line flag.
	// This reads in the command-line flag value and assigns it to the addr
	// variable. You need to call this *before* you use the addr variable
	// otherwise it will always contain the default value of ":4000". If any errors
	// are encountered during parsing the application will be terminated.
	flag.Parse()

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	log.Printf("Starting http server on %s", *addr)
	err := http.ListenAndServe(*addr, mux)
	if err != nil {
		log.Fatalf("Failed to start the server: %s", err)
	}
}
```

```
go run ./cmd/web -addr=":9999"
```

## Type conversions
In the code above we’ve used the `flag.String()` function to define the command-line flag. This has the benefit of converting whatever value the user provides at runtime to a `string` type. If the value can’t be converted to a `string` then the application will log an error and exit. Go also has a range of other functions including `flag.Int()` , `flag.Bool()` and `flag.Float64()`. These work in exactly the same way as `flag.String()` , except they automatically convert the command-line flag value to the appropriate type. You can use the `-help` flag to list all the available command-line flags for an application and their accompanying help text:
```
$ go run ./cmd/web -help
Usage of /tmp/go-build3232423234/b001/exe/web:
    -addr string
        HTTP network address(default ":4000")
```

## Leveled logging
`log.Printf()` and `log.Fatal()` output messages via Go's standard logger, which - by default - prefixes messages with the local date and time and writes them to the standard error stream(which should display in terminal window). The `log.Fatal()` function will also call `os.Exit()` after writing the message, causing the application to immediately exit.

> Use `log.New()` to create two new custom loggers.

```

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

    // Use log.New() to create a logger for writing information messages. This takes
    // three parameters: the destination to write the logs to (os.Stdout), a string
    // prefix for message (INFO followed by a tab), and flags to indicate what
    // additional information to include (local date and time). Note that the flags
    // are joined using the bitwise OR operator |.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

    // Write messages using the two new loggers, instead of the standar logger.
	infoLog.Printf("Starting http server on %s", *addr)
	err := http.ListenAndServe(*addr, mux)
	if err != nil {
		errorLog.Fatalf("Failed to start the server: %s", err)
	}
}
```

Example Output:

```
INFO	2023/05/18 10:46:27 Starting http server on :4000
ERROR	2023/05/18 10:46:27 /home/yehtet/Documents/code/snippetbox/cmd/web/main.go:39: Failed to start the server: listen tcp :4000: bind: address already in use
exit status 1
```

## the http.Server error log

By default, if Go's HTTP server encounters an error, it will log it using the standard logger. For consistency, it'd be better to use our new `errorLog` logger instead.

```
func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

    // Initialize a new http.Server struct. We set the Addr and Handler fields so that
    // the server uses teh same network address and routes as before, and set the
    // ErrorLog field so that the server now uses the custom errorLog logger in the
    // event of any problems.
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	infoLog.Printf("Starting server on %s", *addr)
    // Call the listenAndServe() method on our new http.Server struct.
	err := srv.ListenAndServe()
	errorLog.Fatal(err)

}
```

## Additional logging methods

We've used the `Println()`, `Printf()`, `Fatal()` methods to write log messages, but Go provides a [range of other methods](https://pkg.go.dev/log/#Logger). As a rule of thumb, you should avoid using the `Panic()` and `Fatal()` variations outside of your `main()` function -- it's good practice to return errors instead, and only panic or exit directly from `main()`.

## Concurrent logging

Custom loggers created by `log.New()` are concurrency-safe. You can share a single logger and use it across multiple goroutines and in your handlers without needing to worrry about race conditions. That said, if you have multiple loggers writing to the same destination then you need to be careful and ensure that the destination's underlying `Write()` method is also safe for concurrent use.

## Dependency injection

Problem: our `handlers.go` file, the `home` handler function is still writing error messges using Go's standard logger, not the `errorLog` logger that we want to be using.

```
func home(w http.ResponseWriter, r *http.Request) {
...
    ts, err := template.ParseFiles(files...)
    if err != nil {
        log.Print(err.Error()) // This isn't using our new error logger.
        http.Error(w, "Internal Server Error", 500)
        return
    }
    err = ts.ExecuteTemplate(w, "base", nil)
    if err != nil {
        log.Print(err.Error()) // This isn't using our new error logger.
        http.Error(w, "Internal Server Error", 500)
    }
}
```

This raises a good question: how can we make our new `errorLog` logger available to our `home` function from `main()`? And this question generalizes further. Most web applications will have multiple dependencies that their handlers need to access, such as a database connection pool, centralized error handlers, and template caches. What we really want to answer is: how can we make any dependency available to our handlers?

There are a [few different ways](https://www.alexedwards.net/blog/organising-database-access) to do this, the simplest being to just put the dependencies in global variables. But in general, it is good practice to inject dependencies into your handlers. It makes your code more explicit, less error-prone and easier to unit test than if you use global variables.

For applications where all you handlers are in the same package, like ours, a neat way to inject dependencies is to put them into a custom `application` struct, and then define your handler functions as methods against `application`.

In `main.go`

```
// Define an application struct to hold the application-wide dependencies for the web
// application. For now we'll only include fields for the two custom loggers, but
// we'll add more to it as the build progresses.
type application struct {
    errorLog *log.Logger
    infoLog *log.Logger
}

func main() {
    infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
    errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)

    // Initialize a new instance of our application struct, containing the
    // dependencies
    app := &application {
        errorLog: errorLog,
        infoLog: infoLog,
    }

    mux = http.NewServeMux()

    // swap the route declarations to use the application struct's methods
    // as the handler functions.
    mux.HandleFunc("/", app.home)
}
```

In `handler.go`

```
// Change the signature of the home handler so it is defined as a method
// against *application.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
    ...
}
```


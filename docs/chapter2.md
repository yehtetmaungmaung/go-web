# Chapter 2

Day 1, May 16 2023

> One small step, one giant leap

## Fixed path and subtree patterns
Fixed paths don't end with a trailing slash, whereas subtree do end with a trailing slash.
* `/snippet/view` - fixed path, matches only /snippet/view
* `/` - subtree path, matches everything after /

## Restricting the root url pattern
```
if r.URL.Path != "/" {
    http.NotFound(w, r)
    return
}
```

## The DefaultServeMux
> allow you to register without declaring a servermux, like this:
```
func main() {
    http.HandleFunc("/", home)
    _ := http.ListenAndServe(":4000", nil)
}
```
Behind the scenes, these functions register their routes with `DefaultServeMux`. It's just regular serve mux which is initialized by default and stored in a `net/http` global variable. Here's the relevant line from the Go source code:
```
var DefaultServeMuxd = NewServeMux()
```
> Not recommended for production applications. Use locally-scoped your own servermux instead.

Because `DefaultServeMux` is a global variable, any package can access it and register a route.

## Servemux features and quirks
* longer URL patterns always take precedence over shorter ones.
* request URL paths are automatically sanitized. `/foo/bar/..//baz`  =>  `301 Permanent Redirect` to `/foo/baz`
* if a subtree path has been registerd and a request is received for that subtree path without a trailing slash, then the user will automatically be sent a `301 Permanent Redirect` to the subtree path with the slash added.  `/foo`  =>  `/foo/`.

## Host name matching
It's possible to include host names in your URL patterns:
```
mux := http.NewServeMux()
mux.HandleFunc("foo.example.org/", fooHandler)
```

> `servemux` doesn't support routing based on the request method, it doesn't support clean URLs with variables in them, and it doesn't support regexp-based patterns.

## Customizing HTTP headers
```
func snippetCreate(w http.ResponseWriter, r *http.Request) {
    // Use r.Method to check whether the request is using POST or not.
    if r.Method != "POST" {
        // If it's not, use the w.WriteHeader() method to send a 405 status
        // code and the w.Write() method to write a "Method Not Allowed"
        // response body. We then return from the function so that the
        // subsequent code is not executed.
        w.WriteHeader(405)
        w.Write([]byte("Method Not Allowed"))
        return
    }
    w.Write([]byte("Create a new snippet..."))
}
```
* It's only possible to call `w.WriteHeader()` once per response, and after the status code has been written, it can't be changed. If you try to call `w.WriteHeader()` second time, Go will log a warning message.
* If you don't call `w.WriteHeader()` explicitly, then the first call to `w.Write()` will automatically send a `200 OK` status code to the user. So, if you want to send a non-200 status code, you must call `w.WriteHeader()` before any call to `w.Write()`.

## `Allow` header 
Let the user know which request methods are supported for that particular URL.

```
func snippetCreate(w httpResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        // Use the Header().Set() method to add an 'Allow: POST' header to the
        // response header map. The first parameter is teh header name, and
        // the second parameter is the header value.
        w.Header().Set("Allow", "POST")
        w.WriteHeader(405)
        w.Write([]byte("Method Not Allowed"))
        return
    }
    w.Write([]byte("Create a new snippet..."))
}
```
> **Important:**  Changing the response header map after a call to `w.WriteHeader()` or `w.Write()` will have no effect on the headers that the user receives. You need to make sure that your response header map contains all the headers you want before you call these methods.

## the `http.Error` shortcut
If you want to send a non-200 status code and a plain text response body, it's a good opportunity to use the `http.Error()` shortcut. It's a lightweight helper function which takes a given message and status code, then calls the `w.WriteHeader()` and `w.Write()` methods behind-the-scenes for us.
Here's updated code:
```
func snippetCreate(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        w.Header().Set("Allow", "POST")
        // Use the http.Error() to send a 405 status code and "Method Not Allowed"
        // string as the response body
        http.Error(w, "Method Not Allowed", 405)
        return
    }
    w.Write([]byte("Create a new snippet..."))
}
```
> The pattern of passing `http.ResponseWriter` to other functions is super common in Go. In practice, it's quite rare to use the `w.Write()` and `w.WriteHeader()` methods directly.

## The `net/http` constants
- `http.MethodPost` instead of the string `"POST"`
- `http.StatusMethodNotAllowed` instead of the integer `405`
```
func snippetCreate(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.Header().set("Allow", http.MethodPost)
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }
    w.Write([]byte("Create a new snippet..."))
}
```
> **Hint:** Complete list of the net/http package's constants [here](https://pkg.go.dev/net/http/#pkg-constants).

## System-generated headers and content sniffing
When sending a response, Go will automatically set three system-genereated headers for you:
- `Date`
- `Content-Length`
- `Content-Type`

Go will attempt to set the correct `Content-Type` header for you by content sniffing the response body with the [http.DetectContentType()](https://pkg.go.dev/net/http/#DetectContentType) function. If this function can't guess the content type, Go will fall back to setting the header to `Content-Type: application/octet-stream` instead.


>It can't distinguish JSON from plain text. So, by defaut, JSON responses will be sent with a `Content-Type: text/plain; charset=utf-8` header. You can prevent this from happening by settting the correct header manually like so:
```
w.Header().Set("Content-Type", "application/json")
w.Write([]byte(`{"name": "Alex"}`))
```
## Manipulating the header map
We used `w.Header().Set()` to add new header to the response header map. You can also use `Add()`, `Del()`, `Get()`, `Values()` methods to read and manipulate the header map.
```
// Set a new cache-control header. If an existing "Cache-Control" header exists
// it will be overwritten.
w.Header().Set("Cache-Control", "public, max-age=31536000")
// In contrast, the Add() method appends a new "Cache-Control" header and can
// be called multiple times.
w.Header().Add("Cache-Control", "public")
w.Header().Add("Cache-Control", "max-age=31536000")
// Delete all values for the "Cache-Control" header.
w.Header().Del("Cache-Control")
// Retrieve the first value for the "Cache-Control" header.
w.Header().Get("Cache-Control")
// Retrieve a slice of all values for the "Cache-Control" header.
w.Header().Values("Cache-Control")
```

## URL query strings
/snippet/view?id=2

```
func snippetView(w http.ResponseWriter, r *http.Request) {
    // Extract the value of the id parameter from the query string and try to
    // convert it to an integer using the strconv.Atoi() function. IF it can't 
    // be converted to an integerr, or the value is less than 1, we return a
    // 404 page not found response.
    id, err := strconv.Atoi(r.URL.Query().Get("id))
    if err != nil || id < 1 {
        htt.NotFound(w, r)
        return
    }

    // Use the fmt.Fprintf() function to interpolate the id value with our response
    // and write it to the http.ResponseWriter.
    fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}
```


## The io.writer interface

```
func Fprintf(w io.Writer, format string, a ...any) (n int, err error)
```

But we passed it our `http.ResponseWriter` object instead - and it worked fine. We're able to do this because `io.Writer` type is an interface, and the `http.ResponseWriter` object satisfied the interface because it has a `w.Write()` method.

## HTML templating and inheritance

```
func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

    // Initialize a slice containing the paths to the two files. It's important
    // to note that the file containing our base template must be the *first*
    // file in the slice.
	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/pages/home.tmpl.html",
	}

    // Use the template.ParseFiles() functions to read the files and store
    // templates in a template set. Notice that we can pass the slice of file
    // paths as a variadic parameter?
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

    // Use the ExecuteTemplate() method to write the content of the "base"
    // template as the response. The last parameter to ExecuteTemplate()
    // represents any dynamic data that we want to pass in. which for now
    // we'll leave as nil.
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
```

## Serving static files
`http.FileServer` handler receives a request, it will remove the leading slash from the URL path and then search the `./ui/static` directory for the corresponding file to send to ther user. So, for this to work correctly, we must strip the leading `/static` from the URL path before passing it to the `http.FileServer`.
```
func main() {
	mux := http.NewServeMux()

	// Create a file server which serves files out of the "./ui/static" directory.
	// Note that the path given to the http.Dir function is relative to the
	// project directory root.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Use the mux.Handle() function to register the file server as the handler
	// for all URL paths that start with "/static/". For matching paths, we strip
	// the "/static/" prefix before the request the file server.
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	log.Println("Starting http server on :4000")
	err := http.ListenAndServe(":4000", mux)
	if err != nil {
		log.Fatalf("Failed to start the server: %s", err)
	}
}
```

## The http.Handler interface
Handler is an object which satisfied the `http.Handler` interface:
```
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

So in its simplest form, a handler might look something like this:

```
type home struct {}

func (h *home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}
```

Here we have an object (in this case it’s a home struct, but it could equally be a string or function or anything else), and we’ve implemented a method with the signature
`ServeHTTP(http.ResponseWriter, *http.Request)` on it. That’s all we need to make a
handler. You could then register this with a servemux using the `Handle` method like so:

```
mux := http.NewServeMux()
mux.Handler("/", &home{})
```

When this servemux receives a HTTP request for `"/"`, it will then call `ServeHTTP()` method of the `home` struct - which in turn writes the HTTP response.

## Handler functions
Now, creating an object just so we can implement a `ServeHTTP()` method on it is long-winded
and a bit confusing. Which is why in practice it’s far more common to write your handlers as a
normal function (like we have been so far). For example:

```
func home(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}
```

But this `home` function is just a normal function; it doesn't have a `ServeHTTP()` method. So in itself it isn't a handler. Instead we can transform it into a handler using `http.HandlerFunc()` ADAPTER, LIKE SO:

```
mux := HTTP.NewServeMux()
mux.Handle("/", http.HandlerFunc(home))
```

The `http.HandlerFunc()` adapter works by automatically adding a `ServeHTTP()` method to the `home` function. When executed, this `ServeHTTP()` method then simply calls the contents of the original `home` function. It's a roundabout but convenient way of coercing a normal function into satisfying the `http.Handler` interface.

## Chaining handlers

The `http.ListenAndServe()` function takes a `http.Handler` object as the second parameter...
```
func ListenAndServe(addr string, handler Handler) error
```
.. but we've been passing in a servemux. We were able to do this because the servemux also has a `ServeHTTP()` method, meaning that it too satisfies the `http.Handler` interface.

Thinkof the servemux as just being a special kind of handler, which, instead of providing a response itself, passes the request on to a second handler.

When our server receives a new HTTP request, it calls the servemux's `ServeHTTP()` method. This looks up the relevant handler based on the request URL Path, and in turn calls the handler's `ServeHTTP()` method. You can think of a Go web application as a chain of `ServeHTTP()` methods being called one after another.


## Requests are handled concurrently
There is one more thing that’s really important to point out: all incoming HTTP requests are
served in their own goroutine. For busy servers, this means it’s very likely that the code in or
called by your handlers will be running concurrently. While this helps make Go blazingly fast,
the downside is that you need to be aware of (and protect against) [race conditions](https://www.alexedwards.net/blog/understanding-mutexes) when
accessing shared resources from your handlers.


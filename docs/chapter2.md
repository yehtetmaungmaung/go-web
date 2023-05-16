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
* if a subtree path has been registerd and a request is received for that subtree path without a trailing slash, then the user will automatically be sent a `301 Permanent Redirect` to the subtree path with the slash added.  `/foo/`  =>  `/foo/`.

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
* It's only possible to call `w.WriteHeader()` Once per response, and after the status code has been written, it can't be changed. If you try to call `w.WriteHeader()` second time, Go will log a warning message.
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
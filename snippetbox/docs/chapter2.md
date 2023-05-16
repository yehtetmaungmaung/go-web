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


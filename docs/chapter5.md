# Displaying dynamic data

Day 6, May 21 2023

> [This is war. We fight one battle, and then we fight another one until it's done.](https://www.youtube.com/watch?v=DazhkXUHyGI)


```
// cd/web/handlers.go

// Add a snippetView handler function
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	// Use the SnippetModel object's Get method to retrieve the data for a
	// specific record based on its ID. If no matching record is found,
	// return a 404 Not Found Response.
	snippet, err := app.snippets.Get(id)
	if err != nil {
		// that's why we imported model
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Initialize a slice containing the paths to the view.tmpl.html file, plus
	// the base layout and navigation partial that we made earlier.
	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/pages/view.html",
	}

	// Parse the template files ...
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// And then execute them. Notice how we are passing in the snippet data
	// (a models.Snipptet struct) as the final parameter?
	err = ts.ExecuteTemplate(w, "base", snippet)
	if err != nil {
		app.serverError(w, err)
	}

}
```

Within your HTML templates, any dynamic data that you pass in is represented by
the **`.`** character (referred to as dot). In this specific case, the underlying
type of dot will be a `models.Snippet` struct. When the underlying type of dot 
is a struct, you can render (or yield) the value of any exported field in your
templates by postfixing dot with the field name. So, because our `models.Snippet`
struct has a `Title` field, we could yield the snippet title by writing {{.Title}}
in our templates.

```
// ui/html/pages/view.tmpl.html

{{define "title"}}Snippet #{{.ID}}{{end}}

{{define "main"}}
    <div class='snippet'>
        <div class='metadata'>
            <strong>{{.Title}}</strong>
            <span>#{{.ID}}</span>
        </div>
        <pre><code>{{.Content}}</code></pre>
        <div class='metadata'>
            <time>Created: {{.Created}}</time>
            <time>Expires: {{.Expires}}</time>
        </div>
    </div>
{{end}}
```

## Rendering multiple pieces of data

Go's `html/template` package allows you to pass in one —
and only one — item of dynamic data when rendering a template. But in a real-world
application there are often multiple pieces of dynamic data that you want to display in the
same page.
A lightweight and type-safe way to achieve this is to wrap your dynamic data in a struct which
acts like a single ‘holding structure’ for your data.

```
// cmd/web/templates.go

package main

import "snippetbox.yehtet.net/snippetbox/internal/models"

// Define a templateData type to act as the holding structure for any dynamic
// data that we want to pass to our HTML templates. At the moment it only
// contains one field, we we'll add more to it as the build progresses.
type templateData struct {
	Snippet *models.Snippet
}

```

Update `cmd/web/handlers.go`:

```
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	// Use the SnippetModel object's Get method to retrieve the data for a
	// specific record based on its ID. If no matching record is found,
	// return a 404 Not Found Response.
	snippet, err := app.snippets.Get(id)
	if err != nil {
		// that's why we imported model
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Initialize a slice containing the paths to the view.tmpl.html file, plus
	// the base layout and navigation partial that we made earlier.
	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/pages/view.html",
	}

	//Parse the template files ...
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Create an instance of a templateData struct holding the snippet data
	data := &templateData{
		Snippet: snippet,
	}

	// Pass in the templateData struct when executing the template.
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

}
```

Update `ui/html/pages/view.tmpl.html` to use `templateData` struct:

```
{{define "title"}}Snippet #{{.Snippet.ID}}{{end}}

{{define "main"}}
    <div class='snippet'>
        <div class='metadata'>
            <strong>{{.Snippet.Title}}</strong>
            <span>#{{.Snippet.ID}}</span>
        </div>
        <pre><code>{{.Snippet.Content}}</code></pre>
        <div class='metadata'>
            <time>Created: {{.Snippet.Created}}</time>
            <time>Expires: {{.Snippet.Expires}}</time>
        </div>
    </div>
{{end}}
```

## Dynamic content escaping

The `html/template` package automatically escapes any data that is yielded between `{{ }}`
tags. This behavior is hugely helpful in avoiding cross-site scripting (XSS) attacks, and is the
reason that you should use the `html/template` package instead of the more generic
`text/template` package that Go also provides.
As an example of escaping, if the dynamic data you wanted to yield was:

```
<span>{{"<script>alert('xss attack')</script>"}}</span>
```
It would be rendered harmlessly as:
```
<span>&lt;script&gt;alert(&#39;xss attack&#39;)&lt;/script&gt;</span>
```

It’s really important to note that when you’re invoking one template from another template,
dot needs to be explicitly passed or pipelined to the template being invoked. You do this by
including it at the end of each `{{template}}` or `{{block}}` action, like so:

```
{{template "main" .}}
{{block "sidebar" .}}{{end}}
```
As a general rule, my advice is to get into the habit of always pipelining dot whenever you
invoke a template with the `{{template}}` or `{{block}}` actions, unless you have a good
reason not to.

## Calling methods

If the type that you’re yielding between `{{ }}` tags has methods defined against it, you can
call these methods (so long as they are exported and they return only a single value — or a
single value and an error).
For example, if .Snippet.Created has the underlying type `time.Time` (which it does) you
could render the name of the weekday by calling its `Weekday()` method like so:

```
<span>{{.Snippet.Created.Weekday}}</span>
```

You can also pass parameters to methods. For example, you could use the `AddDate()` method
to add six months to a time like so:
```
<span>{{.Snippet.Created.AddDate 0 6 0}}</span>
```
Notice that this is different syntax to calling functions in Go — the parameters are *not*
surrounded by parentheses and are separated by a single space character, not a comma.

## HTML comments with `html/template`

Finally, the `html/template` package always strips out any HTML comments you include in
your templates, including any [conditional comments](https://en.wikipedia.org/wiki/Conditional_comment).
The reason for this is to help avoid XSS attacks when rendering dynamic content. Allowing
conditional comments would mean that Go isn’t always able to anticipate how a browser will
interpret the markup in a page, and therefore it wouldn’t necessarily be able to escape
everything appropriately. To solve this, Go simply strips out all HTML comments.

| Action | Description |
| :--- | :--- |
| `{{if .Foo}} C1 {{else}} C2 {{end}}` | If `.Foo` is not empty then render the content C1, otherwise render the content C2. |
| `{{with .Foo}} C1 {{else}} C2 {{end}}` | If `.Foo` is not empty, then set dot to the value of `.Foo` and render the content `C1`, otherwise render the content `C2`. |
| `{{range .Foo}} C1 {{else}} C2 {{end}}` | If the length of `.Foo` is greater than zero then loop over each element, setting dot to the value of each element and rendering the content `C1`. If the length of `.Foo` is zero then render the content `C2`. The underlying type of `.Foo` must be an array, slice, map, or channel. |

There are a few things about these actions to point out:
- For all three actions the `{{else}}` clause is optional. For instance, you can write
`{{if .Foo}} C1 {{end}}` if there’s no `C2` content that you want to render.
- The empty values are false, 0, any nil pointer or interface value, and any array, slice, map,
or string of length zero.
- It’s important to grasp that the `with` and `range` actions change the value of dot. Once you
start using them, what *dot represents can be different depending on where you are in the
template and what you’re doing*.


The `html/template` package also provides some template functions which you can use to add
extra logic to your templates and control what is rendered at runtime. You can find a
complete listing of functions [here](https://pkg.go.dev/text/template/#hdr-Functions), but the most important ones are:

| Function | Description |
| :--- | :--- |
| `{{eq .Foo .Bar}}` | Yields true if `.Foo` is equal to `.Bar` |
| `{{ne .Foo .Bar}}` | Yields true if `.Foo` is not equal to `.Bar` |
| `{{not .Foo}}` | Yields the boolean negation of `.Foo` |
| `{{or .Foo .Bar}}` | Yields `.Foo` if `.Foo` is not empty; otherwise yields `.Bar` |
| `{{index .Foo i}}` | Yields the value of `.Foo` at index `i`. The underlying type of `.Foo` must be a map, slice or array, and `i` must be an integer value.
| `{{printf "%s-%s" .Foo .Bar}}` | Yields a formatted string containing the `.Foo` and `.Bar ` values. Works in the same way as fmt.Sprintf(). |
| `{{len .Foo}}` | Yields the length of .Foo as an integer. |
| `{{$bar := len .Foo}}` | Assign the length of `.Foo` to the template variable `$bar` |

The final row is an example of declaring a template variable. Template variables are
particularly useful if you want to store the result from a function and use it in multiple places
in your template. Variable names must be prefixed by a dollar sign and can contain
alphanumeric characters only.

## Using the if and range actions

Update `/cmd/web/templates.go`:
```
package main

import "snippetbox.yehtet.net/snippetbox/internal/models"

// Include a Snippets field.
type templateData struct {
	Snippet  *models.Snippet
	Snippets []*models.Snippet
}
```

Update `cmd/web/handler.go`:
```
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

	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/pages/home.tmpl.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Create an instance of a templateData struct holding the slice of snippets
	data := &templateData{
		Snippets: snippets,
	}

	// Pass in the templateData struct when executing the template.
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

}
```

Update `ui/html/pages/home.tmpl.html`:

```
{{define "title"}}Home{{end}}

{{define "main"}}
    <h2>Latest snippets</h2>
    {{if .Snippets}}
        <table>
            <tr>
                <th>Title</th>
                <th>Created</th>
                <th>Id</th>
            </tr>
            {{range .Snippets}}
            <tr>
                <td><a href="/snippet/view?id={{.ID}}">{{.Title}}</a></td>
                <td>{{.Created}}</td>
                <td>#{{.ID}}</td>
            </tr>
            {{end}}
        </table>
    {{else}}
        <p>There's nothing to see here... yet!</p>
    {{end}}
{{end}}
```

## Combining functions

It’s possible to combine multiple functions in your template tags, using the parentheses `()` to
surround the functions and their arguments as necessary.
For example, the following tag will render the content C1 if the length of `Foo` is greater than
99:
```
{{if (gt (len .Foo) 99)}} C1 {{end}}
```
Or as another example, the following tag will render the content `C1` if `.Foo` equals 1 and `.Bar`
is less than or equal to 20:
```
{{if (and (eq .Foo 1) (le .Bar 20))}} C1 {{end}}
```

## Controlling loop behavior

Within a `{{range}}` action you can use the `{{break}}` command to end the loop early, and
`{{continue}}` to immediately start the next loop iteration.

```
{{range .Foo}}
    // Skip this iteration if the .ID value equals 99.
    {{if eq .ID 99}}
        {{continue}}
    {{end}}
    // ...
{{end}}


{{range .Foo}}
    // End the loop if the .ID value equals 99.
    {{if eq .ID 99}}
        {{break}}
    {{end}}
    // ...
{{end}}
```

## Caching templates

There are two main issues at the moment:
1. Each and every time we render a web page, our application reads and parses the relevant
template files using the `template.ParseFiles()` function. We could avoid this duplicated
work by parsing the files once — when starting the application — and storing the parsed
templates in an in-memory cache.
2. There’s duplicated code in the `home` and `snippetView` handlers, and we could reduce this
duplication by creating a helper function.

Let's create an in-memory map with the type `map[string]*template.Template` to cache
the parsed templates. Update `cmd/web/templates.go`:

```
func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}

	// Use the filepath.Glob() function to get a slice of all filepaths that
	// match the pattern.
	pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	// Loop through the page filepaths one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.tmpl.html') from the full filepath
		// and assign it to the name variable.
		name := filepath.Base(page)

		// Create a slice containing filepaths for our base template, any partial
		// and the page.
		files := []string{
			"./ui/html/base.tmpl",
			"./ui/html/partials/nav.tmpl",
			page,
		}

		// Parse the files into a template set.
		ts, err := template.ParseFiles(files...)
		if err != nil {
			return nil, err
		}

		// Add the template set to the map, using the name of the page
		// (like 'home.tmpl.html') as the key.
		cache[name] = ts
	}

	// Return the map
	return cache, nil

}
```

The next step is to initialize this cache in the `main()` function and make it available to our
handlers as a dependency via the `application` struct, like this:

```
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

	// Initialize a new template cache...
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &models.SnippetModel{DB: db},
		templateCache: templateCache,
	}
...
```
So, at this point, we’ve got an in-memory cache of the relevant template set for each of our
pages, and our handlers have access to this cache via the `application` struct.

Let’s now tackle the second issue of duplicated code, and create a helper method so that we
can easily render the templates from the cache.
Open up your `cmd/web/helpers.go` file and add the following `render()` method:

```
func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	// Retrieve the appropriate template set from the cache based on the page
	// name (like 'home.tmpl.html'). If no entry exists in the cache with the
	// provided name, then create a new error and call the serverError() helper
	// method that we made earlier and return.
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("The template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// Write out the provided HTTP status code
	w.WriteHeader(status)

	// Execute the template set and write the response body.
	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
}
```

Now update the `cmd/web/handlers.go`:

```
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

	// Use the new render helper.
	app.render(w, http.StatusOK, "home.tmpl.html", &templateData{Snippets: snippets})

}

// Add a snippetView handler function
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	// Use the SnippetModel object's Get method to retrieve the data for a
	// specific record based on its ID. If no matching record is found,
	// return a 404 Not Found Response.
	snippet, err := app.snippets.Get(id)
	if err != nil {
		// that's why we imported model
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, http.StatusOK, "view.tmpl.htm", &templateData{Snippet: snippet})

}
```

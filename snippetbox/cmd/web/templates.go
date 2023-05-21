package main

import "snippetbox.yehtet.net/snippetbox/internal/models"

// Include a Snippets field.
type templateData struct {
	Snippet  *models.Snippet
	Snippets []*models.Snippet
}

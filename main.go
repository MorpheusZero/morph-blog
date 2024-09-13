package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Embed the content directory containing TXT files.
//
//go:embed content/*.md
//go:embed views/*.html
var embeddedFiles embed.FS

// TemplateCache holds the parsed templates
var templateCache = template.Must(template.ParseFS(embeddedFiles, "views/*.html"))

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

// healthCheckHandler is the handler function for the /health route.
// It responds with a simple "OK" message.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Set the content type to plain text
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// Write the "OK" response
	fmt.Fprintln(w, "OK")
}

// postHandler is the handler function for the /post/:slug route.
// It reads a TXT file based on the slug and returns its content.
func postHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the slug from the URL path
	slug := strings.TrimPrefix(r.URL.Path, "/post/")
	slug = strings.TrimSuffix(slug, "/")

	// Construct the file path based on the slug
	filePath := fmt.Sprintf("content/%s.md", slug)

	// Read the file from the embedded content
	data, err := embeddedFiles.ReadFile(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	html := template.HTML(string(mdToHTML(data)))

	// Prepare data to pass to the template
	templateData := struct {
		Title   string
		Content template.HTML
	}{
		Title:   slug,
		Content: html,
	}

	// Set the content type to plain text and write the file content
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Render the template with the data
	err = templateCache.ExecuteTemplate(w, "base.html", templateData)
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
	}
}

func main() {
	// Create a new ServeMux to handle routes
	mux := http.NewServeMux()

	// Register the /health route with the healthCheckHandler
	mux.HandleFunc("/health", healthCheckHandler)

	// Register the /post/:slug route with the postHandler
	mux.HandleFunc("/post/", postHandler)

	// Start the HTTP server on port 8080
	// The ListenAndServe function takes an address and a handler.
	// Here, we use ":8080" for the address and our ServeMux for the handler.
	fmt.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		// Log any errors that occur while starting the server
		fmt.Printf("Error starting server: %v\n", err)
	}
}

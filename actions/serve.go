package actions

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

// ServeAction serves a generated HTML report locally.
type ServeAction struct{}

// NewServeAction creates a new ServeAction.
func NewServeAction() *ServeAction {
	return &ServeAction{}
}

// Execute starts a small static server for a single report file.
func (a *ServeAction) Execute(c *cli.Context) error {
	file := c.String("file")
	addr := c.String("addr")
	if file == "" {
		return fmt.Errorf("--file cannot be empty")
	}
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("report file '%s': %w", file, err)
	}

	abs, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("resolve report file: %w", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/report.html" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, abs)
	})

	fmt.Printf("Serving %s at http://%s/\n", abs, addr)
	return http.ListenAndServe(addr, handler)
}

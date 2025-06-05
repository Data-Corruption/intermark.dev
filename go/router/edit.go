package router

import (
	"io"
	"net/http"
	"path/filepath"

	"intermark/go/flags"
	"intermark/go/layout"
	"intermark/go/paths"
	"intermark/go/themes"
)

func (r *Router) setupEditRoutes() {
	// serve landing page
	r.Router.Get("/", func(res http.ResponseWriter, req *http.Request) {
		r.editMu.RLock()
		defer r.editMu.RUnlock()

		// refresh
		if err := r.refresh(); err != nil {
			r.log.Errorf("error refreshing edit mode: %v", err)
			http.Error(res, "Edit refresh error", http.StatusInternalServerError)
			return
		}

		// serve index
		data, err := layout.Render(filepath.Join(paths.PUB_DIR, ".index.md"), r.layout.IndexTmpl, r.templates, r.layout, nil, r.debugMode)
		if err != nil {
			r.log.Errorf("error processing index file: %v\n", err)
			http.Error(res, "Index file error", http.StatusInternalServerError)
			return
		} else {
			res.Header().Set("Content-Type", "text/html; charset=utf-8")
			res.Write([]byte(data))
		}
	})

	// serve page
	r.Router.Get("/p/*", func(res http.ResponseWriter, req *http.Request) {
		r.editMu.RLock()
		defer r.editMu.RUnlock()

		// refresh
		if err := r.refresh(); err != nil {
			r.log.Errorf("error refreshing edit mode: %v", err)
			http.Error(res, "Edit refresh error", http.StatusInternalServerError)
			return
		}

		rel := req.URL.Path[3:] // remove "/p/" prefix

		// get sidebar item
		si, err := r.layout.GetSidebarItem(rel)
		if err != nil {
			r.log.Errorf("error getting sidebar item: %s, %v\n", rel, err)
			if err == layout.ErrItemNotFound {
				http.NotFound(res, req)
				return
			}
			http.Error(res, "Sidebar item error", http.StatusInternalServerError)
			return
		}
		if si.Type != "file" {
			r.log.Errorf("sidebar item %s is hidden or not a file\n", rel)
			http.NotFound(res, req)
			return
		}

		// serve page
		if data, err := si.Render(r.templates, r.layout, nil, r.debugMode); err != nil {
			r.log.Errorf("error executing template: %v\n", err)
			http.Error(res, "Template render error", http.StatusInternalServerError)
			return
		} else {
			res.Header().Set("Content-Type", "text/html; charset=utf-8")
			res.Write([]byte(data))
		}
	})

	// serve edit
	r.Router.Get("/edit", func(res http.ResponseWriter, req *http.Request) {
		r.editMu.RLock()
		defer r.editMu.RUnlock()

		// refresh
		if err := r.refresh(); err != nil {
			r.log.Errorf("error refreshing edit mode: %v", err)
			http.Error(res, "Edit refresh error", http.StatusInternalServerError)
			return
		}

		// render edit page
		if err := r.templates.ExecuteTemplate(res, "edit.html", map[string]any{
			"Layout":   r.layout,
			"Themes":   themes.All,
			"EditMode": flags.PresentAny("-e", "--edit"),
			"EditPage": true,
			"Debug":    r.debugMode,
		}); err != nil {
			r.log.Errorf("error executing template: %v\n", err)
			http.Error(res, "Template render error", http.StatusInternalServerError)
		}
	})

	r.Router.Post("/edit-sidebar", func(res http.ResponseWriter, req *http.Request) {
		r.editMu.RLock()
		defer r.editMu.RUnlock()
		// read body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			r.log.Errorf("error reading request body: %v", err)
			http.Error(res, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		// parse JSON
		if err := r.layout.FromJSON(r.ctx, body); err != nil {
			r.log.Errorf("error updating layout from JSON: %v", err)
			http.Error(res, "Invalid JSON or update error: "+err.Error(), http.StatusBadRequest)
			return
		}
		// load templates
		if err := r.loadTemplates(); err != nil {
			r.log.Errorf("error loading templates: %v", err)
			http.Error(res, "Template load error", http.StatusInternalServerError)
			return
		}
		// execute sidebar template
		err = r.templates.ExecuteTemplate(res, "sidebar", map[string]any{
			"Layout":   r.layout,
			"Themes":   themes.All,
			"EditMode": flags.PresentAny("-e", "--edit"),
			"EditPage": true,
			"Debug":    r.debugMode,
		})
		if err != nil {
			r.log.Errorf("error executing template: %v", err)
			http.Error(res, "Template render error", http.StatusInternalServerError)
			return
		}
	})
}

// refresh loads the templates, layout, and runs Tailwind.
func (r *Router) refresh() error {
	if err := r.loadTemplates(); err != nil {
		return err
	}
	if err := r.RunTailwind(); err != nil {
		return err
	}
	if err := r.layout.FromFile(r.ctx); err != nil {
		return err
	}
	return nil
}

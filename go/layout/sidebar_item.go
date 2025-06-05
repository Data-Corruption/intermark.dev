package layout

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"intermark/go/flags"
	"intermark/go/html"
	"intermark/go/paths"
	"intermark/go/stringsx"
	"intermark/go/themes"
)

// SidebarItem represents an item in the sidebar.
type SidebarItem struct {
	// "file", "folder", "link", "label", or "divider"
	Type string `json:"Type"`

	// Raw HTML for icon, if any.
	Icon template.HTML `json:"Icon"`

	// Label is the text to display in the sidebar for this item.
	Label string `json:"Label"`

	// Bold is true if the item should be displayed in bold.
	Bold bool `json:"Bold"`

	// File Only: template to use for rendering the file. Default is "page-nav-side-toc.html".
	Template string `json:"Template"`

	// Folder Only: whether the folder is collapsed by default.
	Collapsed bool `json:"Collapsed"`

	// Folder Only: whether the folder should not be collapsible.
	DisableCollapse bool `json:"DisableCollapse"` // only for folder

	// Link is the URL to navigate to when this item is clicked.
	// For files, this is the path to the file (e.g., "/p/path").
	Link string `json:"Link"`

	// Path is the relative path to the file or folder in PUB_DIR.
	Path string `json:"Path"`

	// 1-indexed position in parent slice. On update, all items are sorted
	// by this(alphabetically if 0), then all set to their final index+1
	Position int `json:"Position"`

	// Children are the child items of this item, if any.
	Children []*SidebarItem `json:"Children"`
}

// Render executes the page for this sidebar item.
func (si *SidebarItem) Render(templates *template.Template, layout *Layout, pathToHash map[string]string, debug bool) (string, error) {
	if si.Type != "file" {
		return "", fmt.Errorf("sidebar item is not a file or hidden: %v", si)
	}
	path := filepath.Join(paths.PUB_DIR, si.Path)
	if out, err := Render(path, si.Template, templates, layout, pathToHash, debug); err != nil {
		return "", fmt.Errorf("error rendering sidebar item %v: %w", si, err)
	} else {
		return out, nil
	}
}

// Render executes the given page with the content of the given filepath as the content.
// Not in SidebarItem.Render() because it's used for the index page as well.
func Render(path, tmpl string, templates *template.Template, layout *Layout, pathToHash map[string]string, debug bool) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}
	if tmpl == "" {
		return "", fmt.Errorf("template is empty")
	}
	if !strings.HasSuffix(path, ".html") && !strings.HasSuffix(path, ".md") {
		return "", fmt.Errorf("path %s must end with .html or .md", path)
	}

	editMode := flags.PresentAny("-e", "--edit")

	// get content
	data, err := html.FromFile(path, map[string]any{
		"Layout":   layout,
		"Themes":   themes.All,
		"EditMode": editMode,
		"Debug":    debug,
	})
	if err != nil {
		return "", fmt.Errorf("error processing file %s: %w", path, err)
	}

	// execute the template with the data
	var outBuf bytes.Buffer
	if err := templates.ExecuteTemplate(&outBuf, tmpl, map[string]any{
		"Layout":   layout,
		"Content":  template.HTML(string(data)),
		"Themes":   themes.All,
		"EditMode": editMode,
		"EditPage": false,
		"Debug":    debug,
	}); err != nil {
		return "", fmt.Errorf("error executing template %s: %w", tmpl, err)
	}
	out := outBuf.String()

	// if pathToHash is not nil, replace the links in the output
	if pathToHash != nil {
		out = stringsx.FastLinkReplace(out, pathToHash)
	}

	return out, nil
}

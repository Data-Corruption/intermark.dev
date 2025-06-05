package layout

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"intermark/go/config"
	"intermark/go/files"
	"intermark/go/flags"
	"intermark/go/html"
	"intermark/go/paths"
	"intermark/go/sins"
	"intermark/go/themes"

	"github.com/Data-Corruption/rlog/logger"
)

var ErrItemNotFound = fmt.Errorf("item not found")

// Layout represents the overall layout of the site.
type Layout struct {
	// Title is the title of the site, displayed in the header and browser tab.
	Title string `json:"Title"`

	// InlineIcon is a version of the site icon to be used inline, e.g., in the header.
	InlineIcon template.HTML `json:"InlineIcon"`

	// IndexTmpl is the template to use for the index page. Default is "page-nav.html".
	IndexTmpl string `json:"IndexTmpl"`

	// Sidebar is the list of sidebar items.
	Sidebar []*SidebarItem `json:"Sidebar"`

	// runtime stuff, not saved

	IconHref string        `json:"-"` // href to the icon, e.g., "/assets/logo.svg"
	IconType string        `json:"-"` // mime type of the icon
	Footer   template.HTML `json:"-"` // footer content, if any
}

func (l *Layout) FromFile(ctx context.Context) error {
	l.Title = ""
	l.Sidebar = nil
	err := files.LoadJSON(paths.LAYOUT, &l)
	if err != nil {
		if os.IsNotExist(err) {
			// if file not found, create a new layout with default values
			l.Title = "Intermark"
			l.IndexTmpl = "page-nav.html"
			l.Sidebar = []*SidebarItem{}
			// write the default layout to file
			if err := files.SaveJSON(paths.LAYOUT, l, 0o644); err != nil {
				return fmt.Errorf("error creating default layout file: %w", err)
			}
			logger.Infof(ctx, "Layout file not found, created default layout: %s", paths.LAYOUT)
		} else {
			return err
		}
	}
	return l.Update(ctx)
}

func (l *Layout) FromJSON(ctx context.Context, data []byte) error {
	l.Title = ""
	l.Sidebar = nil
	if err := json.Unmarshal(data, &l); err != nil {
		return err
	}
	return l.Update(ctx)
}

func (l *Layout) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

// little and lazy way to check if url safe
func isURLSafePath(path string) bool {
	escaped := url.PathEscape(path)
	unescaped, err := url.PathUnescape(escaped)
	if err != nil {
		return false
	}
	return unescaped == path
}

// Update adds missing items to the layout sidebar and removes ones that are not in the filesystem.
func (l *Layout) Update(ctx context.Context) error {
	if config.GetData(ctx).LogLevel == "debug" { // check to avoid $$$ l.Debug()
		logger.Debugf(ctx, "Pre Update Layout:\n\n%s\n", l.Debug())
	}

	// flatten current sidebar by path
	fsTypeMap := make(map[string]*SidebarItem)
	nonFsTypeMap := make(map[string][]*SidebarItem)
	l.Walk(func(si *SidebarItem) (bool, error) {
		if si.Type == "folder" || si.Type == "file" {
			if si.Path != "" {
				fsTypeMap[si.Path] = si
			}
		} else {
			nonFsTypeMap[si.Path] = append(nonFsTypeMap[si.Path], si)
		}
		return false, nil
	})

	// tree nodes map: path -> pointer to SidebarItem
	tree := make(map[string]*SidebarItem)
	// virtual root
	root := &SidebarItem{Children: nonFsTypeMap[""]}
	tree[""] = root

	// walk filesystem
	err := filepath.WalkDir(paths.PUB_DIR, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == paths.PUB_DIR {
			return nil
		}
		rel, err := filepath.Rel(paths.PUB_DIR, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		parts := strings.Split(rel, "/")
		// if any parts start with ".", skip
		for _, part := range parts {
			if strings.HasPrefix(part, ".") {
				return nil
			}
		}
		parent := ""
		for i, name := range parts { // if not root item, loop through path parts and create parent nodes
			cur := strings.Join(parts[:i+1], "/")
			if !d.IsDir() && !isURLSafePath(name) {
				return fmt.Errorf("path is not url safe, please rename: %s", cur)
			}
			// if not yet created, build node
			if _, exists := tree[cur]; !exists {
				t := sins.Ternary(d.IsDir() || i < len(parts)-1, "folder", "file") // if dir or we are not at the last part yet
				node := &SidebarItem{Type: t, Path: cur, Label: name, Template: "page-nav-side-toc.html", Children: nonFsTypeMap[cur]}
				if o := fsTypeMap[cur]; o != nil {
					node.Label = o.Label
					node.Bold = o.Bold
					node.Link = o.Link
					node.Template = sins.Ternary(o.Template == "", node.Template, o.Template)
					node.Icon = o.Icon
					node.Position = o.Position
					node.Collapsed = o.Collapsed
					node.DisableCollapse = o.DisableCollapse
				}
				// if node is a file, set link to cur - ext
				if node.Type == "file" {
					node.Link = "/p/" + strings.TrimSuffix(node.Path, filepath.Ext(node.Path))
				}
				tree[cur] = node
				tree[parent].Children = append(tree[parent].Children, node)
			}
			parent = cur
		}
		return nil
	})
	if err != nil {
		return err
	}

	if config.GetData(ctx).LogLevel == "debug" {
		logger.Debugf(ctx, "Tree:\n\n%s\n", func() string {
			out := ""
			for k, v := range tree {
				if k == "" { // not sure why "" in in the tree here and i'm to lazy to think, its working so...
					continue
				}
				out += fmt.Sprintf("%s: %s\n", k, v.Label)
			}
			return out
		}())
	}

	// normalize sorting and positions
	var normalize func(node *SidebarItem)
	normalize = func(node *SidebarItem) {
		// sort by Position>0 first, then by Label for Position==0 or ties
		sort.Slice(node.Children, func(i, j int) bool {
			a, b := node.Children[i], node.Children[j]
			// both have pos
			if a.Position > 0 && b.Position > 0 {
				return a.Position < b.Position
			}
			// one has pos
			if a.Position > 0 {
				return true
			}
			if b.Position > 0 {
				return false
			}
			// both pos zero
			return a.Label < b.Label
		})
		// update positions to index+1
		for i := range node.Children {
			node.Children[i].Position = i + 1
			normalize(node.Children[i])
		}
	}
	normalize(root)
	l.Sidebar = root.Children
	if config.GetData(ctx).LogLevel == "debug" { // check to avoid $$$ l.Debug()
		logger.Debugf(ctx, "Post Update Layout:\n\n%s\n", l.Debug())
	}

	// load icon
	iconPath, found := files.FirstExists(paths.ASS_DIR, "icon.ico", "icon.svg", "icon.png", "icon.jpg", "icon.jpeg", "icon.avif")
	iconExt := filepath.Ext(iconPath)
	if found {
		l.IconHref = "/assets/icon" + iconExt
		switch iconExt {
		case ".ico":
			l.IconType = "image/x-icon"
		case ".svg":
			l.IconType = "image/svg+xml"
		case ".png":
			l.IconType = "image/png"
		case ".jpg", ".jpeg":
			l.IconType = "image/jpeg"
		case ".avif":
			l.IconType = "image/avif"
		default:
			logger.Warnf(ctx, "Unsupported icon file type: %s", iconExt)
		}
	}

	// load footer
	l.Footer = ""
	fPath := filepath.Join(paths.PUB_DIR, ".footer.md")
	if exists, err := files.Exists(fPath); err != nil {
		logger.Errorf(ctx, "issue checking for footer file %s", err.Error())
	} else if exists {
		fData, err := html.FromFile(fPath, map[string]any{
			"Layout":   l,
			"Themes":   themes.All,
			"EditMode": flags.PresentAny("-e", "--edit"),
			"Debug":    config.GetData(ctx).LogLevel == "debug",
		})
		if err != nil {
			return fmt.Errorf("error converting footer markdown to html: %w", err)
		}
		l.Footer = template.HTML(fData)
	} else {
		logger.Warn(ctx, "No footer file found")
	}

	logger.Debugf(ctx, "Layout updated:\n\tTitle: %s\n\tInlineIcon Len: %d\n\tIndexTmpl: %s\n\tSidebar Items: %d\n\tIconHref: %s\n\tIconType: %s\n\tFooter Len: %d",
		l.Title, len(l.InlineIcon), l.IndexTmpl, len(l.Sidebar), l.IconHref, l.IconType, len(l.Footer),
	)

	return files.SaveJSON(paths.LAYOUT, l, 0o644)
}

// helper func for recursing through the sidebar, exiting if f returns true or an error
func (l *Layout) Walk(f func(*SidebarItem) (bool, error)) error {
	var enum func(item *SidebarItem) (bool, error)
	enum = func(item *SidebarItem) (bool, error) {
		if stop, err := f(item); err != nil || stop {
			return stop, err
		}
		for i := range item.Children {
			if stop, err := enum(item.Children[i]); err != nil || stop {
				return stop, err
			}
		}
		return false, nil
	}
	for i := range l.Sidebar {
		if stop, err := enum(l.Sidebar[i]); err != nil || stop {
			return err
		}
	}
	return nil
}

func (l *Layout) Debug() string {
	out := fmt.Sprintf("Layout: %s\n", l.Title)

	var printItem func(item *SidebarItem, prefix string, isLast bool)
	printItem = func(item *SidebarItem, prefix string, isLast bool) {
		// pick connector just by isLast
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// print the current node
		out += fmt.Sprintf("%s%s%s (Type=%s, Path=%s, Pos=%d)\n",
			prefix, connector,
			item.Label, item.Type, item.Path, item.Position,
		)

		// build the indent for children
		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}

		// recurse
		for i, child := range item.Children {
			last := i == len(item.Children)-1
			printItem(child, childPrefix, last)
		}
	}

	// kick it off for each top‑level item
	for i, item := range l.Sidebar {
		last := i == len(l.Sidebar)-1
		printItem(item, "", last)
	}

	// return the output
	return out
}

// Path does not include extension
func (l *Layout) GetSidebarItem(path string) (*SidebarItem, error) {
	var item *SidebarItem
	err := l.Walk(func(si *SidebarItem) (bool, error) {
		if strings.TrimSuffix(si.Path, filepath.Ext(si.Path)) == path {
			item = si
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrItemNotFound
	}
	return item, nil
}

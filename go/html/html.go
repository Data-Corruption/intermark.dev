package html

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"

	"intermark/go/templates"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	_html "github.com/yuin/goldmark/renderer/html"
	"golang.org/x/net/html"
)

var markdown goldmark.Markdown

func init() {
	markdown = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			_html.WithHardWraps(),
			_html.WithUnsafe(),
		),
	)
}

// FromFile reads a file from the given path, converts it from Markdown to HTML if it's a Markdown file,
// and adds IDs to headers if missing.
func FromFile(path string, tmplData map[string]any) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// extract raw blocks if present
	dataStr, raws, err := extractRawBlocks(string(data))
	if err != nil {
		return nil, fmt.Errorf("error extracting raw blocks from file %s: %w", path, err)
	}
	data = []byte(dataStr)

	if strings.HasSuffix(strings.ToLower(path), ".md") {
		data, err = FromMarkdown(data)
		if err != nil {
			return nil, fmt.Errorf("error converting markdown file %s: %w", path, err)
		}
	}

	data, err = idHeaders(data)
	if err != nil {
		return nil, fmt.Errorf("error adding IDs to headers in file %s: %w", path, err)
	}

	funcs := template.FuncMap{"dict": templates.Dict}
	cnt_tmpl := template.New("").Funcs(funcs)
	cnt_out, err := cnt_tmpl.Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("error parsing content as template %s: %w", path, err)
	}
	// execute cnt template
	var dataBuf bytes.Buffer
	if err := cnt_out.Execute(&dataBuf, tmplData); err != nil {
		return nil, fmt.Errorf("error executing content as template %s: %w", path, err)
	}
	cntStr := dataBuf.String()

	// if there are raws, swap them back in
	if len(raws) > 0 {
		for k, v := range raws {
			cntStr = strings.ReplaceAll(cntStr, k, v)
		}
	}

	return []byte(cntStr), nil
}

// FromMarkdown converts a Markdown string to HTML
func FromMarkdown(md []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := markdown.Convert(md, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// idHeaders adds IDs to headers in a Markdown document.
func idHeaders(data []byte) ([]byte, error) {
	root, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// give them bitches IDs
	forEachHeader(root, func(n *html.Node) {
		id := getID(n)
		if id == "" {
			// slugify the header text to create an ID
			id = slugify(n)
			if id == "" {
				return // skip if slugify failed
			}
			// ensure the ID is unique
			existing := getElementById(root, id)
			if existing != nil {
				// append a number to make it unique
				i := 1
				for existing != nil {
					if i > 100 {
						panic(fmt.Sprintf("too many elements with ID %s", id))
					}
					id = fmt.Sprintf("%s-%d", id, i)
					existing = getElementById(root, id)
					i++
				}
			}
			// set the ID attribute
			n.Attr = append(n.Attr, html.Attribute{
				Key: "id",
				Val: id,
			})
		}
	})

	var buf bytes.Buffer
	if err := html.Render(&buf, root); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func slugify(n *html.Node) string {
	text := getTextContent(n)
	slug := strings.ToLower(strings.Join(strings.Fields(text), "-"))
	// replace any non-alphanumeric characters with a hyphen
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return '-'
	}, slug)
}

// forEachHeader traverses the HTML node tree and calls the given function for each header element found.
func forEachHeader(n *html.Node, fn func(*html.Node)) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			fn(n)
			return // don't traverse children of headers
		case "script", "style":
			return // skip script and style elements
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEachHeader(c, fn)
	}
}

func extractRawBlocks(src string) (out string, raws map[string]string, err error) {
	raws = make(map[string]string)
	var builder strings.Builder
	index := 0

	// constants for the literal tag strings (adjust spacing if needed)
	const openTag = "{{< raw >}}"
	const closeTag = "{{< /raw >}}"

	for {
		// find the next opening tag
		start := strings.Index(src, openTag)
		if start < 0 {
			// no more openers—copy the rest and break
			builder.WriteString(src)
			break
		}

		// write everything up to that opener into our output
		builder.WriteString(src[:start])
		src = src[start+len(openTag):] // consume the opener

		// now initialize depth = 1, and look for the matching closer
		depth := 1
		scan := 0
		for scan < len(src) {
			// find next occurrence of either openTag or closeTag
			nextOpen := strings.Index(src[scan:], openTag)
			if nextOpen != -1 {
				nextOpen += scan
			}
			nextClose := strings.Index(src[scan:], closeTag)
			if nextClose != -1 {
				nextClose += scan
			}

			// if no closer at all → error
			if nextClose < 0 {
				return "", nil, fmt.Errorf("unmatched raw tag")
			}

			// if we see an opener before we see the closer, bump depth
			if nextOpen >= 0 && nextOpen < nextClose {
				depth++
				scan = nextOpen + len(openTag)
				continue
			}

			// otherwise we see a closer
			depth--
			// if depth==0, this closer matches our original opener
			if depth == 0 {
				// “body” is everything from start of src to nextClose
				body := src[:nextClose]
				key := fmt.Sprintf("@@RAW%d@@", index)
				raws[key] = body
				builder.WriteString(key)
				index++

				// consume up through the closer
				src = src[nextClose+len(closeTag):]
				break
			}

			// still inside nested raw, consume up through this closer
			scan = nextClose + len(closeTag)
		}

		if depth != 0 {
			return "", nil, fmt.Errorf("unmatched raw tag")
		}
		// continue loop in case there are more raw blocks
	}

	return builder.String(), raws, nil
}

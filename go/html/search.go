package html

import (
	"bytes"
	"errors"
	"strings"

	"intermark/go/sins"

	"github.com/Data-Corruption/rlog/logger"
	"golang.org/x/net/html"
)

// Doc is the halfway point between HTML and lunrjs index.
type Doc struct {
	ID    string `json:"id"`    // relpath + frag
	URL   string `json:"url"`   // /p/ + relpath + frag
	Title string `json:"title"` // Header text
	Body  string `json:"body"`  // All text content until next header
}

// ExtractDocs, data in, docs out. Logs if given a logger.
func ExtractDocs(relpath string, data []byte, docs *[]Doc, log *logger.Logger) error {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	content := getElementById(doc, "_content")
	if content == nil {
		return errors.New("no element with id '_content' found")
	}
	extractDocsTraverse(content, relpath, docs, log)
	return nil
}

func extractDocsTraverse(n *html.Node, relpath string, docs *[]Doc, log *logger.Logger) {
	if hasAttr(n, "data-nosearch") || hasAttr(n.Parent, "data-nosearch") {
		return
	}

	// handle headers
	if n.Type == html.ElementNode {
		switch n.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			id := getID(n)
			if id == "" {
				return
			}
			// strip .md or .html from the relpath
			routePath := strings.TrimSuffix(relpath, ".md")
			routePath = strings.TrimSuffix(routePath, ".html")
			frag := "#" + id
			url := sins.Ternary(routePath == "/", "/", "/p/"+routePath) + frag
			doc := Doc{
				ID:    routePath + frag,
				URL:   url,
				Title: getTextContent(n),
				Body:  "",
			}
			*docs = append(*docs, doc)
			if log != nil {
				log.Debugf("Found header <%s>: %s", n.Data, frag)
			}
			return // don't traverse children of headers
		}
	}

	// for any node (element or text), add its text content to the current doc
	if len(*docs) > 0 {
		text := ""
		switch n.Type {
		case html.TextNode:
			text = n.Data
		case html.ElementNode:
			if n.Data == "script" || n.Data == "style" {
				return
			}
		}
		// if got text, append it to the current doc
		if text != "" && strings.TrimSpace(text) != "" {
			(*docs)[len(*docs)-1].Body += text
			if log != nil {
				shortText := ""
				if len(strings.TrimSpace(text)) > 50 {
					shortText = strings.TrimSpace(text)[:50] + "..."
				} else {
					shortText = strings.TrimSpace(text)
				}
				log.Debugf("Appending text to doc %s: %s", (*docs)[len(*docs)-1].ID, shortText)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractDocsTraverse(c, relpath, docs, log)
	}
}

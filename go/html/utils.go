package html

import (
	"strings"

	"golang.org/x/net/html"
)

func getTextContent(n *html.Node) string {
	var buf strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return buf.String()
}

func getElementById(n *html.Node, id string) *html.Node {
	if n.Type == html.ElementNode {
		if getID(n) == id {
			return n
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := getElementById(c, id); found != nil {
			return found
		}
	}
	return nil
}

func getID(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "id" {
			return attr.Val
		}
	}
	return ""
}

func hasAttr(n *html.Node, key string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return true
		}
	}
	return false
}

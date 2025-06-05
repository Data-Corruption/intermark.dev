package stringsx

import "strings"

const delimiters = " \t\n\"'()<>[]{}"

// FastLinkReplace replaces /assets/ links with /a/hash links.
// It skips escaped ones with extra slashes, removing one in the process.
func FastLinkReplace(input string, pathToHash map[string]string) string {
	var b strings.Builder
	b.Grow(len(input))
	const token = "/assets/"
	const tokenLen = len(token)
	for i := 0; i < len(input); {
		// find next asset
		off := strings.Index(input[i:], token)
		if off < 0 {
			b.WriteString(input[i:])
			break
		}
		start := i + off
		// write everything up to the asset
		b.WriteString(input[i:start])
		// locate end of the asset path
		j := start + tokenLen
		if pos := strings.IndexAny(input[j:], delimiters); pos >= 0 {
			j += pos
		} else {
			j = len(input)
		}
		asset := input[start:j]
		if start >= 1 && input[start-1] == '/' { // if 2+ leading '/', drop one e.g. "///assets/foo.png" -> "//assets/foo.png"
			b.WriteString(asset[1:])
		} else if hash, ok := pathToHash[asset]; ok {
			b.WriteString("/a/" + hash)
		} else { // fallback
			b.WriteString(asset)
		}
		// advance past consumed asset
		i = j
	}
	return b.String()
}

// Package jsonsafe holds a JSON document from outside to a size and a depth
// before it is parsed (NFR-10): a TileJSON document, a user's style.
package jsonsafe

// Within reports whether a document is no longer than maxBytes and nests no
// deeper than maxDepth, counting brackets outside strings. It does not say
// the document is JSON; the parser says that, afterwards.
func Within(body []byte, maxBytes, maxDepth int) bool {
	if maxBytes <= 0 || maxDepth <= 0 {
		return false
	}
	if len(body) == 0 || len(body) > maxBytes {
		return false
	}
	depth, inString, escaped := 0, false, false
	for _, c := range body {
		switch {
		case escaped:
			escaped = false
		case inString:
			escaped = c == '\\'
			inString = c != '"'
		case c == '"':
			inString = true
		case c == '{' || c == '[':
			depth++
			if depth > maxDepth {
				return false
			}
		case c == '}' || c == ']':
			depth--
		}
	}
	return true
}

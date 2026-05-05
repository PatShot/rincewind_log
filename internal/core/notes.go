// Package core provides core logic for handling the Notes and Database structures
package core

import (
	"fmt"
	"strings"
)

// Note contains the structure of the note, comprising of
// ID - string - the ID of the note being taken
// Meta - any meta fields (Book, Chapter, etc.)
// Content - the content of the note
type Note struct {
	ID      string
	Subject string
	Meta    map[string]string
	Content string
}

// Returns a String for either writing the note to file, or writing to Log.
func (n *Note) String() string {
	var sb strings.Builder

	if book, ok := n.Meta["book"]; ok {
		fmt.Fprintf(&sb, "**Book:** %s\n", book)
	}
	if chapter, ok := n.Meta["chapter"]; ok {
		fmt.Fprintf(&sb, "**Chapter:** %s\n", chapter)
	}

	sb.WriteString(n.Content)

	return sb.String()
}

package parser

type TokenType int
type EntryType string

const (
	TokenShortcut TokenType = iota
	TokenNoteFlag
	TokenCodeMarker
	TokenLiteral
)

const (
	TypeNote  EntryType = "note"
	TypeLog   EntryType = "log"
	TypeTodo  EntryType = "todo"
	TypeQuote EntryType = "quote"
	TypeTask  EntryType = "task"
	TypeDone  EntryType = "done"
)

type Token struct {
	Type    TokenType
	Literal string // Raw string
	Key     string // if this is a shortcut defined in shorthand.go, this holds the mapped key.
}

// Log Entry is the Data Access Obj
type LogEntry struct {
	ID       string
	BatchID  string
	Sequence int
	Date     string
	Time     string
	Subject  string
	Meta     map[string]string
	Type     EntryType
	Content  string
}

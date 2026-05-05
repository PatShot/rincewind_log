package core

var NotesMap = map[string]string{
	"b ":  "**Book:** ",
	"c ":  "**Chapter:** ",
	"n ":  "* ",     // n for note
	"q ":  "> ",     // q for quote
	"td ": "- [ ] ", // td for todo
	"tx ": "- [x] ", // tx for task done
}

var MultiLinePrefix = map[string]string{
	"xtx": "text",
	"xgo": "go",
	"xpy": "python",
	"xjs": "javascript",
	"xru": "rust",
}

const ERASE_DEFAULT_STRING = "xclr"

var MetaMap = map[string]string{
	"b ": "book",
	"c ": "chapter",
}

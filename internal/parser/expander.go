package parser

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rincewind_log/internal/core"
)

type Expander struct {
	shortcuts       map[string]string
	contentType     map[string]EntryType
	multiLinePrefix map[string]string
	incode          bool
	codeLanguage    string
	defaultBook     string
	defaultChapter  string
	subject         string
	batchID         string
}

func NewExpander(subject string, activeMeta map[string]string, batchID string) *Expander {
	return &Expander{
		shortcuts: core.MetaMap,
		contentType: map[string]EntryType{
			"n":  TypeNote,
			"q":  TypeQuote,
			"td": TypeTodo,
			"tx": TypeDone,
			"+t": TypeTask,
		},
		multiLinePrefix: core.MultiLinePrefix,
		incode:          false,
		defaultBook:     activeMeta["book"],
		defaultChapter:  activeMeta["chapter"],
		subject:         subject,
		batchID:         batchID,
	}
}

// lexLine scans a string and emits Tokens based on the current state.
// This follows the grammar <metaPart><contentTrigger><content string>
// Of course, this is negated if any of this is inCode.
func (e *Expander) lexLine(line string) []Token {
	trimmedLine := strings.TrimSpace(line)
	if trimmedLine == "" {
		return nil
	}

	// 1. Check if Code Block starts from Multiline prefixes defined in the Expander
	if lang, isCodeTrigger := e.multiLinePrefix[trimmedLine]; isCodeTrigger {
		return []Token{
			{Type: TokenCodeMarker, Key: lang, Literal: lang},
		}
	}

	// 2. Check if Code Block starts from a general ``` <backticks>
	if strings.HasPrefix(trimmedLine, "```") {
		lang := strings.TrimPrefix(trimmedLine, "```")
		return []Token{
			{Type: TokenCodeMarker, Key: strings.TrimSpace(lang), Literal: trimmedLine},
		}
	}

	// 3. If inside a CodeBlock, then everything is a Literal
	if e.incode {
		return []Token{
			{Type: TokenLiteral, Key: "", Literal: line},
		}
	}

	var tokens []Token

	// 4. Hunt for the Content Triggers
	metaPart := trimmedLine
	notePart := ""
	var foundTrigger string

	for trigger := range e.contentType {
		// Check if it starts exactly with the trigger
		if strings.HasPrefix(trimmedLine, trigger+" ") || trimmedLine == trigger {
			metaPart = ""
			notePart = strings.TrimSpace(strings.TrimPrefix(trimmedLine, trigger))
			foundTrigger = trigger
			break
		}
		// Check if the trigger is buried in the middle
		idx := strings.Index(trimmedLine, " "+trigger+" ")
		if idx != -1 {
			metaPart = strings.TrimSpace(trimmedLine[:idx])
			notePart = strings.TrimSpace(trimmedLine[idx+len(trigger)+1:])
			foundTrigger = trigger
			break
		}
	}

	// 5. Tokenize Metadata
	if metaPart != "" {
		words := strings.Fields(metaPart)
		var currentLiteral []string

		flushLiteral := func() {
			if len(currentLiteral) > 0 {
				tokens = append(tokens, Token{
					Type:    TokenLiteral,
					Literal: strings.Join(currentLiteral, " "),
				})
				currentLiteral = []string{}
			}
		}

		for _, word := range words {
			if mappedKey, isShortcut := e.shortcuts[word]; isShortcut {
				flushLiteral()
				tokens = append(tokens, Token{
					Type:    TokenShortcut,
					Literal: word,
					Key:     mappedKey,
				})
			} else {
				currentLiteral = append(currentLiteral, word)
			}
		}
		flushLiteral()
	}

	// 6. Tokenize the Content Literals and the Note; RHS grammar
	if foundTrigger != "" {
		tokens = append(tokens, Token{
			Type:    TokenNoteFlag,
			Literal: foundTrigger,
			Key:     foundTrigger,
		})
		// If note is not blank, then Token Changes to a Literal
		if notePart != "" {
			tokens = append(tokens, Token{
				Type:    TokenLiteral,
				Literal: notePart,
			})
		}
	}

	return tokens
}

// Parse each line in the input, in such a way that tokens generated from lexLine
// Are turned into metablocks and notes.
func (e *Expander) Parse(input string) []LogEntry {
	var entries []LogEntry
	lines := strings.Split(input, "\n")

	var activeEntry *LogEntry
	var contentBuffer []string
	seqCounter := 0

	// flush the contentBuffer to save the note
	// If activeEntry is not nil and there still are elements within content Buffer,
	// activeEntry is added to remaining entries
	// and then the contentBuffer turns into a blank string array
	flushEntry := func() {
		if activeEntry != nil && len(contentBuffer) > 0 {
			activeEntry.Content = strings.Join(contentBuffer, "\n")
			entries = append(entries, *activeEntry)
		}
		contentBuffer = []string{}
	}

	// For lines in input, get Token array from lexLine, then loop over Token array
	// with a switch and case for each declared Token Type, where we can add actions
	// per Token Type.
	for _, line := range lines {
		tokens := e.lexLine(line)
		var pendingKey string

		// If tokens are all empty, preserve spacing for multiline note
		if tokens == nil {
			if activeEntry != nil {
				contentBuffer = append(contentBuffer, line)
			}
			continue
		}

		for _, token := range tokens {
			switch token.Type {
			case TokenCodeMarker:
				e.incode = !e.incode
				if e.incode {
					// Opening ticks for md, and language
					// ```python
					e.codeLanguage = token.Key
					contentBuffer = append(contentBuffer, "```"+token.Key)
				} else {
					// Closing ticks for md
					// ```
					e.codeLanguage = ""
					contentBuffer = append(contentBuffer, "```")
				}
			case TokenShortcut:
				// Since it's a shortcut, see if it's a Literal or a NoteFlag
				// The key is Pending a look.
				pendingKey = token.Key
			case TokenLiteral:
				if pendingKey != "" {
					val := token.Literal
					// Wipe State Memory
					if val == core.ERASE_DEFAULT_STRING {
						val = ""
					}
					// Add logic for any new pendingKeys here
					if pendingKey == "book" {
						e.defaultBook = val
					} else if pendingKey == "chapter" {
						e.defaultChapter = val
					}
					pendingKey = ""
				} else {
					contentBuffer = append(contentBuffer, token.Literal)
				}
			case TokenNoteFlag:
				// When hitting a flag like 'td' or 'n', Save the note, by
				// flushing whatever was being built in contentBuffer
				flushEntry()
				entryType := e.contentType[token.Key]

				now := time.Now()

				activeEntry = &LogEntry{
					ID:       uuid.New().String(),
					BatchID:  e.batchID,
					Sequence: seqCounter,
					Date:     now.Format("2006-01-02"),
					Time:     now.Format("15:04:05"),
					Subject:  e.subject,
					Type:     entryType,
					Meta: map[string]string{
						"book":    e.defaultBook,
						"chapter": e.defaultChapter,
					},
				}
				seqCounter++
			default:
				panic("Unknown token type")
			}
		}
	}
	flushEntry()
	return entries
}

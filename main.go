package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

var wordLists = map[string][]string{
	"easy": {"the", "and", "for", "are", "but", "not", "you", "all", "can", "her",
		"was", "one", "our", "out", "day", "get", "has", "him", "his", "how",
		"man", "new", "now", "old", "see", "two", "way", "who", "boy", "did",
		"use", "may", "say", "she", "too", "any", "had", "let", "put", "end"},
	"medium": {"about", "after", "again", "below", "could", "every", "first", "found",
		"great", "house", "large", "learn", "never", "other", "place", "plant",
		"point", "right", "small", "sound", "spell", "still", "study", "their",
		"there", "these", "thing", "think", "three", "water", "where", "which",
		"world", "would", "write", "people", "number", "change", "animal", "follow", 
		"develop", "system", "program", "code", "logic", "minimal"},
	"hard": {"through", "between", "important", "children", "different", "following",
		"sentence", "without", "together", "something", "sometimes", "mountain",
		"question", "discover", "interest", "government", "experience", "character",
		"necessary", "beginning", "beautiful", "community", "development", "education",
		"everything", "information", "international", "organization", "understanding"},
}

type TypingTest struct {
	words        []string
	targetText   []rune
	typedText    []rune
	currentPos   int
	errors       int
	correctChars int
	keystrokes   int
	startTime    time.Time
	endTime      time.Time
	isFinished   bool
}

func initTest(numWords int, diff string) *TypingTest {
	words, ok := wordLists[diff]
	if !ok {
		words = wordLists["medium"]
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	selected := make([]string, numWords)
	for i := 0; i < numWords; i++ {
		selected[i] = words[rng.Intn(len(words))]
	}
	text := strings.Join(selected, " ")
	return &TypingTest{
		words:      selected,
		targetText: []rune(text),
		typedText:  make([]rune, 0, len(text)),
	}
}

func tbPrint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func (t *TypingTest) draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	w, h := termbox.Size()

	// Ghostty minimalist aesthetic: muted dim gray future text, bright white typed text.
	fgDim := termbox.Attribute(242) // macOS/Linux 256-color palette (Dark Gray)
	fgCorrect := termbox.ColorWhite | termbox.AttrBold
	fgError := termbox.ColorRed
	fgBg := termbox.ColorDefault // Will natively inherit the terminal's built-in background transparency

	maxWidth := 70
	if w < 75 {
		maxWidth = w - 4
	}

	startX := (w - maxWidth) / 2
	if startX < 2 {
		startX = 2
	}
	startY := (h / 2) - 3

	// Header - Super clean
	tbPrint(startX, startY-3, termbox.Attribute(246), fgBg, "go-typing")
	tbPrint(startX, startY-2, termbox.Attribute(238), fgBg, strings.Repeat("─", maxWidth))

	x, y := startX, startY

	// Render text body
	for i, char := range t.targetText {
		var fg termbox.Attribute = fgDim
		
		if i < t.currentPos {
			if t.typedText[i] == char {
				fg = fgCorrect
			} else {
				fg = fgError
				if char == ' ' {
					char = '_'
				}
			}
		}

		bg := fgBg
		// Sleek ghostty block cursor
		if i == t.currentPos && !t.isFinished {
			fg = termbox.ColorBlack
			bg = termbox.ColorWhite
		}

		termbox.SetCell(x, y, char, fg, bg)
		x++
		
		// Soft wrap
		if x >= startX+maxWidth && char == ' ' {
			x = startX
			y++
		}
	}

	// End of text cursor block
	if t.currentPos >= len(t.targetText) && !t.isFinished {
		termbox.SetCell(x, y, ' ', termbox.ColorBlack, termbox.ColorWhite)
	}

	// Footer + Live Stats
	statY := y + 4
	tbPrint(startX, statY-1, termbox.Attribute(238), fgBg, strings.Repeat("─", maxWidth))

	if !t.startTime.IsZero() || t.isFinished {
		elapsed := time.Since(t.startTime)
		if t.isFinished {
			elapsed = t.endTime.Sub(t.startTime)
		}
		
		wpm := 0.0
		if elapsed.Minutes() > 0 {
			wpm = (float64(t.correctChars) / 5.0) / elapsed.Minutes()
		}
		
		accuracy := 100.0
		if t.keystrokes > 0 {
			accuracy = float64(t.correctChars) / float64(t.keystrokes) * 100.0
		}

		// Clean numeric layout
		if t.isFinished {
			stats := fmt.Sprintf("WPM: %.1f    ACC: %.1f%%    ERR: %d    TIME: %.1fs", wpm, accuracy, t.errors, elapsed.Seconds())
			tbPrint(startX, statY, termbox.ColorCyan|termbox.AttrBold, fgBg, stats)
			tbPrint(startX, statY+2, termbox.Attribute(245), fgBg, "[r] restart   [q] quit")
		} else {
			prog := fmt.Sprintf("%d/%d", t.currentPos, len(t.targetText))
			wpmStr := fmt.Sprintf("WPM: %.0f", wpm)
			tbPrint(startX, statY, termbox.Attribute(245), fgBg, prog+" chars   "+wpmStr)
		}
	} else {
		// Waiting to start
		tbPrint(startX, statY, termbox.Attribute(242), fgBg, "start typing...   [esc] to quit")
	}

	termbox.Flush()
}

func main() {
	// Let arguments drive the configuration to avoid clunky UI Menus
	numWords := flag.Int("w", 30, "number of words (e.g. 50)")
	diff := flag.String("d", "medium", "difficulty (easy, medium, hard)")
	flag.Parse()

	err := termbox.Init()
	if err != nil {
		fmt.Println("Failed to initialize termbox:", err)
		os.Exit(1)
	}
	defer termbox.Close()

	// Essential for the sleek gray ghostty colors
	termbox.SetOutputMode(termbox.Output256)

	test := initTest(*numWords, *diff)
	test.draw()

mainloop:
	for {
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				break mainloop
			}

			if test.isFinished {
				if ev.Ch == 'q' || ev.Ch == 'Q' {
					break mainloop
				}
				if ev.Ch == 'r' || ev.Ch == 'R' || ev.Key == termbox.KeyEnter {
					test = initTest(*numWords, *diff)
					test.draw()
				}
				continue
			}

			// Handle deletions smoothly
			if ev.Key == termbox.KeyBackspace || ev.Key == termbox.KeyBackspace2 {
				if test.currentPos > 0 {
					test.currentPos--
					if test.typedText[len(test.typedText)-1] == test.targetText[test.currentPos] {
						test.correctChars--
					} else {
						test.errors--
					}
					test.typedText = test.typedText[:len(test.typedText)-1]
					test.draw()
				}
				continue
			}

			if ev.Key == termbox.KeyEnter {
				continue
			}

			char := ev.Ch
			if ev.Key == termbox.KeySpace {
				char = ' '
			}

			if char != 0 {
				if test.startTime.IsZero() {
					test.startTime = time.Now()
				}

				test.typedText = append(test.typedText, char)
				test.keystrokes++

				if char == test.targetText[test.currentPos] {
					test.correctChars++
				} else {
					test.errors++
				}

				test.currentPos++

				if test.currentPos >= len(test.targetText) {
					test.isFinished = true
					test.endTime = time.Now()
				}

				test.draw()
			}

		case termbox.EventResize:
			test.draw()
		}
	}
}

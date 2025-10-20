package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/nsf/termbox-go"
)

var wordLists = map[string][]string{
	"easy": {"the", "and", "for", "are", "but", "not", "you", "all", "can", "her",
		"was", "one", "our", "out", "day", "get", "has", "him", "his", "how",
		"man", "new", "now", "old", "see", "two", "way", "who", "boy", "did"},
	"medium": {"about", "after", "again", "below", "could", "every", "first", "found",
		"great", "house", "large", "learn", "never", "other", "place", "plant",
		"point", "right", "small", "sound", "spell", "still", "study", "their",
		"there", "these", "thing", "think", "three", "water"},
	"hard": {"through", "between", "important", "children", "different", "following",
		"sentence", "without", "together", "something", "sometimes", "mountain",
		"question", "discover", "interest", "government", "experience", "character",
		"necessary", "beginning"},
}

type TypingTest struct {
	numWords   int
	difficulty string
	words      []string
	text       string
	startTime  time.Time
	endTime    time.Time
	typedChars []rune
	errors     int
	currentPos int
}

func NewTypingTest(numWords int, difficulty string) *TypingTest {
	test := &TypingTest{
		numWords:   numWords,
		difficulty: difficulty,
		typedChars: make([]rune, 0),
	}
	test.generateWords()
	return test
}

func (t *TypingTest) generateWords() {
	words := wordLists[t.difficulty]
	t.words = make([]string, t.numWords)

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < t.numWords; i++ {
		t.words[i] = words[rand.Intn(len(words))]
	}
	t.text = strings.Join(t.words, " ")
}

func (t *TypingTest) calculateAccuracy() float64 {
	if len(t.typedChars) == 0 {
		return 100.0
	}
	correct := len(t.typedChars) - t.errors
	return (float64(correct) / float64(len(t.typedChars))) * 100.0
}

func (t *TypingTest) calculateWPM(elapsed time.Duration) float64 {
	if elapsed.Seconds() == 0 {
		return 0
	}
	charsTyped := len(t.typedChars)
	minutes := elapsed.Minutes()
	return (float64(charsTyped) / 5.0) / minutes
}

func (t *TypingTest) displayText() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	headerY := 0
	drawText(0, headerY, "============================================================", termbox.ColorWhite, termbox.ColorDefault)
	drawText(0, headerY+1, center("TYPING TEST - Type the text below", 60), termbox.ColorWhite|termbox.AttrBold, termbox.ColorDefault)
	drawText(0, headerY+2, "============================================================", termbox.ColorWhite, termbox.ColorDefault)

	textRunes := []rune(t.text)
	startY := 5
	x, y := 0, startY

	for i, char := range textRunes {
		var fg termbox.Attribute

		if i < t.currentPos {
			if i < len(t.typedChars) {
				if t.typedChars[i] == char {
					fg = termbox.ColorGreen
				} else {
					fg = termbox.ColorRed
				}
			}
		} else if i == t.currentPos {
			fg = termbox.ColorYellow | termbox.AttrBold
		} else {
			fg = termbox.ColorWhite | termbox.AttrDim
		}

		termbox.SetCell(x, y, char, fg, termbox.ColorDefault)
		x++

		if x >= 60 {
			x = 0
			y++
		}
	}

	if !t.startTime.IsZero() {
		elapsed := time.Since(t.startTime)
		wpm := t.calculateWPM(elapsed)
		accuracy := t.calculateAccuracy()
		stats := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%% | Errors: %d", wpm, accuracy, t.errors)
		drawText(0, y+2, stats, termbox.ColorCyan|termbox.AttrBold, termbox.ColorDefault)
	}
	termbox.Flush()
}

func (t *TypingTest) showResults() {
	elapsed := t.endTime.Sub(t.startTime)
	wpm := t.calculateWPM(elapsed)
	accuracy := t.calculateAccuracy()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	y := 2
	drawText(0, y, "============================================================", termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(0, y, center("Results", 60), termbox.ColorWhite|termbox.AttrBold, termbox.ColorDefault)
	y++
	drawText(0, y, "============================================================", termbox.ColorWhite, termbox.ColorDefault)
	y += 2

	drawText(2, y, fmt.Sprintf("Words Per Minute (WPM): %.2f", wpm), termbox.ColorGreen, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Accuracy: %.2f%%", accuracy), termbox.ColorGreen, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Time: %.2fs", elapsed.Seconds()), termbox.ColorGreen, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Total Characters: %d", len(t.typedChars)), termbox.ColorGreen, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Errors: %d", t.errors), termbox.ColorRed, termbox.ColorDefault)
	y += 2

	drawText(0, y, "============================================================", termbox.ColorWhite, termbox.ColorDefault)
	y += 2
	drawText(0, y, "Press any key to continue...", termbox.ColorYellow, termbox.ColorDefault)

	termbox.Flush()

	keyboard.GetSingleKey()
}

func (t *TypingTest) run() error {
	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	y := 2

	drawText(0, y, "============================================================", termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(0, y, center("TYPING TEST", 60), termbox.ColorWhite|termbox.AttrBold, termbox.ColorDefault)
	y++
	drawText(0, y, "============================================================", termbox.ColorWhite, termbox.ColorDefault)
	y += 2

	drawText(0, y, fmt.Sprintf("Difficulty: %s | Words: %d", strings.ToUpper(t.difficulty), t.numWords), termbox.ColorCyan, termbox.ColorDefault)
	y += 2

	drawText(0, y, "Instructions:", termbox.ColorYellow, termbox.ColorDefault)
	y++
	drawText(0, y, "- Type the words exactly as shown", termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(0, y, "- Press ENTER when done or to skip", termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(0, y, "- Press ESC to quit", termbox.ColorWhite, termbox.ColorDefault)
	y += 2

	drawText(0, y, "Press SPACE to start...", termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault)
	termbox.Flush()

	if err := keyboard.Open(); err != nil {
		return err
	}
	defer keyboard.Close()

	for {
		_, key, err := keyboard.GetKey()
		if err != nil {
			return err
		}
		if key == keyboard.KeySpace || key == keyboard.KeyEnter {
			break
		}
		if key == keyboard.KeyEsc {
			return nil
		}
	}

	t.displayText()
	t.startTime = time.Now()

	textRunes := []rune(t.text)

	for t.currentPos < len(textRunes) {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return err
		}

		if key == keyboard.KeyEsc {
			return nil
		}
		if key == keyboard.KeyEnter {
			break
		}
		if key == keyboard.KeyBackspace {
			if t.currentPos > 0 && len(t.typedChars) > 0 {
				t.currentPos--
				lastChar := t.typedChars[len(t.typedChars)-1]
				if lastChar != textRunes[t.currentPos] {
					t.errors--
				}
				t.typedChars = t.typedChars[:len(t.typedChars)-1]
				t.displayText()
			}
			continue
		}

		if key == keyboard.KeySpace {
			char = ' '
		}

		if char != 0 {
			t.typedChars = append(t.typedChars, char)

			if char != textRunes[t.currentPos] {
				t.errors++
			}

			t.currentPos++
			t.displayText()
		}
	}

	t.endTime = time.Now()
	termbox.Close()

	if err := termbox.Init(); err != nil {
		return err
	}

	t.showResults()

	return nil
}

func drawText(x, y int, text string, fg, bg termbox.Attribute) {
	for i, ch := range text {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}

func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := (width - len(s)) / 2
	return strings.Repeat(" ", padding) + s
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println(center("TYPING PRACTICE", 60))
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("\nSelect difficulty:")
		fmt.Println("1. Easy (short words)")
		fmt.Println("2. Medium (moderate words)")
		fmt.Println("3. Hard (longer words)")

		fmt.Print("\nEnter choice (1-3, default 2): ")
		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)

		difficultyMap := map[string]string{
			"1": "easy",
			"2": "medium",
			"3": "hard",
		}

		difficulty, ok := difficultyMap[choiceStr]
		if !ok {
			difficulty = "medium"
		}

		fmt.Print("Number of words (default 30): ")
		numWordsStr, _ := reader.ReadString('\n')
		numWordsStr = strings.TrimSpace(numWordsStr)

		numWords := 30
		if n, err := strconv.Atoi(numWordsStr); err == nil && n > 0 {
			numWords = n
		}

		test := NewTypingTest(numWords, difficulty)

		if err := test.run(); err != nil {
			fmt.Printf("\nError: %v\n", err)
			break
		}

		fmt.Print("\nTry again? (y/n): ")
		again, _ := reader.ReadString('\n')
		again = strings.TrimSpace(strings.ToLower(again))

		if again != "y" {
			fmt.Println("\nGoodbye!")
			break
		}
	}
}

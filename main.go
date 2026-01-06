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
	"easy": {"the", "and", "for", "are", "but", "not", "you", "all", "can", "her","you",
		"was", "one", "our", "out", "day", "get", "has", "him", "his", "how",
		"man", "new", "now", "old", "see", "two", "way", "who", "boy", "did",
		"use", "may", "say", "she", "too", "any", "had", "let", "put", "end"},
	"medium": {"about", "after", "again", "below", "could", "every", "first", "found",
		"great", "house", "large", "learn", "never", "other", "place", "plant",
		"point", "right", "small", "sound", "spell", "still", "study", "their",
		"there", "these", "thing", "think", "three", "water", "where", "which",
		"world", "would", "write", "people", "number", "change", "animal", "follow"},
	"hard": {"through", "between", "important", "children", "different", "following",
		"sentence", "without", "together", "something", "sometimes", "mountain",
		"question", "discover", "interest", "government", "experience", "character",
		"necessary", "beginning", "beautiful", "community", "development", "education",
		"everything", "information", "international", "organization", "understanding"},
}

type TypingTest struct {
	numWords     int
	difficulty   string
	words        []string
	text         string
	startTime    time.Time
	endTime      time.Time
	typedChars   []rune
	errors       int
	currentPos   int
	maxWPM       float64
	keystrokes   int
	correctChars int
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
	words, ok := wordLists[t.difficulty]
	if !ok {
		words = wordLists["medium"]
	}

	t.words = make([]string, t.numWords)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < t.numWords; i++ {
		t.words[i] = words[rng.Intn(len(words))]
	}
	t.text = strings.Join(t.words, " ")
}

func (t *TypingTest) calculateAccuracy() float64 {
	if t.keystrokes == 0 {
		return 100.0
	}
	return (float64(t.correctChars) / float64(t.keystrokes)) * 100.0
}

func (t *TypingTest) calculateWPM(elapsed time.Duration) float64 {
	if elapsed.Seconds() == 0 {
		return 0
	}
	minutes := elapsed.Minutes()
	// Standard WPM calculation: (correct characters / 5) / minutes
	return (float64(t.correctChars) / 5.0) / minutes
}

func (t *TypingTest) calculateRawWPM(elapsed time.Duration) float64 {
	if elapsed.Seconds() == 0 {
		return 0
	}
	minutes := elapsed.Minutes()
	return (float64(len(t.typedChars)) / 5.0) / minutes
}

func (t *TypingTest) displayText() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	headerY := 0
	drawText(0, headerY, strings.Repeat("=", 70), termbox.ColorWhite, termbox.ColorDefault)
	drawText(0, headerY+1, center("TYPING TEST - Type the text below", 70), termbox.ColorWhite|termbox.AttrBold, termbox.ColorDefault)
	drawText(0, headerY+2, strings.Repeat("=", 70), termbox.ColorWhite, termbox.ColorDefault)

	textRunes := []rune(t.text)
	startY := 5
	x, y := 2, startY
	const lineWidth = 66

	for i, char := range textRunes {
		var fg termbox.Attribute

		if i < t.currentPos {
			if i < len(t.typedChars) {
				if t.typedChars[i] == char {
					fg = termbox.ColorGreen
				} else {
					fg = termbox.ColorRed | termbox.AttrBold
				}
			}
		} else if i == t.currentPos {
			fg = termbox.ColorYellow | termbox.AttrBold | termbox.AttrUnderline
		} else {
			fg = termbox.ColorWhite
		}

		termbox.SetCell(x, y, char, fg, termbox.ColorDefault)
		x++

		if x >= lineWidth+2 {
			x = 2
			y++
		}
	}

	if !t.startTime.IsZero() {
		elapsed := time.Since(t.startTime)
		wpm := t.calculateWPM(elapsed)
		rawWPM := t.calculateRawWPM(elapsed)
		accuracy := t.calculateAccuracy()

		if wpm > t.maxWPM {
			t.maxWPM = wpm
		}

		statsY := y + 3
		drawText(0, statsY, strings.Repeat("─", 70), termbox.ColorWhite|termbox.AttrDim, termbox.ColorDefault)

		stats1 := fmt.Sprintf("WPM: %.1f | Raw: %.1f | Peak: %.1f", wpm, rawWPM, t.maxWPM)
		stats2 := fmt.Sprintf("Accuracy: %.1f%% | Errors: %d | Progress: %d/%d",
			accuracy, t.errors, t.currentPos, len(textRunes))

		drawText(2, statsY+1, stats1, termbox.ColorCyan|termbox.AttrBold, termbox.ColorDefault)
		drawText(2, statsY+2, stats2, termbox.ColorMagenta, termbox.ColorDefault)

		drawText(0, statsY+3, strings.Repeat("─", 70), termbox.ColorWhite|termbox.AttrDim, termbox.ColorDefault)
		drawText(2, statsY+4, "ESC: Quit | BACKSPACE: Delete | ENTER: Skip", termbox.ColorYellow|termbox.AttrDim, termbox.ColorDefault)
	}

	termbox.Flush()
}

func (t *TypingTest) showResults() {
	elapsed := t.endTime.Sub(t.startTime)
	wpm := t.calculateWPM(elapsed)
	rawWPM := t.calculateRawWPM(elapsed)
	accuracy := t.calculateAccuracy()

	// Calculate grade based on WPM and accuracy
	grade := t.calculateGrade(wpm, accuracy)

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	y := 2
	drawText(0, y, strings.Repeat("=", 70), termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(0, y, center("RESULTS", 70), termbox.ColorWhite|termbox.AttrBold, termbox.ColorDefault)
	y++
	drawText(0, y, strings.Repeat("=", 70), termbox.ColorWhite, termbox.ColorDefault)
	y += 2

	// Performance grade
	gradeColor := termbox.ColorGreen
	if grade == "B" || grade == "C" {
		gradeColor = termbox.ColorYellow
	} else if grade == "D" || grade == "F" {
		gradeColor = termbox.ColorRed
	}

	drawText(2, y, fmt.Sprintf("Grade: %s", grade), gradeColor|termbox.AttrBold, termbox.ColorDefault)
	y += 2

	// Main stats
	drawText(2, y, fmt.Sprintf("Net WPM:        %.2f", wpm), termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Raw WPM:        %.2f", rawWPM), termbox.ColorCyan, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Peak WPM:       %.2f", t.maxWPM), termbox.ColorMagenta, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Accuracy:       %.2f%%", accuracy), getAccuracyColor(accuracy), termbox.ColorDefault)
	y += 2

	// Detailed stats
	drawText(2, y, fmt.Sprintf("Time Elapsed:   %.2fs", elapsed.Seconds()), termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Total Keystrokes: %d", t.keystrokes), termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Correct:        %d", t.correctChars), termbox.ColorGreen, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Errors:         %d", t.errors), termbox.ColorRed, termbox.ColorDefault)
	y++
	drawText(2, y, fmt.Sprintf("Difficulty:     %s", strings.ToUpper(t.difficulty)), termbox.ColorYellow, termbox.ColorDefault)
	y += 2

	drawText(0, y, strings.Repeat("=", 70), termbox.ColorWhite, termbox.ColorDefault)
	y += 2
	drawText(0, y, "Press any key to continue...", termbox.ColorYellow|termbox.AttrBold, termbox.ColorDefault)

	termbox.Flush()
	keyboard.GetSingleKey()
}

func (t *TypingTest) calculateGrade(wpm, accuracy float64) string {
	// Weighted score: WPM matters more, but accuracy is crucial
	score := wpm * (accuracy / 100.0)

	if score >= 50 && accuracy >= 95 {
		return "A+"
	} else if score >= 40 && accuracy >= 90 {
		return "A"
	} else if score >= 30 && accuracy >= 85 {
		return "B"
	} else if score >= 20 && accuracy >= 80 {
		return "C"
	} else if score >= 10 && accuracy >= 70 {
		return "D"
	}
	return "F"
}

func getAccuracyColor(accuracy float64) termbox.Attribute {
	if accuracy >= 95 {
		return termbox.ColorGreen | termbox.AttrBold
	} else if accuracy >= 85 {
		return termbox.ColorYellow
	}
	return termbox.ColorRed
}

func (t *TypingTest) run() error {
	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	y := 2

	drawText(0, y, strings.Repeat("=", 70), termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(0, y, center("TYPING TEST", 70), termbox.ColorWhite|termbox.AttrBold, termbox.ColorDefault)
	y++
	drawText(0, y, strings.Repeat("=", 70), termbox.ColorWhite, termbox.ColorDefault)
	y += 2

	drawText(0, y, fmt.Sprintf("Difficulty: %s | Words: %d", strings.ToUpper(t.difficulty), t.numWords), termbox.ColorCyan|termbox.AttrBold, termbox.ColorDefault)
	y += 2

	drawText(0, y, "Instructions:", termbox.ColorYellow|termbox.AttrBold, termbox.ColorDefault)
	y++
	drawText(2, y, "• Type the words exactly as shown", termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(2, y, "• Green = correct, Red = incorrect", termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(2, y, "• Use BACKSPACE to correct mistakes", termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(2, y, "• Press ENTER to finish early", termbox.ColorWhite, termbox.ColorDefault)
	y++
	drawText(2, y, "• Press ESC to quit anytime", termbox.ColorWhite, termbox.ColorDefault)
	y += 2

	drawText(0, y, "Press SPACE to start...", termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault)
	termbox.Flush()

	if err := keyboard.Open(); err != nil {
		return err
	}
	defer keyboard.Close()

	// Wait for space or enter to start
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

		if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
			if t.currentPos > 0 && len(t.typedChars) > 0 {
				t.currentPos--
				lastChar := t.typedChars[len(t.typedChars)-1]

				// Update error count and correct chars
				if lastChar != textRunes[t.currentPos] {
					t.errors--
				} else {
					t.correctChars--
				}

				t.typedChars = t.typedChars[:len(t.typedChars)-1]
				t.keystrokes-- // Backspace undoes a keystroke
				t.displayText()
			}
			continue
		}

		if key == keyboard.KeySpace {
			char = ' '
		}

		if char != 0 {
			t.typedChars = append(t.typedChars, char)
			t.keystrokes++

			if char == textRunes[t.currentPos] {
				t.correctChars++
			} else {
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

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		clearScreen()
		fmt.Println(strings.Repeat("=", 70))
		fmt.Println(center("TYPING PRACTICE", 70))
		fmt.Println(strings.Repeat("=", 70))
		fmt.Println("\nSelect difficulty:")
		fmt.Println("  1. Easy   - Short common words (3-4 letters)")
		fmt.Println("  2. Medium - Moderate words (5-6 letters)")
		fmt.Println("  3. Hard   - Longer words (7+ letters)")

		fmt.Print("\nEnter choice (1-3, default 2): ")
		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)

		difficultyMap := map[string]string{
			"1": "easy",
			"2": "medium",
			"3": "hard",
			"":  "medium",
		}

		difficulty, ok := difficultyMap[choiceStr]
		if !ok {
			difficulty = "medium"
		}

		fmt.Print("Number of words (10-100, default 30): ")
		numWordsStr, _ := reader.ReadString('\n')
		numWordsStr = strings.TrimSpace(numWordsStr)

		numWords := 30
		if n, err := strconv.Atoi(numWordsStr); err == nil && n > 0 {
			if n > 100 {
				numWords = 100
			} else if n < 10 {
				numWords = 10
			} else {
				numWords = n
			}
		}

		test := NewTypingTest(numWords, difficulty)

		if err := test.run(); err != nil {
			fmt.Printf("\nError: %v\n", err)
			break
		}

		fmt.Print("\nTry again? (y/n, default y): ")
		again, _ := reader.ReadString('\n')
		again = strings.TrimSpace(strings.ToLower(again))

		if again == "n" || again == "no" {
			clearScreen()
			fmt.Println("\n" + center("Thank you for practicing!", 70))
			fmt.Println(center("Keep typing to improve your speed!", 70) + "\n")
			break
		}
	}
}

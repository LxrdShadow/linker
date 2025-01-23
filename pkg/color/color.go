package color

import (
	"fmt"
	"io"
)

const (
	GRAY   = 30
	RED    = 31
	GREEN  = 32
	YELLOW = 33
	BLUE   = 34
	PINK   = 35
	CYAN   = 36
)

// Wraps a string with ANSI escape codes for coloring
func colorize(color int, text string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, text)
}

// Core formatting function
func ColorFormat(color int, format string, a ...any) string {
	formattedText := fmt.Sprintf(format, a...)
	return colorize(color, formattedText)
}

func Sprint(color int, a ...any) string {
	return colorize(color, fmt.Sprint(a...))
}

func Sprintf(color int, format string, a ...any) string {
	return ColorFormat(color, format, a...)
}

func Sprintln(color int, a ...any) string {
	return colorize(color, fmt.Sprintln(a...))
}

func Fprint(w io.Writer, color int, a ...any) {
	fmt.Fprint(w, Sprint(color, a...))
}

func Fprintf(w io.Writer, color int, format string, a ...any) {
	fmt.Fprint(w, Sprintf(color, format, a...))
}

func Fprintln(w io.Writer, color int, a ...any) {
	fmt.Fprint(w, Sprintln(color, a...))
}

func Print(color int, a ...any) {
	fmt.Print(Sprint(color, a...))
}

func Printf(format string, color int, a ...any) {
	fmt.Print(Sprintf(color, format, a...))
}

func Println(color int, a ...any) {
	fmt.Print(Sprintln(color, a...))
}

package color

import (
	"fmt"
	"io"
	"strings"
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

const Base = "\x1b[%dm%s\x1b[0m"

func joinBaseFormat(format string) string {
	return "\x1b[%dm" + format + "\x1b[0m"
}

func ColorFormat(format string, color int, a ...string) string {
	var builder strings.Builder
	combinedFmt := Base

	if format != "" && format != "\n" {
		combinedFmt = joinBaseFormat(format)
	}

	for i, str := range a {
		builder.WriteString(str)
		if i < len(a)-1 {
			builder.WriteString(" ")
		}
	}

	return fmt.Sprintf(combinedFmt, color, builder.String())
}

func Sprint(color int, a ...string) string {
	return ColorFormat("", color, a...)
}

func Sprintf(format string, color int, a ...string) string {
	return ColorFormat(format, color, a...)
}

func Sprintln(color int, a ...string) string {
	return ColorFormat("\n", color, a...)
}

func Fprint(w io.Writer, color int, a ...string) {
	fmt.Fprint(w, ColorFormat("", color, a...))
}

func Fprintf(w io.Writer, format string, color int, a ...string) {
	fmt.Fprint(w, ColorFormat(format, color, a...))
}

func Fprintln(w io.Writer, color int, a ...string) {
	fmt.Fprint(w, ColorFormat("\n", color, a...))
}

func Print(color int, a ...string) {
	fmt.Print(ColorFormat("", color, a...))
}

func Printf(format string, color int, a ...string) {
	fmt.Print(ColorFormat(format, color, a...))
}

func Println(color int, a ...string) {
	fmt.Print(ColorFormat("\n", color, a...))
}

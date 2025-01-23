package log

import (
	"fmt"
	"os"

	"github.com/LxrdShadow/linker/pkg/color"
)

var (
	errorPrefix   = color.Sprint(color.RED, "[ERROR]: ")
	warningPrefix = color.Sprint(color.YELLOW, "[WARNING]: ")
	infoPrefix    = color.Sprint(color.BLUE, "[INFO]: ")
	successPrefix = color.Sprint(color.GREEN, "[SUCCESS]: ")
)

func Error(msg any) {
	fmt.Fprint(os.Stderr, errorPrefix, msg)
}

func Errorf(format string, msg any) {
	fmt.Fprint(os.Stderr, errorPrefix)
	fmt.Fprintf(os.Stderr, format, msg)
}

func Warning(msg any) {
	fmt.Print(warningPrefix, msg)
}

func Warningf(format string, msg any) {
	fmt.Print(warningPrefix)
	fmt.Printf(format, msg)
}

func Info(msg any) {
	fmt.Print(infoPrefix, msg)
}

func Infof(format string, msg any) {
	fmt.Print(infoPrefix)
	fmt.Printf(format, msg)
}

func Success(msg any) {
	fmt.Print(successPrefix, msg)
}

func Successf(format string, msg any) {
	fmt.Print(successPrefix)
	fmt.Printf(format, msg)
}

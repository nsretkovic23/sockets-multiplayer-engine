package helpers

import (
	"fmt"

	"github.com/fatih/color"
)

const (
	INFO = "\U0001F6C8"
)

func PrintRed(message string) {
	red := color.New(color.FgHiRed).SprintfFunc()
	fmt.Println(red(message))
}

func PrintYellow(message string) {
	yellow := color.New(color.FgHiYellow).SprintfFunc()
	fmt.Println(yellow(message))
}

func PrintBlue(message string) {
	blue := color.New(color.FgHiBlue).SprintfFunc()
	fmt.Println(blue(message))
}

func PrintGreen(message string) {
	green := color.New(color.FgHiGreen).SprintfFunc()
	fmt.Println(green(message))
}

func PrintInfo(message string) {
	info := color.New(color.FgCyan).SprintfFunc()
	fmt.Println(info(INFO + " " + message))
}

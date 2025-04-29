package console

import "fmt"

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
)

var NoColor bool

func SetNoColor(v bool) {
	NoColor = v
}

func PrintSuccess(msg string) {
	if NoColor {
		fmt.Println(msg)
	} else {
		fmt.Printf("%s%s%s\n", colorGreen, msg, colorReset)
	}
}

func PrintFailure(msg string) {
	if NoColor {
		fmt.Println(msg)
	} else {
		fmt.Printf("%s%s%s\n", colorRed, msg, colorReset)
	}
}

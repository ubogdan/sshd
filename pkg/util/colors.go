package util

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const (
	SystemMessageFormat = "\033[1;90m"

	// ColorReset resets the color
	ColorReset = "\033[0m"

	// ColorBold makes the following text bold
	ColorBold = "\033[1m"

	// ColorDim dims the following text
	ColorDim = "\033[2m"

	// Italic makes the following text italic
	Italic = "\033[3m"

	// Underline underlines the following text
	Underline = "\033[4m"

	// ColorBlink blinks the following text
	ColorBlink = "\033[5m"

	// ColorInvert inverts the following text
	ColorInvert = "\033[7m"
)

var colors = []string{"31", "32", "33", "34", "35", "36", "37", "91", "92", "93", "94", "95", "96", "97"}

// deColor is used for removing ANSI Escapes
var deColor = regexp.MustCompile("\033\\[[\\d;]+m")

// DeColorString removes all color from the given string
func DeColorString(s string) string {
	s = deColor.ReplaceAllString(s, "")
	return s
}

func randomReadableColor() int {
	for {
		i := rand.Intn(256)
		if (16 <= i && i <= 18) || (232 <= i && i <= 237) {
			// Remove the ones near black, this is kinda sadpanda.
			continue
		}
		return i
	}
}

// RandomColor256 returns a random (of 256) color
func RandomColor256() string {
	return fmt.Sprintf("\033[38;05;%dm", randomReadableColor())
}

// RandomColor returns a random color
func RandomColor() string {
	return colors[rand.Intn(len(colors))]
}

// ColorString returns a message in the given color
func ColorString(color string, msg string) string {
	return ColorBold + "\033[" + color + "m" + msg + ColorReset
}

// init initializes the random seed
func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// ContinuousFormat is a horrible hack to "continue" the previous string color
// and format after a RESET has been encountered.
//
// This is not HTML where you can just do a </style> to resume your previous formatting!
func ContinuousFormat(format string, str string) string {
	return SystemMessageFormat + strings.Replace(str, ColorReset, format, -1) + ColorReset
}

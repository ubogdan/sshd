package util

import "regexp"

func NormalString(in string) string {
	reStripText := regexp.MustCompile("[^0-9A-Za-z_.-]")
	return reStripText.ReplaceAllString(in, "")
}

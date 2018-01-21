package yajirobe

import (
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
)

// encoding

func toUtf8(str string) string {
	s, _ := japanese.ShiftJIS.NewDecoder().String(str)
	return s
}

func toSjis(str string) string {
	s, _ := japanese.ShiftJIS.NewEncoder().String(str)
	return s
}

// strings

func parseSeparatedInt(s string) int64 {
	s = strings.Replace(s, ",", "", -1)
	s = regexp.MustCompile(`\d+`).FindString(s)
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

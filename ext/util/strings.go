package util

import (
	"strings"
	"unicode"
)

//UnderScoreCase Convert string to lower case name.like: "UserId" -> "user_id"
func UnderScoreCase(n string) string {
	s := ([]rune)(n)
	var newBytes []rune
	for idx, ch := range s {
		//if ch >= 'A' && ch <= 'Z' {
		if unicode.IsUpper(ch) {
			//if idx != 0 && (idx < len(s)-1) && s[idx+1] >= 'a' && s[idx+1] <= 'z' {
			if idx != 0 && (idx < len(s)-1) && unicode.IsLower(s[idx+1]) {
				newBytes = append(newBytes, '_')
			}
			newBytes = append(newBytes, unicode.ToLower(rune(ch)))
		} else {
			newBytes = append(newBytes, ch)
		}
	}
	return string(newBytes)
}

func MiddleScoreCase(n string) string {
	s := ([]rune)(n)
	var newBytes []rune
	for idx, ch := range s {
		//if ch >= 'A' && ch <= 'Z' {
		if unicode.IsUpper(ch) {
			//if idx != 0 && (idx < len(s)-1) && s[idx+1] >= 'a' && s[idx+1] <= 'z' {
			if idx != 0 && (idx < len(s)-1) && unicode.IsLower(s[idx+1]) {
				newBytes = append(newBytes, '-')
			}
			newBytes = append(newBytes, unicode.ToLower(rune(ch)))
		} else {
			newBytes = append(newBytes, ch)
		}
	}
	return string(newBytes)
}

//SmallCamelCase String to small CameCase name
func SmallCamelCase(name string) string {
	if len(name) == 0 {
		return name
	}
	runes := ([]rune)(name)
	var newRunes []rune
	newRunes = append(newRunes, unicode.ToLower(runes[0]))
	var doUpper = false
	for _, r := range runes[1:] {
		if r == '_' {
			doUpper = true
			continue
		}
		if doUpper {
			newRunes = append(newRunes, unicode.ToUpper(r))
			doUpper = false
		} else {
			newRunes = append(newRunes, r)
		}
	}
	return string(newRunes)
}

func BigCamelCase(name string) string {
	if len(name) == 0 {
		return name
	}
	runes := ([]rune)(name)
	var newRunes []rune
	newRunes = append(newRunes, unicode.ToUpper(runes[0]))
	var doUpper = false
	for _, r := range runes[1:] {
		if r == '_' {
			doUpper = true
			continue
		}
		if doUpper {
			newRunes = append(newRunes, unicode.ToUpper(r))
			doUpper = false
		} else {
			newRunes = append(newRunes, r)
		}
	}

	return string(newRunes)
}

func Quote(name, quoteStr string) string {
	if strings.HasPrefix(name, quoteStr) && strings.HasSuffix(name, quoteStr) {
		return name
	}
	quoteBytes := []rune(quoteStr)
	b := []rune(name)
	buf := make([]rune, len(quoteBytes)*2+len(b))
	copy(buf, quoteBytes)
	copy(buf[len(quoteBytes):], b)
	copy(buf[len(quoteBytes)+len(b):], quoteBytes)
	return string(buf)
}

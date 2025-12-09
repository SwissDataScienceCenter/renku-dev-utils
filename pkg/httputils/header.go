package httputils

import (
	"fmt"
)

var tCharRunes = map[rune]struct{}{}

func ParseTChar(value string) (res, rem string, err error) {
	if value == "" {
		return "", "", fmt.Errorf("empty string")
	}
	first := rune(value[0])
	if _, found := tCharRunes[first]; found {
		return value[:1], value[1:], nil
	}
	return "", "", fmt.Errorf("not a tchar")
}

func ParseToken(value string) (res, rem string, err error) {
	res, rem, err = ParseTChar(value)
	if err != nil {
		return "", "", fmt.Errorf("not a token")
	}
	for {
		nres, nrem, nerr := ParseTChar(rem)
		if nerr != nil {
			break
		}
		res = res + nres
		rem = nrem
	}
	return res, rem, nil
}

func init() {
	initTCharRunes()
}

func initTCharRunes() {
	tCharRunes['!'] = struct{}{}
	tCharRunes['#'] = struct{}{}
	tCharRunes['$'] = struct{}{}
	tCharRunes['%'] = struct{}{}
	tCharRunes['&'] = struct{}{}
	tCharRunes['\''] = struct{}{}
	tCharRunes['*'] = struct{}{}
	tCharRunes['+'] = struct{}{}
	tCharRunes['-'] = struct{}{}
	tCharRunes['.'] = struct{}{}
	tCharRunes['^'] = struct{}{}
	tCharRunes['_'] = struct{}{}
	tCharRunes['`'] = struct{}{}
	tCharRunes['|'] = struct{}{}
	tCharRunes['~'] = struct{}{}
	for r := '0'; r <= '9'; r++ {
		tCharRunes[r] = struct{}{}
	}
	for r := 'A'; r <= 'Z'; r++ {
		tCharRunes[r] = struct{}{}
	}
	for r := 'a'; r <= 'z'; r++ {
		tCharRunes[r] = struct{}{}
	}
}

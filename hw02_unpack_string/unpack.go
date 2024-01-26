package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrInvalidString       = errors.New("invalid string")
	ErrUnusedEscaping      = errors.New("unused escaping")
	ErrInvalidEscaping     = errors.New("invalid escaping")
	ErrDigitIsNotExpected  = errors.New("digit is not expected")
	ErrConvertingToNumeric = errors.New("invalid converting to numeric")
)

func Unpack(input string) (string, error) {
	// Place your code here.

	if len(input) == 0 {
		return "", nil
	}

	var sb strings.Builder
	var unpackedRune rune

	startUnpackRune := true
	escapingDetected := false

	for _, rune := range input {
		switch {
		case escapingDetected:
			if !(unicode.IsDigit(rune) || string(rune) == `\`) {
				return "", ErrInvalidEscaping
			}
			unpackedRune = rune
			escapingDetected = false
		case !escapingDetected && string(rune) == `\`:
			escapingDetected = true
			if unpackedRune != 0 {
				sb.WriteRune(unpackedRune)
			}
			unpackedRune = 0
			startUnpackRune = false
		case startUnpackRune:
			if unicode.IsDigit(rune) {
				return "", ErrDigitIsNotExpected
			}
			unpackedRune = rune
			startUnpackRune = false
		case !startUnpackRune && string(rune) == "0":
			startUnpackRune = true
		case !startUnpackRune && unicode.IsDigit(rune):
			digit, err := strconv.Atoi(string(rune))
			if err != nil {
				return "", ErrConvertingToNumeric
			}
			sb.WriteString(strings.Repeat(string(unpackedRune), digit))
			startUnpackRune = true
		case !startUnpackRune && !unicode.IsDigit(rune):
			sb.WriteRune(unpackedRune)
			unpackedRune = rune
		default:
			return "", ErrInvalidString
		}
	}

	if escapingDetected {
		return "", ErrUnusedEscaping
	}

	if !startUnpackRune {
		sb.WriteRune(unpackedRune)
	}
	return sb.String(), nil
}

package driver

import (
	"math"
	"strings"
)

func (d *Driver) EncodeNumberToAlphabet(number int64) string {
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	base := int64(len(alphabet))
	var encoded strings.Builder

	for number > 0 {
		remainder := number % base
		number = number / base
		encoded.WriteByte(alphabet[int64(math.Max(float64(remainder-1), 0))])
	}

	encodedStr := encoded.String()
	var reversed strings.Builder

	for i := len(encodedStr) - 1; i >= 0; i-- {
		reversed.WriteByte(encodedStr[i])
	}

	return reversed.String()
}
package driver

import "fmt"

func (d *Driver) EncodeNumberToAlphabet(number int) string {
	var result string
	for number > 0 {
		number--
		result = fmt.Sprintf("%c", number%26+97) + result
		number /= 26
	}
	return result
}

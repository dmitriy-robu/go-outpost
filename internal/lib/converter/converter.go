package converter

import "strconv"

func ConvertAmountFloatToInt(amount float64) int {
	return int(amount * 100)
}

func ConvertAmountIntToSting(amount int) string {
	return strconv.Itoa(amount / 100)
}

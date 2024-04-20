package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

func ValidateAmount(amount string) (string, bool) { //, p_type_selected string
	//returns "" or a fixed result (.00000)
	is_valid := false
	new_amt_str := ""
	if len(amount) > 1 {
		before, after, found := strings.Cut(amount, ".")
		if found {
			slice_end := len(after)
			if slice_end > 5 {
				slice_end = 5
			}
			new_amt_str = before + "." + after[0:slice_end]
		}
	}

	if new_amt_str != "" {
		amount = new_amt_str
	}
	//fmt.Println(p_type_selected)
	if ConvertToAtomicUnits(amount) > 0 { /*|| (ConvertToAtomicUnits(amount) >= 0 && p_type_selected == "token")*/
		is_valid = true
	}

	return new_amt_str, is_valid
}

func ConvertToAtomicUnits(amount string) int {
	deri, _ := strconv.ParseFloat(amount, 64)
	deri = 100000 * float64(deri)
	return int(deri)
}

func ConvertToDeroUnits(amount int) string {
	if amount == 0 {
		return "0"
	}
	dero := float64(amount) * float64(.00001)
	strconv.FormatFloat(dero, 'f', -1, 64)
	s := fmt.Sprintf("%.5f", dero)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")

}

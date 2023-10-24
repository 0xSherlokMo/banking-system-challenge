package calculator

import (
	"math"
	"math/big"
	"strconv"
)

const (
	decimalPlaces = 1e6
)

func PreciseAdd(first float64, second float64) float64 {
	result, _ := strconv.ParseFloat(
		new(big.Float).
			Add(
				new(big.Float).SetFloat64(first),
				new(big.Float).SetFloat64(second)).
			String(),
		64)
	result = math.Round(result*decimalPlaces) / decimalPlaces
	return result
}

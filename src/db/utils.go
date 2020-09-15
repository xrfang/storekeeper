package db

import (
	"math"
	"strconv"
)

func float(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return math.Round(f*100) / 100, nil
}

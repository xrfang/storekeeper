package db

import (
	"fmt"
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

func fval(v interface{}) float64 {
	switch v.(type) {
	case float64:
		return v.(float64)
	case int:
		return float64(v.(int))
	case int64:
		return float64(v.(int64))
	}
	panic(fmt.Errorf("fval: type of `%v` (%T) not supported", v, v))
}

func ival(v interface{}) int64 {
	switch v.(type) {
	case int64:
		return v.(int64)
	case int:
		return int64(v.(int))
	case float64:
		return int64(v.(float64))
	}
	panic(fmt.Errorf("ival: type of `%v` (%T) not supported", v, v))
}

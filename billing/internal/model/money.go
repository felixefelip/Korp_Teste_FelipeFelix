package model

import "math"

func RoundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

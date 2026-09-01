package model_test

import (
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestRoundMoneyKeepsTwoDecimals(t *testing.T) {
	assert.Equal(t, 11.16, model.RoundMoney(11.1564))
	assert.Equal(t, 2.39, model.RoundMoney(2.388))
	assert.Equal(t, 301.0, model.RoundMoney(301))
}

func TestRoundMoneyRoundsHalfAwayFromZero(t *testing.T) {
	assert.Equal(t, 0.01, model.RoundMoney(0.005))
	assert.Equal(t, -0.01, model.RoundMoney(-0.005))
}

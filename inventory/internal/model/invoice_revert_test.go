package model_test

import (
	"encoding/json"
	"testing"

	"inventory/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func revertRequest() model.InvoiceStockRevertRequest {
	return model.InvoiceStockRevertRequest{
		InvoiceID:     7,
		InvoiceNumber: "NF-0007",
		CausationID:   "cause-2",
	}
}

func movementOf(movementType string, quantity int) model.StockMovement {
	return model.StockMovement{
		ProductID: 42,
		Type:      movementType,
		Origin:    model.MovementOriginInvoice,
		Quantity:  quantity,
		Confirmed: true,
	}
}

func revertReasonOf(t *testing.T, event model.OutboxEvent) string {
	t.Helper()

	var payload struct {
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(event.Payload, &payload))

	return payload.Reason
}

func TestRevertOfAnInvoiceWithNoMovementsSucceedsWithNothingToErase(t *testing.T) {
	decision, err := model.ResolveInvoiceRevert(revertRequest(), nil, available(10))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRevertedKey, decision.Event.RoutingKey,
		"a second revert must not leave the invoice hanging")
	assert.Empty(t, decision.Movements)
}

func TestRevertOfAnOutgoingInvoiceGivesTheStockBack(t *testing.T) {
	movements := []model.StockMovement{movementOf(model.MovementOut, 5)}

	decision, err := model.ResolveInvoiceRevert(revertRequest(), movements, available(3))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRevertedKey, decision.Event.RoutingKey,
		"undoing an outbound movement only adds stock, it can never go negative")
	assert.Len(t, decision.Movements, 1)
}

func TestRevertOfAnIncomingInvoiceIsRefusedWhenTheStockWasUsed(t *testing.T) {
	movements := []model.StockMovement{movementOf(model.MovementIn, 10)}

	decision, err := model.ResolveInvoiceRevert(revertRequest(), movements, available(4))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRevertRejectedKey, decision.Event.RoutingKey)
	assert.Equal(t, model.ReasonStockAlreadyUsed, revertReasonOf(t, decision.Event))
	assert.Empty(t, decision.Movements, "nothing is erased when the revert is refused")
}

func TestRevertOfAnIncomingInvoiceWorksWhileTheStockIsStillThere(t *testing.T) {
	movements := []model.StockMovement{movementOf(model.MovementIn, 10)}

	decision, err := model.ResolveInvoiceRevert(revertRequest(), movements, available(10))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRevertedKey, decision.Event.RoutingKey)
	assert.Len(t, decision.Movements, 1)
}

func TestBalanceAfterRemovingUndoesEachMovementInItsDirection(t *testing.T) {
	movements := []model.StockMovement{
		movementOf(model.MovementIn, 10),
		movementOf(model.MovementOut, 4),
	}

	balances := model.BalanceAfterRemoving(movements, available(20))

	assert.Equal(t, 14, balances[42], "20 - 10 entrada + 4 saida")
}

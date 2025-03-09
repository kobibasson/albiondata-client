package client

import (
	"github.com/ao-data/albiondata-client/log"
)

// eventInventoryDeleteItem contains data for the evInventoryDeleteItem event
type eventInventoryDeleteItem struct {
	ItemID int `mapstructure:"1"`
}

// Process handles the evInventoryDeleteItem event
func (e *eventInventoryDeleteItem) Process(state *albionState) {
	if state.Inventory == nil {
		log.Debug("Inventory tracker not initialized, ignoring evInventoryDeleteItem event")
		return
	}

	log.Debugf("Processing evInventoryDeleteItem: ItemID=%d", e.ItemID)
	state.Inventory.RemoveItem(e.ItemID)
} 
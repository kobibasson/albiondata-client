package client

import (
	"github.com/ao-data/albiondata-client/log"
)

// eventInventoryPutItem contains data for the evInventoryPutItem event
type eventInventoryPutItem struct {
	SlotID   int `mapstructure:"0"`
	ItemID   int `mapstructure:"1"`
	Quantity int `mapstructure:"2"`
}

// Process handles the evInventoryPutItem event
func (e *eventInventoryPutItem) Process(state *albionState) {
	if state.Inventory == nil {
		log.Debug("Inventory tracker not initialized, ignoring evInventoryPutItem event")
		return
	}

	log.Debugf("Processing evInventoryPutItem: ItemID=%d, Quantity=%d, SlotID=%d", e.ItemID, e.Quantity, e.SlotID)
	
	// No longer checking for bank vault access - using asset overview tabs instead
	state.Inventory.AddOrUpdateItem(e.ItemID, e.Quantity, e.SlotID)
} 
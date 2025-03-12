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

	if e.Quantity < 0 {
		log.Debugf("Processing evInventoryPutItem with negative quantity: ItemID=%d, Quantity=%d, SlotID=%d", e.ItemID, e.Quantity, e.SlotID)
	} else {
		log.Debugf("Processing evInventoryPutItem: ItemID=%d, Quantity=%d, SlotID=%d", e.ItemID, e.Quantity, e.SlotID)
	}
	
	state.Inventory.AddOrUpdateItem(e.ItemID, e.Quantity, e.SlotID)
} 
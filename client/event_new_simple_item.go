package client

import (
	"github.com/ao-data/albiondata-client/log"
)

// eventNewSimpleItem contains data for the evNewSimpleItem event
type eventNewSimpleItem struct {
	SlotID   int `mapstructure:"0"`
	ItemID   int `mapstructure:"1"`
	Quantity int `mapstructure:"2"`
	// ContainerID int `mapstructure:"4"` // Optional: Add ContainerID field if needed
}

// Process handles the evNewSimpleItem event
func (e *eventNewSimpleItem) Process(state *albionState) {
	if state.Inventory == nil {
		log.Debug("Inventory tracker not initialized, ignoring evNewSimpleItem event")
		return
	}

	if e.Quantity < 0 {
		log.Debugf("Processing evNewSimpleItem with negative quantity: ItemID=%d, Quantity=%d, SlotID=%d", e.ItemID, e.Quantity, e.SlotID)
	} else {
		log.Debugf("Processing evNewSimpleItem: ItemID=%d, Quantity=%d, SlotID=%d", e.ItemID, e.Quantity, e.SlotID)
	}
	
	// No longer checking for bank vault access - using asset overview tabs instead
	state.Inventory.AddOrUpdateItem(e.ItemID, e.Quantity, e.SlotID)
} 
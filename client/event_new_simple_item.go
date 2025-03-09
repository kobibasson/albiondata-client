package client

import (
	"github.com/ao-data/albiondata-client/log"
)

// eventNewSimpleItem contains data for the evNewSimpleItem event
type eventNewSimpleItem struct {
	ItemID   int `mapstructure:"1"`
	Quantity int `mapstructure:"2"`
}

// Process handles the evNewSimpleItem event
func (e *eventNewSimpleItem) Process(state *albionState) {
	if state.Inventory == nil {
		log.Debug("Inventory tracker not initialized, ignoring evNewSimpleItem event")
		return
	}

	if e.Quantity < 0 {
		log.Debugf("Processing evNewSimpleItem with negative quantity: ItemID=%d, Quantity=%d", e.ItemID, e.Quantity)
	} else {
		log.Debugf("Processing evNewSimpleItem: ItemID=%d, Quantity=%d", e.ItemID, e.Quantity)
	}
	
	state.Inventory.AddOrUpdateItem(e.ItemID, e.Quantity)
} 
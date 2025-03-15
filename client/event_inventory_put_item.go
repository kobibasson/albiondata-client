package client

import (
	"time"

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
	
	// Check if this event occurred within 3 seconds of a bank vault access
	isBank := false
	locationID := ""
	now := time.Now().Unix()
	
	if state.LastBankVaultTime > 0 && now - state.LastBankVaultTime <= 3 {
		isBank = true
		locationID = state.LastBankVaultLocationID
		log.Debugf("Item is in a bank vault: LocationID=%s", locationID)
	}
	
	state.Inventory.AddOrUpdateItemWithBank(e.ItemID, e.Quantity, e.SlotID, isBank, locationID)
} 
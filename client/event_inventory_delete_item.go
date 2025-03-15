package client

import (
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// eventInventoryDeleteItem contains data for the evInventoryDeleteItem event
type eventInventoryDeleteItem struct {
	SlotID int `mapstructure:"0"`
	ItemID int `mapstructure:"1"`
}

// Process handles the evInventoryDeleteItem event
func (e *eventInventoryDeleteItem) Process(state *albionState) {
	if state.Inventory == nil {
		log.Debug("Inventory tracker not initialized, ignoring evInventoryDeleteItem event")
		return
	}

	log.Debugf("Processing evInventoryDeleteItem: ItemID=%d, SlotID=%d", e.ItemID, e.SlotID)
	
	// Check if this event occurred within 3 seconds of a bank vault access
	isBank := false
	locationID := ""
	now := time.Now().Unix()
	
	if state.LastBankVaultTime > 0 && now - state.LastBankVaultTime <= 3 {
		isBank = true
		locationID = state.LastBankVaultLocationID
		log.Debugf("Item is in a bank vault: LocationID=%s", locationID)
	}
	
	state.Inventory.RemoveItemWithBank(e.ItemID, isBank, locationID)
} 
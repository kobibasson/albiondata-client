package client

import (
	"time"

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
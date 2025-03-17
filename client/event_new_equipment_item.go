package client

import (
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// eventNewEquipmentItem contains data for the evNewEquipmentItem event
type eventNewEquipmentItem struct {
	SlotID   int  `mapstructure:"0"`
	ItemID   int  `mapstructure:"1"`
	Quantity int  `mapstructure:"2"`
	Equipped bool `mapstructure:"3"` // Whether the item is equipped
	// Similar structure to evNewSimpleItem
}

// Process handles the evNewEquipmentItem event
func (e *eventNewEquipmentItem) Process(state *albionState) {
	if state.Inventory == nil {
		log.Debug("Inventory tracker not initialized, ignoring evNewEquipmentItem event")
		return
	}

	if e.Quantity < 0 {
		log.Debugf("Equipment item: ID=%d, Qty=%d, Slot=%d, Equipped=%v (negative quantity)", 
			e.ItemID, e.Quantity, e.SlotID, e.Equipped)
	} else {
		log.Debugf("Equipment item: ID=%d, Qty=%d, Slot=%d, Equipped=%v", 
			e.ItemID, e.Quantity, e.SlotID, e.Equipped)
	}
	
	// Check if this event occurred within 4 seconds of a cluster change
	shouldSendWebhook := false
	
	timeSinceClusterChange := time.Since(state.Inventory.lastClusterChange)
	if timeSinceClusterChange <= 4*time.Second {
		shouldSendWebhook = true
	}
	
	// Check if this event occurred within 3 seconds of a bank vault access
	// We'll use the state variables for backward compatibility, but the inventory tracker
	// now handles bank vault access internally
	isBank := state.Inventory.isBank
	locationID := state.Inventory.currentBankLocation
	
	// For backward compatibility, also check the state variables
	if !isBank && state.LastBankVaultTime > 0 && time.Now().Unix() - state.LastBankVaultTime <= 3 {
		isBank = true
		locationID = state.LastBankVaultLocationID
	}
	
	// Temporarily store the webhook URL if we don't want to send updates
	var webhookURL string
	if !shouldSendWebhook {
		webhookURL = state.Inventory.webhookURL
		state.Inventory.webhookURL = "" // Temporarily disable webhook
	}
	
	// Update inventory with equipped information
	state.Inventory.AddOrUpdateEquipmentItem(e.ItemID, e.Quantity, e.SlotID, e.Equipped, isBank, locationID)
	
	// Restore webhook URL if we disabled it
	if !shouldSendWebhook {
		state.Inventory.webhookURL = webhookURL
	}
} 
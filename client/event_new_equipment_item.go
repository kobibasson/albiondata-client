package client

import (
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// eventNewEquipmentItem contains data for the evNewEquipmentItem event
type eventNewEquipmentItem struct {
	SlotID   int `mapstructure:"0"`
	ItemID   int `mapstructure:"1"`
	Quantity int `mapstructure:"2"`
	// Similar structure to evNewSimpleItem
}

// Process handles the evNewEquipmentItem event
func (e *eventNewEquipmentItem) Process(state *albionState) {
	if state.Inventory == nil {
		log.Debug("Inventory tracker not initialized, ignoring evNewEquipmentItem event")
		return
	}

	if e.Quantity < 0 {
		log.Debugf("Equipment item: ID=%d, Qty=%d, Slot=%d (negative quantity)", e.ItemID, e.Quantity, e.SlotID)
	} else {
		log.Debugf("Equipment item: ID=%d, Qty=%d, Slot=%d", e.ItemID, e.Quantity, e.SlotID)
	}
	
	// Check if this event occurred within 4 seconds of a cluster change
	shouldSendWebhook := false
	
	timeSinceClusterChange := time.Since(state.Inventory.lastClusterChange)
	if timeSinceClusterChange <= 4*time.Second {
		shouldSendWebhook = true
	}
	
	// Check if this event occurred within 3 seconds of a bank vault access
	isBank := false
	locationID := ""
	
	if state.LastBankVaultTime > 0 && time.Now().Unix() - state.LastBankVaultTime <= 3 {
		isBank = true
		locationID = state.LastBankVaultLocationID
	}
	
	// Temporarily store the webhook URL if we don't want to send updates
	var webhookURL string
	if !shouldSendWebhook {
		webhookURL = state.Inventory.webhookURL
		state.Inventory.webhookURL = "" // Temporarily disable webhook
	}
	
	// Update inventory
	state.Inventory.AddOrUpdateItemWithBank(e.ItemID, e.Quantity, e.SlotID, isBank, locationID)
	
	// Restore webhook URL if we disabled it
	if !shouldSendWebhook {
		state.Inventory.webhookURL = webhookURL
	}
} 
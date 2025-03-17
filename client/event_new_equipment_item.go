package client

import (
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
		log.Debugf("Processing evNewEquipmentItem with negative quantity: ItemID=%d, Quantity=%d, SlotID=%d, Equipped=%v", 
			e.ItemID, e.Quantity, e.SlotID, e.Equipped)
	} else {
		log.Debugf("Processing evNewEquipmentItem: ItemID=%d, Quantity=%d, SlotID=%d, Equipped=%v", 
			e.ItemID, e.Quantity, e.SlotID, e.Equipped)
	}
	
	// Handle equipment items the same way as simple items, but preserve equipped status
	// No longer checking for bank vault access - using asset overview tabs instead
	// Get bank state from inventory tracker
	isBank := state.Inventory.isBank
	locationID := state.Inventory.currentBankLocation
	
	// Update inventory with equipped information
	state.Inventory.AddOrUpdateEquipmentItem(e.ItemID, e.Quantity, e.SlotID, e.Equipped, isBank, locationID)
} 
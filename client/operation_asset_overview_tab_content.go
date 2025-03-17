package client

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// operationAssetOverviewTabContent contains data for the opAssetOverviewTabContent operation
type operationAssetOverviewTabContent struct {
	TabID       []byte `mapstructure:"0"`
	Unknown1    int    `mapstructure:"1"`
	ItemIDs     []int  `mapstructure:"2"`
	Positions   []int  `mapstructure:"3"`
	Quantities  []int  `mapstructure:"4"`
	PriceMin    []int  `mapstructure:"5"`
	PriceMax    []int  `mapstructure:"6"`
	Quality     []int  `mapstructure:"7"`
	OwnerName   string `mapstructure:"8"`
	Enchantment [][]int `mapstructure:"9"`
	Tier        [][]int `mapstructure:"10"`
	Unknown2    []int  `mapstructure:"11"`
	Unknown3    int    `mapstructure:"13"`
}

// Process handles the opAssetOverviewTabContent operation
// IMPORTANT: This is the ONLY operation that should trigger bank queues.
// Bank vault access notifications should only update state, not trigger queues.
func (op operationAssetOverviewTabContent) Process(state *albionState) {
	tabIDHex := hex.EncodeToString(op.TabID)
	log.Infof("[BANK-OVERRIDE] Processing bank tab content: TabID=%s, Items=%d", tabIDHex, len(op.ItemIDs))
	
	if state.Inventory == nil {
		log.Warnf("[BANK-OVERRIDE] Cannot process bank tab content: inventory tracker not initialized")
		return
	}
	
	// Store the current bank tab ID
	state.CurrentBankTabID = string(op.TabID)
	
	// Get the tab name if available
	tabName := "Unknown Bank Tab"
	if name, ok := state.BankTabs[state.CurrentBankTabID]; ok {
		tabName = name
	}
	
	log.Infof("[BANK-OVERRIDE] Bank tab: %s (ID: %s)", tabName, tabIDHex)
	
	// Update the last tab content time
	currentTime := time.Now().Unix()
	state.LastTabContentTime = currentTime
	
	// Increment the location ID for each tab content operation
	// This helps track different bank tabs (1-5)
	state.LastTabContentLocationID++
	if state.LastTabContentLocationID > 5 {
		state.LastTabContentLocationID = 1
	}
	locationID := state.LastTabContentLocationID
	locationIDStr := fmt.Sprintf("%d", locationID)
	
	log.Infof("[BANK-OVERRIDE] Queue opened: Bank inventory queue (tab content trigger, LocationID=%s)", locationIDStr)
	
	// Notify the inventory tracker of bank vault access
	// This updates the state but does NOT trigger bank queues
	state.Inventory.NotifyBankVaultAccess(locationIDStr)
	
	// Process the items in this bank tab
	if len(op.ItemIDs) > 0 {
		for i, itemID := range op.ItemIDs {
			if i < len(op.Quantities) {
				quantity := op.Quantities[i]
				
				// Get quality if available
				quality := 1
				if i < len(op.Quality) {
					quality = op.Quality[i]
				}
				
				// Get position if available
				position := 0
				if i < len(op.Positions) {
					position = op.Positions[i]
				}
				
				log.Debugf("[BANK-OVERRIDE] Bank item: ID=%d, Quantity=%d, Quality=%d, Position=%d, LocationID=%s", 
					itemID, quantity, quality, position, locationIDStr)
				
				// Create a unique slot ID for bank items
				slotID := 2000000 + position // Use 2 million range for bank items to distinguish from equipment
				
				log.Infof("[BANK-OVERRIDE] Processing bank item: ID=%d, Quantity=%d, Tab=%s, LocationID=%s", 
					itemID, quantity, tabName, locationIDStr)
				
				// Send a webhook update for this bank item
				// Each item is added to the queue with the current location ID
				state.Inventory.AddOrUpdateBankItem(itemID, quantity, slotID, tabName, locationIDStr)
			}
		}
		
		// Check if it's been more than 10 seconds since the last tab content event
		// If so, send the batch update immediately
		if currentTime - state.LastTabContentTime > 10 {
			log.Infof("[BANK-OVERRIDE] Sending immediate bank batch: more than 10s since last tab content")
			state.Inventory.SendBankItemsBatch(tabName)
			
			// Reset the location ID to 0 after sending the batch
			if locationID == 5 {
				state.LastTabContentLocationID = 0 // Will become 1 on next event
				log.Infof("[BANK-OVERRIDE] Reset location ID to 0 after sending batch with locationId=5")
			}
		} else {
			// Otherwise, set a timer to send the batch after 10 seconds if no more tab content events arrive
			log.Infof("[BANK-OVERRIDE] Setting bank batch timer: 10s for location ID %s", locationIDStr)
			state.Inventory.SetBankBatchTimer(10 * time.Second, tabName)
			
			// If this is location ID 5, we need to make sure the next tab content operation starts with location ID 1
			if locationID == 5 {
				state.LastTabContentLocationID = 0 // Will become 1 on next event
				log.Infof("[BANK-OVERRIDE] Reset location ID to 0 for next tab content operation")
			}
		}
		
		log.Infof("[BANK-OVERRIDE] Bank queue status: %d pending events", len(state.Inventory.pendingEvents))
	} else {
		log.Warnf("[BANK-OVERRIDE] No items found in bank tab: %s", tabName)
	}
} 
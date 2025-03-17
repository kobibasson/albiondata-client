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
func (op operationAssetOverviewTabContent) Process(state *albionState) {
	tabIDHex := hex.EncodeToString(op.TabID)
	fmt.Printf("\n[OPERATION] ===== DETECTED opAssetOverviewTabContent: TabID=%s, Items=%d =====\n", tabIDHex, len(op.ItemIDs))
	log.Infof("Processing opAssetOverviewTabContent: TabID=%s, Items=%d", tabIDHex, len(op.ItemIDs))
	
	// Log the raw operation data for debugging
	log.Infof("opAssetOverviewTabContent operation detected with TabID=%s, Items=%d", tabIDHex, len(op.ItemIDs))
	log.Infof("Owner Name: %s, Unknown1: %d, Unknown3: %d", op.OwnerName, op.Unknown1, op.Unknown3)
	
	if state.Inventory == nil {
		fmt.Printf("[BANK CONTENT] Inventory tracker not initialized, ignoring opAssetOverviewTabContent\n")
		log.Warn("[BANK CONTENT] Inventory tracker not initialized, ignoring opAssetOverviewTabContent")
		fmt.Printf("[OPERATION] ===== END opAssetOverviewTabContent =====\n\n")
		return
	}
	
	// Store the current bank tab ID
	state.CurrentBankTabID = string(op.TabID)
	
	// Get the tab name if available
	tabName := "Unknown Bank Tab"
	if name, ok := state.BankTabs[state.CurrentBankTabID]; ok {
		tabName = name
	}
	
	fmt.Printf("[BANK CONTENT] Current tab: %s (ID: %s)\n", tabName, tabIDHex)
	log.Infof("[BANK CONTENT] Current tab: %s (ID: %s)", tabName, tabIDHex)
	
	fmt.Printf("[BANK CONTENT] Inventory tracker state:\n")
	fmt.Printf("  Webhook URL: %s\n", state.Inventory.webhookURL)
	fmt.Printf("  Webhook Enabled: %v\n", state.Inventory.webhookURL != "")
	fmt.Printf("  Pending Events: %d\n", len(state.Inventory.pendingEvents))
	
	log.Infof("[BANK CONTENT] Inventory tracker state - Webhook URL: %s, Enabled: %v", 
		state.Inventory.webhookURL, state.Inventory.webhookURL != "")
	log.Infof("[BANK CONTENT] Pending webhook events: %d", len(state.Inventory.pendingEvents))
	log.Infof("[BANK CONTENT] Bank state - Is Bank: %v, Location: %s", 
		state.Inventory.isBank, state.Inventory.currentBankLocation)
	
	// Process the items in this bank tab
	if len(op.ItemIDs) > 0 {
		fmt.Printf("[BANK CONTENT] Processing %d items in bank tab: %s\n", len(op.ItemIDs), tabName)
		log.Infof("[BANK CONTENT] Processing %d items in bank tab: %s", len(op.ItemIDs), tabName)
		
		// Implement the new location ID algorithm
		currentTime := time.Now().Unix()
		locationID := 1 // Default to 1 for first event
		
		// Check if we need to increment the location ID or reset it
		if state.LastTabContentLocationID > 0 {
			// If more than 10 seconds have passed since the last tab content event, reset to 1
			if currentTime - state.LastTabContentTime > 10 {
				locationID = 1
				log.Infof("[BANK CONTENT] More than 10 seconds since last tab content event, resetting location ID to 1")
			} else {
				// Increment the location ID, but cap at 5
				locationID = state.LastTabContentLocationID + 1
				if locationID > 5 {
					locationID = 1
					log.Infof("[BANK CONTENT] Location ID exceeded 5, resetting to 1")
				}
			}
		}
		
		// Update the state with the new location ID and timestamp
		state.LastTabContentLocationID = locationID
		state.LastTabContentTime = currentTime
		
		// Create a location ID string
		locationIDStr := fmt.Sprintf("%d", locationID)
		
		log.Infof("[BANK CONTENT] Using location ID: %s for tab content batch", locationIDStr)
		fmt.Printf("[BANK CONTENT] Using location ID: %s for tab content batch\n", locationIDStr)
		
		// Notify the inventory tracker that we're accessing bank items
		state.Inventory.NotifyBankVaultAccess(locationIDStr)
		fmt.Printf("[BANK CONTENT] Notified inventory tracker of bank vault access: %s\n", locationIDStr)
		log.Infof("[BANK CONTENT] Notified inventory tracker of bank vault access: %s", locationIDStr)
		
		// Process each item
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
				
				fmt.Printf("[BANK ITEM] ID=%d, Quantity=%d, Quality=%d, Position=%d\n", 
					itemID, quantity, quality, position)
				log.Infof("[BANK ITEM] ID=%d, Quantity=%d, Quality=%d, Position=%d", 
					itemID, quantity, quality, position)
				
				// Create a unique slot ID for bank items
				slotID := 2000000 + position // Use 2 million range for bank items to distinguish from equipment
				
				// Send a webhook update for this bank item
				state.Inventory.AddOrUpdateBankItem(itemID, quantity, slotID, tabName, locationIDStr)
				log.Infof("[BANK ITEM] Added/updated bank item: ID=%d, Quantity=%d, SlotID=%d", 
					itemID, quantity, slotID)
			}
		}
		
		// Check if it's been more than 10 seconds since the last tab content event
		// If so, send the batch update immediately
		if currentTime - state.LastTabContentTime > 10 {
			// Send a batch update with all bank items
			fmt.Printf("[BANK CONTENT] More than 10 seconds since last tab content event, sending batch update now\n")
			log.Infof("[BANK CONTENT] More than 10 seconds since last tab content event, sending batch update now")
			state.Inventory.SendBankItemsBatch(tabName, locationIDStr)
		} else {
			// Otherwise, just log that we're waiting to send the batch
			fmt.Printf("[BANK CONTENT] Queueing items for batch update (will send after 10 seconds of inactivity)\n")
			log.Infof("[BANK CONTENT] Queueing items for batch update (will send after 10 seconds of inactivity)")
			
			// Set a timer to send the batch after 10 seconds if no more tab content events arrive
			state.Inventory.SetBankBatchTimer(10 * time.Second, tabName, locationIDStr)
		}
		
		// Print pending events after processing
		fmt.Printf("[BANK CONTENT] Pending events after processing: %d\n", len(state.Inventory.pendingEvents))
		log.Infof("[BANK CONTENT] Pending events after processing: %d", len(state.Inventory.pendingEvents))
	} else {
		fmt.Printf("[BANK CONTENT] No items found in bank tab: %s\n", tabName)
		log.Warn("[BANK CONTENT] No items found in bank tab: %s", tabName)
	}
	
	fmt.Printf("[OPERATION] ===== END opAssetOverviewTabContent =====\n\n")
	log.Infof("Completed processing opAssetOverviewTabContent operation")
} 
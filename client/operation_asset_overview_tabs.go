package client

import (
	"encoding/hex"
	"fmt"

	"github.com/ao-data/albiondata-client/log"
)

// operationAssetOverviewTabs contains data for the opAssetOverviewTabs operation
type operationAssetOverviewTabs struct {
	RequestID []byte `mapstructure:"0"`
	TabIDs    [][]byte `mapstructure:"1"`
	TabNames  []string `mapstructure:"2"`
	TabIcons  []string `mapstructure:"3"`
	Unknown1  []int    `mapstructure:"4"`
	ItemCount []int    `mapstructure:"5"`
	Unknown2  []int    `mapstructure:"6"`
}

// Process handles the opAssetOverviewTabs operation
func (op operationAssetOverviewTabs) Process(state *albionState) {
	fmt.Printf("\n[OPERATION] ===== DETECTED opAssetOverviewTabs: %d tabs =====\n", len(op.TabIDs))
	log.Infof("Processing opAssetOverviewTabs: %d tabs", len(op.TabIDs))
	
	// Log the raw operation data for debugging
	log.Infof("opAssetOverviewTabs operation detected with %d tabs", len(op.TabIDs))
	if len(op.RequestID) > 0 {
		log.Infof("RequestID: %s", hex.EncodeToString(op.RequestID))
	}
	
	// Store the bank tabs information in the state
	if state.Inventory != nil && len(op.TabIDs) > 0 {
		// This operation provides information about bank tabs
		// We'll store this information for use when processing tab content
		state.BankTabs = make(map[string]string)
		
		fmt.Printf("[BANK TABS] Found %d bank tabs:\n", len(op.TabIDs))
		log.Infof("[BANK TABS] Found %d bank tabs", len(op.TabIDs))
		
		for i, tabID := range op.TabIDs {
			if i < len(op.TabNames) {
				// Convert byte array to string for use as map key
				tabIDStr := string(tabID)
				tabIDHex := hex.EncodeToString(tabID)
				state.BankTabs[tabIDStr] = op.TabNames[i]
				fmt.Printf("[BANK TAB] ID: %s, Name: %s, Items: %d\n", 
					tabIDHex, op.TabNames[i], op.ItemCount[i])
				log.Infof("[BANK TAB] ID: %s, Name: %s, Items: %d", 
					tabIDHex, op.TabNames[i], op.ItemCount[i])
			}
		}
		
		// Print inventory tracker state if available
		if state.Inventory != nil {
			fmt.Printf("[BANK TABS] Inventory tracker state:\n")
			fmt.Printf("  Webhook URL: %s\n", state.Inventory.webhookURL)
			fmt.Printf("  Webhook Enabled: %v\n", state.Inventory.webhookURL != "")
			log.Infof("[BANK TABS] Inventory tracker state - Webhook URL: %s, Enabled: %v", 
				state.Inventory.webhookURL, state.Inventory.webhookURL != "")
			
			// Print pending events count
			fmt.Printf("  Pending Events: %d\n", len(state.Inventory.pendingEvents))
			log.Infof("[BANK TABS] Pending webhook events: %d", len(state.Inventory.pendingEvents))
			
			// Print bank state
			fmt.Printf("  Is Bank: %v\n", state.Inventory.isBank)
			fmt.Printf("  Current Bank Location: %s\n", state.Inventory.currentBankLocation)
			log.Infof("[BANK TABS] Bank state - Is Bank: %v, Location: %s", 
				state.Inventory.isBank, state.Inventory.currentBankLocation)
		} else {
			fmt.Printf("[BANK TABS] Inventory tracker not initialized\n")
			log.Warn("[BANK TABS] Inventory tracker not initialized")
		}
	} else {
		if state.Inventory == nil {
			fmt.Printf("[BANK TABS] Inventory tracker not initialized\n")
			log.Warn("[BANK TABS] Inventory tracker not initialized")
		}
		if len(op.TabIDs) == 0 {
			fmt.Printf("[BANK TABS] No tabs found\n")
			log.Warn("[BANK TABS] No tabs found in operation")
		}
	}
	fmt.Printf("[OPERATION] ===== END opAssetOverviewTabs =====\n\n")
	log.Infof("Completed processing opAssetOverviewTabs operation")
} 
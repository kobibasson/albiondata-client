package client

import (
	"strings"
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// Helper function to get current time in Unix format
func getNow() int64 {
	return time.Now().Unix()
}

// Helper function to extract UUID from location ID
func extractUUID(locationID string) string {
	// Split the location ID at the @ sign and return the first part
	parts := strings.Split(locationID, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return locationID // Return the original if no @ sign is found
}

// eventBankVaultInfo contains data for the evBankVaultInfo event
type eventBankVaultInfo struct {
	EventType int    `mapstructure:"0"`
	LocationID string `mapstructure:"1"`
}

// Process handles the evBankVaultInfo event
func (e *eventBankVaultInfo) Process(state *albionState) {
	// Log the event for debugging
	log.Debugf("Processing evBankVaultInfo: LocationID=%s (Used for filtering evNewSimpleItem events)", e.LocationID)
	
	// Extract just the UUID part from the location ID
	uuid := extractUUID(e.LocationID)
	
	// Store the bank vault info in the state for reference and filtering
	state.LastBankVaultLocationID = uuid
	state.LastBankVaultTime = getNow()
	
	// Update global variables for reference
	bankVaultMutex.Lock()
	defer bankVaultMutex.Unlock()
	
	lastBankVaultLocationID = uuid
	lastBankVaultTime = getNow()
	
	log.Debugf("Updated LastBankVaultTime to %d for filtering evNewSimpleItem events", state.LastBankVaultTime)
	
	// No longer notifying inventory tracker from this event
	// Bank vault access is now handled by asset overview tab content operations
} 
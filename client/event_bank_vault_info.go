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
	log.Debugf("Processing evBankVaultInfo: LocationID=%s", e.LocationID)
	
	// Extract just the UUID part from the location ID
	uuid := extractUUID(e.LocationID)
	log.Debugf("Extracted UUID from location ID: %s", uuid)
	
	// Store the bank vault info in the state
	state.LastBankVaultLocationID = uuid
	state.LastBankVaultTime = getNow()
	
	// Update global variables for webhook access
	bankVaultMutex.Lock()
	defer bankVaultMutex.Unlock()
	
	lastBankVaultLocationID = uuid
	lastBankVaultTime = getNow()
	
	log.Debugf("Updated global bank vault info: LocationID=%s, Time=%d", lastBankVaultLocationID, lastBankVaultTime)
} 
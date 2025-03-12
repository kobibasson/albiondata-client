package client

import (
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// eventUpdateChatSettings contains data for the evUpdateChatSettings event
type eventUpdateChatSettings struct {
	// Add fields if needed based on the actual data structure
}

// Process handles the evUpdateChatSettings event
func (e eventUpdateChatSettings) Process(state *albionState) {
	log.Debug("Got UpdateChatSettings event (likely player leaving)...")
	
	// Update the last leave time in the state
	state.LastLeaveTime = time.Now().Unix()
	
	if state.Inventory != nil {
		log.Debug("Notifying inventory tracker of leave event via UpdateChatSettings...")
		state.Inventory.OnLeave()
	}
} 
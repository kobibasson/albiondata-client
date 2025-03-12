package client

import (
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// eventLeave contains data for the evLeave event
// NOTE: This event is no longer used. We now use evUpdateChatSettings instead
// because evLeave happens for other players too, not just the current player.
// This file is kept for reference only.
type eventLeave struct {
	// Add fields if needed based on the actual data structure
}

// Process handles the evLeave event
// NOTE: This event is no longer used. We now use evUpdateChatSettings instead.
func (e eventLeave) Process(state *albionState) {
	log.Debug("Got Leave event...")
	
	// Update the last leave time in the state
	state.LastLeaveTime = time.Now().Unix()
	
	if state.Inventory != nil {
		log.Debug("Notifying inventory tracker of leave event...")
		state.Inventory.OnLeave()
	}
} 
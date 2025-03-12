package client

import (
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// operationChangeCluster contains data for the opChangeCluster operation
type operationChangeCluster struct {
	// Add fields if needed based on the actual data structure
}

// Process handles the opChangeCluster operation
func (op operationChangeCluster) Process(state *albionState) {
	log.Debug("Got ChangeCluster operation...")
	
	// Update the last cluster change time in the state
	state.LastClusterChangeTime = time.Now().Unix()
	
	if state.Inventory != nil {
		log.Debug("Notifying inventory tracker of cluster change...")
		state.Inventory.OnClusterChange()
	}
} 
package client

import (
	"strconv"
	"strings"
	"time"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
)

// operationJoin contains data for the opJoin operation
type operationJoin struct {
	// Add fields if needed based on the actual data structure
}

// Process handles the opJoin operation
func (op operationJoin) Process(state *albionState) {
	log.Debug("Got Join operation...")
	
	// Update the last join time in the state
	state.LastJoinTime = time.Now().Unix()
	
	if state.Inventory != nil {
		log.Debug("Notifying inventory tracker of join operation...")
		state.Inventory.OnJoin()
	}
}

// operationJoinResponse contains data for the opJoin response
type operationJoinResponse struct {
	CharacterID   lib.CharacterID `mapstructure:"1"`
	CharacterName string          `mapstructure:"2"`
	Location      string          `mapstructure:"7"`
}

// Process handles the opJoin response
func (op operationJoinResponse) Process(state *albionState) {
	log.Debugf("Got JoinResponse operation...")

	// Reset the AODataServerID here. This leads to a fresh execution
	// of SetServerID() incase the player switched servers
	state.AODataServerID = 0

	// Hack for second caerleon marketplace
	if strings.HasSuffix(op.Location, "-Auction2") {
		op.Location = strings.Replace(op.Location, "-Auction2", "", -1)
	}

	// Allow for smugglers rest locations markets to be parsed by setting a valid location int
	if strings.HasPrefix(op.Location, "BLACKBANK-") {
		op.Location = strings.Replace(op.Location, "BLACKBANK-", "", -1)
	}

	loc, err := strconv.Atoi(op.Location)
	if err != nil {
		log.Debugf("Unable to convert zoneID to int. Probably an instance.")
		state.LocationId = -2
	} else {
		state.LocationId = loc
	}
	log.Infof("Updating player location to %v.", op.Location)

	if state.CharacterId != op.CharacterID {
		log.Infof("Updating player ID to %v.", op.CharacterID)
	}
	state.CharacterId = op.CharacterID

	if state.CharacterName != op.CharacterName {
		log.Infof("Updating player to %v.", op.CharacterName)
	}
	state.CharacterName = op.CharacterName
	
	// Update character information in the inventory tracker
	state.UpdateInventoryCharacterInfo()
	
	// Also update the last join time in the state
	state.LastJoinTime = time.Now().Unix()
	
	if state.Inventory != nil {
		log.Debug("Notifying inventory tracker of join operation from response...")
		state.Inventory.OnJoin()
	}
}

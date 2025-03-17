package client

import (
	"strings"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
)

// CacheSize limit size of messages in cache
const CacheSize = 8192

type marketHistoryInfo struct {
	albionId  int32
	timescale lib.Timescale
	quality   uint8
}

type albionState struct {
	LocationId           int
	LocationString       string
	CharacterId          lib.CharacterID
	CharacterName        string
	GameServerIP         string
	AODataServerID       int
	AODataIngestBaseURL  string
	WaitingForMarketData bool
	Inventory            *PlayerInventory
	LastClusterChangeTime int64 // Timestamp of the last cluster change
	LastJoinTime         int64 // Timestamp of the last join operation
	LastLeaveTime        int64 // Timestamp of the last leave event
	LastBankVaultLocationID string // ID of the last bank vault accessed
	LastBankVaultTime    int64 // Timestamp of the last bank vault access
	BankTabs             map[string]string // Map of bank tab IDs to names
	CurrentBankTabID     string // Current bank tab being viewed
	LastTabContentLocationID int // Last location ID used for tab content
	LastTabContentTime   int64 // Timestamp of the last tab content operation

	// A lot of information is sent out but not contained in the response when requesting marketHistory (e.g. ID)
	// This information is stored in marketHistoryInfo
	// This array acts as a type of cache for that info
	// The index is the message number (param255) % CacheSize
	marketHistoryIDLookup [CacheSize]marketHistoryInfo
	// TODO could this be improved?!
}

func (state albionState) IsValidLocation() bool {
	if state.LocationId < 0 {
		return false
	}
	return true
}

// UpdateInventoryCharacterInfo updates the character information in the inventory tracker
func (state *albionState) UpdateInventoryCharacterInfo() {
	if state.Inventory == nil {
		return
	}

	if state.CharacterId != "" && state.CharacterName != "" {
		state.Inventory.UpdateCharacterInfo(string(state.CharacterId), state.CharacterName)
	}
}

func (state albionState) GetServer() (int, string) {
	// default to 0
	var serverID = 0
	var AODataIngestBaseURL = ""

	// if we happen to have a server id stored in state, lets re-default to that
	if state.AODataServerID != 0 {
		serverID = state.AODataServerID
	}
	if state.AODataIngestBaseURL != "" {
		AODataIngestBaseURL = state.AODataIngestBaseURL
	}

	// we get packets from other than game servers, so determine if it's a game server
	// based on soruce ip and if its east/west servers
	var isAlbionIP = false
	if strings.HasPrefix(state.GameServerIP, "5.188.125.") {
		// west server class c ip range
		serverID = 1
		isAlbionIP = true
		AODataIngestBaseURL = "http+pow://pow.west.albion-online-data.com"
	} else if strings.HasPrefix(state.GameServerIP, "5.45.187.") {
		// east server class c ip range
		isAlbionIP = true
		serverID = 2
		AODataIngestBaseURL = "http+pow://pow.east.albion-online-data.com"
	} else if strings.HasPrefix(state.GameServerIP, "193.169.238.") {
		// eu server class c ip range
		isAlbionIP = true
		serverID = 3
		AODataIngestBaseURL = "http+pow://pow.europe.albion-online-data.com"
	}

	// if this was a known albion online server ip, then let's log it
	if isAlbionIP {
		log.Tracef("Returning Server ID %v (ip src: %v)", serverID, state.GameServerIP)
		log.Tracef("Returning AODataIngestBaseURL %v (ip src: %v)", AODataIngestBaseURL, state.GameServerIP)
	}

	return serverID, AODataIngestBaseURL
}

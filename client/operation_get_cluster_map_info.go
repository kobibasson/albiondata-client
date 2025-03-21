package client

import (
	"strconv"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

type operationGetClusterMapInfo struct {
}

func (op operationGetClusterMapInfo) Process(state *albionState) {
	log.Debug("Got GetClusterMapInfo operation...")
}

type operationGetClusterMapInfoResponse struct {
	ZoneID          string   `mapstructure:"0"`
	BuildingType    []int    `mapstructure:"5"`
	AvailableFood   []int    `mapstructure:"10"`
	Reward          []int    `mapstructure:"12"`
	AvailableSilver []int    `mapstructure:"13"`
	Owners          []string `mapstructure:"14"`
	Buildable       []bool   `mapstructure:"19"`
	IsForSale       []bool   `mapstructure:"27"`
	BuyPrice        []int    `mapstructure:"28"`
}

func (op operationGetClusterMapInfoResponse) Process(state *albionState) {
	log.Debug("Got response to GetClusterMapInfo operation...")

	zoneInt, err := strconv.Atoi(op.ZoneID)
	if err != nil {
		log.Debugf("Unable to convert zoneID to int. Probably an instance.. ZoneID: %v", op.ZoneID)
		return
	}

	upload := lib.MapDataUpload{
		ZoneID:          zoneInt,
		BuildingType:    op.BuildingType,
		AvailableFood:   op.AvailableFood,
		Reward:          op.Reward,
		AvailableSilver: op.AvailableSilver,
		Owners:          op.Owners,
		Buildable:       op.Buildable,
		IsForSale:       op.IsForSale,
		BuyPrice:        op.BuyPrice,
	}

	identifier, _ := uuid.NewV4()
	log.Infof("Sending map data to ingest (Identifier: %s)", identifier)
	sendMsgToPublicUploaders(upload, lib.NatsMapDataIngest, state, identifier.String())
}

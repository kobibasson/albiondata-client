package client

import (
	"encoding/gob"
	"encoding/hex"
	"os"

	"github.com/ao-data/albiondata-client/log"
	photon "github.com/ao-data/photon_spectator"
)

//Router struct definitions
type Router struct {
	albionstate         *albionState
	newOperation        chan operation
	recordPhotonCommand chan photon.PhotonCommand
	quit                chan bool
}

func newRouter() *Router {
	return &Router{
		albionstate:         &albionState{LocationId: -1, LastTabContentLocationID: 0},
		newOperation:        make(chan operation, 1000),
		recordPhotonCommand: make(chan photon.PhotonCommand, 1000),
		quit:                make(chan bool, 1),
	}
}

func (r *Router) run() {
	var encoder *gob.Encoder
	var file *os.File
	if ConfigGlobal.RecordPath != "" {
		file, err := os.Create(ConfigGlobal.RecordPath)
		if err != nil {
			log.Error("Could not open commands output file ", err)
		} else {
			encoder = gob.NewEncoder(file)
		}
	}

	for {
		select {
		case <-r.quit:
			log.Debug("Closing router...")
			if file != nil {
				err := file.Close()
				if err != nil {
					log.Error("Could not close commands output file ", err)
				}
			}
			return
		case op := <-r.newOperation:
			// Check if this is an asset overview tab content operation and log it
			if assetOp, ok := op.(*operationAssetOverviewTabContent); ok {
				tabIDHex := ""
				if len(assetOp.TabID) > 0 {
					tabIDHex = hex.EncodeToString(assetOp.TabID)
				}
				log.Infof("Router received operationAssetOverviewTabContent with TabID=%s, Items=%d", 
					tabIDHex, len(assetOp.ItemIDs))
			}
			
			go op.Process(r.albionstate)
		case command := <-r.recordPhotonCommand:
			if encoder != nil {
				err := encoder.Encode(command)
				if err != nil {
					log.Error("Could not encode command ", err)
				}
			}
		}
	}
}

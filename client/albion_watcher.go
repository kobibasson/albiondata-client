package client

import (
	"fmt"
	"time"

	"github.com/ao-data/albiondata-client/log"
)

type albionProcessWatcher struct {
	known     []int
	devices   []string
	listeners map[int][]*listener
	quit      chan bool
	r         *Router
}

func newAlbionProcessWatcher() *albionProcessWatcher {
	return &albionProcessWatcher{
		listeners: make(map[int][]*listener),
		quit:      make(chan bool),
		r:         newRouter(),
	}
}

func (apw *albionProcessWatcher) run() error {
	log.Print("Watching Albion")
	physicalInterfaces, err := getAllPhysicalInterface()
	if err != nil {
		return err
	}
	apw.devices = physicalInterfaces
	log.Debugf("Will listen to these devices: %v", apw.devices)
	
	// Initialize inventory tracker if output path or webhook URL is provided
	if ConfigGlobal.InventoryOutputPath != "" || ConfigGlobal.InventoryWebhookURL != "" {
		fmt.Println("=================================================")
		
		// Initialize with output path if provided
		if ConfigGlobal.InventoryOutputPath != "" {
			fmt.Printf("INVENTORY TRACKING ENABLED: %s\n", ConfigGlobal.InventoryOutputPath)
			fmt.Println("Inventory changes will be displayed in real-time")
		}
		
		// Initialize with webhook if provided
		if ConfigGlobal.InventoryWebhookURL != "" {
			fmt.Printf("WEBHOOK ENABLED: Sending updates to %s\n", ConfigGlobal.InventoryWebhookURL)
			fmt.Println("Webhook debug information will be displayed in real-time")
			
			// Log if auth header is provided (without showing the actual value)
			if ConfigGlobal.InventoryWebhookAuthHeader != "" {
				authValue := ConfigGlobal.InventoryWebhookAuthHeader
				if len(authValue) > 10 {
					// Show first 5 and last 3 characters of the auth value
					authValue = authValue[:5] + "..." + authValue[len(authValue)-3:]
				}
				fmt.Printf("Authorization header will be added to webhook requests: %s\n", authValue)
			}
		}
		
		fmt.Println("=================================================")
		apw.r.albionstate.Inventory = NewPlayerInventory(
			ConfigGlobal.InventoryOutputPath, 
			ConfigGlobal.InventoryWebhookURL,
			ConfigGlobal.InventoryWebhookAuthHeader,
		)
	}
	
	go apw.r.run()

	for {
		select {
		case <-apw.quit:
			apw.closeWatcher()
			return nil
		default:
			if len(apw.listeners) == 0 {
				apw.createListeners()
			}
			time.Sleep(time.Second)
		}
	}
}

func (apw *albionProcessWatcher) closeWatcher() {
	log.Print("Albion watcher closed")

	for port := range apw.listeners {
		for _, l := range apw.listeners[port] {
			l.stop()
		}

		delete(apw.listeners, port)
	}

	apw.r.quit <- true
}

func (apw *albionProcessWatcher) createListeners() {
	filtered := [1]int{5056} // keep overdesign to listen on many ports

	for _, port := range filtered {
		for _, device := range apw.devices {
			l := newListener(apw.r)
			go l.startOnline(device, port)

			apw.listeners[port] = append(apw.listeners[port], l)
		}
	}
}

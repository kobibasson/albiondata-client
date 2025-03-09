package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// InventoryItem represents an item in the player's inventory
type InventoryItem struct {
	ItemID   int   `json:"item_id"`
	Quantity int   `json:"quantity"`
	LastSeen int64 `json:"last_seen"`
}

// PlayerInventory represents the player's inventory
type PlayerInventory struct {
	CharacterID   string                  `json:"character_id"`
	CharacterName string                  `json:"character_name"`
	Items         map[int]*InventoryItem  `json:"items"`
	LastUpdated   int64                   `json:"last_updated"`
	mutex         sync.Mutex
	outputPath    string
	// Flag to enable inventory change messages
	verboseOutput bool
	// Webhook URL to send inventory updates
	webhookURL    string
}

// NewPlayerInventory creates a new player inventory tracker
func NewPlayerInventory(outputPath string, webhookURL string) *PlayerInventory {
	return &PlayerInventory{
		Items:         make(map[int]*InventoryItem),
		outputPath:    outputPath,
		verboseOutput: true, // Enable verbose output by default when inventory tracking is enabled
		webhookURL:    webhookURL,
	}
}

// UpdateCharacterInfo updates the character information
func (pi *PlayerInventory) UpdateCharacterInfo(characterID, characterName string) {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	pi.CharacterID = characterID
	pi.CharacterName = characterName
	
	if pi.verboseOutput {
		fmt.Printf("[Inventory] Character info updated: %s (%s)\n", characterName, characterID)
	}
	
	// We don't need to send character info via webhook anymore since we're only tracking item changes
}

// WebhookPayload represents the data sent to the webhook
type WebhookPayload struct {
	ItemID    int   `json:"item_id"`
	Quantity  int   `json:"quantity"`
	Timestamp int64 `json:"timestamp"`
}

// sendWebhookUpdate sends an update to the configured webhook URL
func (pi *PlayerInventory) sendWebhookUpdate(action string, itemID int, quantity int, oldQuantity int) {
	if pi.webhookURL == "" {
		return
	}

	payload := WebhookPayload{
		ItemID:    itemID,
		Quantity:  quantity,
		Timestamp: time.Now().Unix(),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Failed to marshal webhook payload: %v", err)
		return
	}

	// Always print webhook debug info
	fmt.Printf("[Webhook] Sending: %s for item %d (Qty: %d)\n", 
		action, itemID, quantity)
	
	// Send the payload to the webhook URL
	resp, err := http.Post(pi.webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Errorf("Failed to send webhook: %v", err)
		fmt.Printf("[Webhook] ERROR: Failed to send webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Errorf("Webhook returned error status: %d", resp.StatusCode)
		fmt.Printf("[Webhook] ERROR: Server returned status code %d\n", resp.StatusCode)
	} else {
		fmt.Printf("[Webhook] SUCCESS: Server responded with status code %d\n", resp.StatusCode)
		
		// Read and print response body for debugging
		respBody, err := io.ReadAll(resp.Body)
		if err == nil && len(respBody) > 0 {
			fmt.Printf("[Webhook] Response: %s\n", string(respBody))
		}
	}
}

// AddOrUpdateItem adds or updates an item in the inventory
func (pi *PlayerInventory) AddOrUpdateItem(itemID int, quantity int) {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	now := time.Now().Unix()
	var action string
	var oldQuantity int
	
	// Handle negative quantities (likely due to byte overflow in the game client)
	originalQuantity := quantity
	if quantity < 0 {
		if quantity >= -127 {
			// For values between -1 and -127, add 256 to get the actual quantity
			// This is similar to how the auction code handles negative item amounts
			quantity = 256 + quantity
			log.Debugf("Converting negative quantity %d to %d for item %d (likely a stack overflow)", 
				originalQuantity, quantity, itemID)
		} else {
			// For very negative values, we're not sure how to interpret them
			// Log a warning but keep the original value
			log.Warnf("Received very negative quantity %d for item %d, keeping as is", 
				quantity, itemID)
		}
	}

	if item, exists := pi.Items[itemID]; exists {
		oldQuantity = item.Quantity
		action = "Updated"
		item.Quantity = quantity
		item.LastSeen = now
	} else {
		oldQuantity = 0
		action = "Added"
		pi.Items[itemID] = &InventoryItem{
			ItemID:   itemID,
			Quantity: quantity,
			LastSeen: now,
		}
	}

	pi.LastUpdated = now
	
	// Print inventory change message even without debug flags
	if pi.verboseOutput {
		var changeMsg string
		if action == "Added" {
			changeMsg = fmt.Sprintf("[Inventory] %s item: ID=%d, Quantity=%d", action, itemID, quantity)
		} else {
			diff := quantity - oldQuantity
			if diff > 0 {
				changeMsg = fmt.Sprintf("[Inventory] %s item: ID=%d, Quantity=%d (+%d)", action, itemID, quantity, diff)
			} else if diff < 0 {
				changeMsg = fmt.Sprintf("[Inventory] %s item: ID=%d, Quantity=%d (%d)", action, itemID, quantity, diff)
			} else {
				changeMsg = fmt.Sprintf("[Inventory] %s item: ID=%d, Quantity=%d (no change)", action, itemID, quantity)
			}
		}
		fmt.Println(changeMsg)
	}
	
	// Send webhook update
	if pi.webhookURL != "" {
		// For webhook, we'll send the corrected quantity but include the original in the debug message
		if originalQuantity < 0 && quantity != originalQuantity {
			fmt.Printf("[Webhook] Note: Converting negative quantity %d to %d for item %d\n", 
				originalQuantity, quantity, itemID)
		}
		pi.sendWebhookUpdate(action, itemID, quantity, oldQuantity)
	}
	
	// Only save to file if an output path is provided
	if pi.outputPath != "" {
		pi.SaveToFile()
	}
}

// RemoveItem removes an item from the inventory
func (pi *PlayerInventory) RemoveItem(itemID int) {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	var oldQuantity int
	if item, exists := pi.Items[itemID]; exists {
		if pi.verboseOutput {
			fmt.Printf("[Inventory] Removed item: ID=%d, Quantity was %d\n", itemID, item.Quantity)
		}
		oldQuantity = item.Quantity
		
		// Send webhook update
		if pi.webhookURL != "" {
			pi.sendWebhookUpdate("Removed", itemID, 0, oldQuantity)
		}
	}

	delete(pi.Items, itemID)
	pi.LastUpdated = time.Now().Unix()
	
	// Only save to file if an output path is provided
	if pi.outputPath != "" {
		pi.SaveToFile()
	}
}

// SetVerboseOutput sets whether to print verbose output
func (pi *PlayerInventory) SetVerboseOutput(verbose bool) {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()
	pi.verboseOutput = verbose
}

// SaveToFile saves the inventory to a JSON file
func (pi *PlayerInventory) SaveToFile() {
	if pi.outputPath == "" {
		return
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(pi.outputPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Errorf("Failed to create directory for inventory file: %v", err)
			return
		}
	}

	data, err := json.MarshalIndent(pi, "", "  ")
	if err != nil {
		log.Errorf("Failed to marshal inventory data: %v", err)
		return
	}

	err = os.WriteFile(pi.outputPath, data, 0644)
	if err != nil {
		log.Errorf("Failed to write inventory file: %v", err)
		return
	}

	if pi.verboseOutput {
		log.Debugf("Saved inventory to %s", pi.outputPath)
	}
} 
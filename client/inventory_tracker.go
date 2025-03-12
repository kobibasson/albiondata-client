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
	SlotID   int   `json:"slot_id"`    // Added SlotID field
	LastSeen int64 `json:"last_seen"`
}

// WebhookEvent represents a single inventory event to be sent to the webhook
type WebhookEvent struct {
	ItemID        int    `json:"item_id"`
	Quantity      int    `json:"quantity"`
	SlotID        int    `json:"slot_id"`
	Delta         int    `json:"delta"`
	Action        string `json:"action"`
	Timestamp     int64  `json:"timestamp"`
}

// WebhookPayload represents the data sent to the webhook
type WebhookPayload struct {
	Events        []WebhookEvent `json:"events"`
	CharacterID   string         `json:"character_id"`
	CharacterName string         `json:"character_name"`
	BatchTimestamp int64         `json:"batch_timestamp"`
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
	// Queue for pending webhook events
	pendingEvents []WebhookEvent
	// Timer for sending batched events
	batchTimer    *time.Timer
	// Mutex for the pending events queue
	eventMutex    sync.Mutex
}

// NewPlayerInventory creates a new player inventory tracker
func NewPlayerInventory(outputPath string, webhookURL string) *PlayerInventory {
	pi := &PlayerInventory{
		Items:         make(map[int]*InventoryItem),
		outputPath:    outputPath,
		verboseOutput: true, // Enable verbose output by default when inventory tracking is enabled
		webhookURL:    webhookURL,
		pendingEvents: make([]WebhookEvent, 0),
	}
	
	// Initialize the batch timer but don't start it yet
	pi.batchTimer = time.AfterFunc(3*time.Second, func() {
		pi.sendBatchedEvents()
	})
	pi.batchTimer.Stop() // Don't start the timer until we have events
	
	return pi
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
}

// queueWebhookEvent adds an event to the pending events queue and resets the timer
func (pi *PlayerInventory) queueWebhookEvent(action string, itemID int, quantity int, oldQuantity int, slotID int) {
	if pi.webhookURL == "" {
		return
	}

	// Calculate the delta (change in quantity)
	delta := quantity - oldQuantity

	event := WebhookEvent{
		ItemID:    itemID,
		Quantity:  quantity,
		SlotID:    slotID,
		Delta:     delta,
		Action:    action,
		Timestamp: time.Now().Unix(),
	}

	// Format delta with a sign for better visibility in logs
	var deltaStr string
	if delta > 0 {
		deltaStr = fmt.Sprintf("+%d", delta)
	} else {
		deltaStr = fmt.Sprintf("%d", delta)
	}

	// Always print webhook debug info
	fmt.Printf("[Webhook] Queuing: %s for item %d (Qty: %d, Delta: %s)\n", 
		action, itemID, quantity, deltaStr)

	// Add the event to the queue
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.pendingEvents = append(pi.pendingEvents, event)
	
	// Reset the timer to 3 seconds
	pi.batchTimer.Reset(3 * time.Second)
	
	fmt.Printf("[Webhook] Batch size: %d events (waiting 3s for more events)\n", len(pi.pendingEvents))
}

// sendBatchedEvents sends all pending events in a single webhook request
func (pi *PlayerInventory) sendBatchedEvents() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	// If there are no events, do nothing
	if len(pi.pendingEvents) == 0 {
		return
	}
	
	fmt.Printf("[Webhook] Sending batch of %d events\n", len(pi.pendingEvents))
	
	// Create the payload
	payload := WebhookPayload{
		Events:        pi.pendingEvents,
		CharacterID:   pi.CharacterID,
		CharacterName: pi.CharacterName,
		BatchTimestamp: time.Now().Unix(),
	}
	
	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Failed to marshal webhook payload: %v", err)
		return
	}
	
	// Print the full JSON payload with nice formatting
	prettyJSON, err := json.MarshalIndent(payload, "", "  ")
	if err == nil {
		fmt.Printf("[Webhook] Payload:\n%s\n", string(prettyJSON))
	} else {
		fmt.Printf("[Webhook] Payload: %s\n", string(jsonPayload))
	}
	
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
	
	// Clear the pending events
	pi.pendingEvents = make([]WebhookEvent, 0)
}

// AddOrUpdateItem adds or updates an item in the inventory
func (pi *PlayerInventory) AddOrUpdateItem(itemID int, quantity int, slotID int) {
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
		item.SlotID = slotID
		item.LastSeen = now
	} else {
		oldQuantity = 0
		action = "Added"
		pi.Items[itemID] = &InventoryItem{
			ItemID:   itemID,
			Quantity: quantity,
			SlotID:   slotID,
			LastSeen: now,
		}
	}

	pi.LastUpdated = now
	
	// Print inventory change message even without debug flags
	if pi.verboseOutput {
		var changeMsg string
		diff := quantity - oldQuantity
		
		// Format delta with a sign for better visibility
		var deltaStr string
		if diff > 0 {
			deltaStr = fmt.Sprintf("+%d", diff)
		} else if diff < 0 {
			deltaStr = fmt.Sprintf("%d", diff)
		} else {
			deltaStr = "±0"
		}
		
		if action == "Added" {
			changeMsg = fmt.Sprintf("[Inventory] %s item: ID=%d, Quantity=%d, SlotID=%d (Delta: %s)", 
				action, itemID, quantity, slotID, deltaStr)
		} else {
			changeMsg = fmt.Sprintf("[Inventory] %s item: ID=%d, Quantity=%d, SlotID=%d (Delta: %s)", 
				action, itemID, quantity, slotID, deltaStr)
		}
		fmt.Println(changeMsg)
	}
	
	// Queue webhook update
	if pi.webhookURL != "" {
		// For webhook, we'll send the corrected quantity but include the original in the debug message
		if originalQuantity < 0 && quantity != originalQuantity {
			fmt.Printf("[Webhook] Note: Converting negative quantity %d to %d for item %d\n", 
				originalQuantity, quantity, itemID)
		}
		pi.queueWebhookEvent(action, itemID, quantity, oldQuantity, slotID)
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
			// Calculate delta (always negative for removals)
			delta := -item.Quantity
			fmt.Printf("[Inventory] Removed item: ID=%d, Quantity was %d (Delta: %d)\n", 
				itemID, item.Quantity, delta)
		}
		oldQuantity = item.Quantity
		
		// Queue webhook update
		if pi.webhookURL != "" {
			pi.queueWebhookEvent("Removed", itemID, 0, oldQuantity, item.SlotID)
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
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// Global variables for bank vault information
var (
	lastBankVaultTime     int64
	lastBankVaultLocationID string
	bankVaultMutex        sync.Mutex
)

// Helper function to get absolute value of int64
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// InventoryItem represents an item in the player's inventory
type InventoryItem struct {
	ItemID   int   `json:"item_id"`
	Quantity int   `json:"quantity"`
	SlotID   int   `json:"slot_id"`    // Added SlotID field
	LastSeen int64 `json:"last_seen"`
}

// WebhookEvent represents a single inventory event
type WebhookEvent struct {
	ItemID     int    `json:"item_id"`
	Quantity   int    `json:"quantity"`
	Delta      int    `json:"delta"`      // Change in quantity (positive for increase, negative for decrease)
	Action     string `json:"action"`     // The action that triggered the webhook (Added, Updated, Removed)
	SlotID     int    `json:"slot_id,omitempty"` // Optional slot ID for equipment items
	Timestamp  int64  `json:"timestamp"`
}

// WebhookPayload represents the data sent to the webhook
type WebhookPayload struct {
	Events        []WebhookEvent `json:"events"`
	CharacterID   string         `json:"character_id"`
	CharacterName string         `json:"character_name"`
	BatchTimestamp int64         `json:"batch_timestamp"`
	Override      bool           `json:"override"`      // Flag to indicate if this batch should override previous data
	Bank          bool           `json:"bank"`          // Flag to indicate if the events are related to a bank
	LocationID    string         `json:"location_id,omitempty"`   // ID of the bank location
}

// SingleItemWebhookPayload represents a single item update sent to the webhook
type SingleItemWebhookPayload struct {
	ItemID        int    `json:"item_id"`
	Quantity      int    `json:"quantity"`
	Delta         int    `json:"delta"`      // Change in quantity (positive for increase, negative for decrease)
	Action        string `json:"action"`     // The action that triggered the webhook (Added, Updated, Removed)
	CharacterID   string `json:"character_id"`   // The ID of the character whose inventory was updated
	CharacterName string `json:"character_name"` // The name of the character whose inventory was updated
	SlotID        int    `json:"slot_id,omitempty"` // Optional slot ID for equipment items
	IsBank        bool   `json:"is_bank,omitempty"` // Whether this item is in a bank vault
	LocationID    string `json:"location_id,omitempty"` // Bank vault location ID if applicable
	Timestamp     int64  `json:"timestamp"`
	Override      bool   `json:"override"`  // Flag to indicate if this update should override previous data
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
	// Timestamp of the last cluster change
	lastClusterChange time.Time
	// Timestamp of the last join operation
	lastJoinTime      time.Time
	// Timestamp of the last leave event
	lastLeaveTime     time.Time
	// Timestamp of the last bank access
	lastBankAccess    time.Time
	// Current bank location
	currentBankLocation string
	// Flag to indicate if the player is in a bank
	isBank             bool
	// Timestamp of the last event added to the batch
	lastEventTime      time.Time
}

// NewPlayerInventory creates a new player inventory tracker
func NewPlayerInventory(outputPath string, webhookURL string) *PlayerInventory {
	pi := &PlayerInventory{
		Items:         make(map[int]*InventoryItem),
		outputPath:    outputPath,
		verboseOutput: true, // Enable verbose output by default when inventory tracking is enabled
		webhookURL:    webhookURL,
		pendingEvents: make([]WebhookEvent, 0),
		lastClusterChange: time.Now(),
		lastJoinTime:      time.Now(),
		lastLeaveTime:     time.Now(),
		lastBankAccess:    time.Now(),
		lastEventTime:     time.Now(),
	}
	
	// Initialize the batch timer but don't start it yet
	pi.batchTimer = time.AfterFunc(10*time.Second, func() {
		pi.sendBatchedEvents()
	})
	pi.batchTimer.Stop() // Don't start the timer until we have events
	
	return pi
}

// OnClusterChange is called when a cluster change operation is detected
func (pi *PlayerInventory) OnClusterChange() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.lastClusterChange = time.Now()
	fmt.Printf("[CLUSTER] Changed at %v\n", pi.lastClusterChange)
	
	// Print debug state after cluster change
	pi.mutex.Lock()
	defer pi.mutex.Unlock()
	
	fmt.Printf("\n[DEBUG] Inventory Tracker State after cluster change:\n")
	fmt.Printf("  Character: %s (%s)\n", pi.CharacterName, pi.CharacterID)
	fmt.Printf("  Items: %d\n", len(pi.Items))
	fmt.Printf("  Webhook URL: %s\n", pi.webhookURL)
	fmt.Printf("  Webhook Enabled: %v\n", pi.webhookURL != "")
	fmt.Printf("  Pending Events: %d\n", len(pi.pendingEvents))
}

// OnJoin is called when a join operation is detected
func (pi *PlayerInventory) OnJoin() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.lastJoinTime = time.Now()
	log.Debugf("Inventory tracker notified of join operation at timestamp %v", pi.lastJoinTime)
}

// OnLeave is called when a leave event is detected
func (pi *PlayerInventory) OnLeave() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.lastLeaveTime = time.Now()
	log.Debugf("Inventory tracker notified of leave event at timestamp %v", pi.lastLeaveTime)
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
	// Call the new method with default values for bank information
	pi.queueWebhookEventWithBank(action, itemID, quantity, oldQuantity, slotID, false, "")
}

// sendBatchedEvents sends all pending events in a single webhook request
func (pi *PlayerInventory) sendBatchedEvents() {
	if pi.webhookURL == "" {
		return
	}
	
	if len(pi.pendingEvents) == 0 {
		return
	}

	// Check if we're within 4 seconds of a cluster change
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	if timeSinceClusterChange <= 4*time.Second {
		fmt.Printf("[WEBHOOK BATCH] Skipping batch send - too close to cluster change (%v)\n", timeSinceClusterChange)
		return
	}

	fmt.Printf("[WEBHOOK BATCH] Sending batch of %d events\n", len(pi.pendingEvents))

	// Check for equipment items in the batch
	hasEquipmentItems := false
	for _, event := range pi.pendingEvents {
		if event.SlotID >= 1000000 { // Equipment items typically have slot IDs over 1,000,000
			hasEquipmentItems = true
			break
		}
	}
	
	if hasEquipmentItems {
		fmt.Printf("[WEBHOOK BATCH] Batch includes equipment items (slots > 1,000,000)\n")
	} else {
		fmt.Printf("[WEBHOOK BATCH] Batch contains only inventory items (no equipment)\n")
	}

	// Determine if we should set override flag
	// Use the time between cluster change and the last event added to the batch
	timeBetweenClusterAndLastEvent := pi.lastEventTime.Sub(pi.lastClusterChange)
	
	// Updated override logic: true if last event occurred within 4 sec of cluster change, false otherwise
	override := timeBetweenClusterAndLastEvent <= 4*time.Second
	
	if override {
		fmt.Printf("[WEBHOOK BATCH] Override=true: Last event occurred within 4s of cluster change (%v after)\n", 
			timeBetweenClusterAndLastEvent)
	} else {
		fmt.Printf("[WEBHOOK BATCH] Override=false: Last event occurred more than 4s after cluster change (%v after)\n", 
			timeBetweenClusterAndLastEvent)
	}

	payload := WebhookPayload{
		Events:         pi.pendingEvents,
		CharacterID:    pi.CharacterID,
		CharacterName:  pi.CharacterName,
		BatchTimestamp: time.Now().Unix(),
		Override:       override,
		Bank:          pi.isBank,
	}

	if pi.isBank && pi.currentBankLocation != "" {
		payload.LocationID = pi.currentBankLocation
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Failed to marshal batched webhook payload: %v", err)
		return
	}

	// Send the payload to the webhook URL
	resp, err := http.Post(pi.webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Errorf("Failed to send batched webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		log.Errorf("Batched webhook returned error status: %d", resp.StatusCode)
	} else {
		fmt.Printf("[WEBHOOK BATCH] Successfully sent batch of %d events\n", len(pi.pendingEvents))
	}

	// Clear pending events after send
	pi.pendingEvents = nil
}

// AddOrUpdateItem adds or updates an item in the inventory
func (pi *PlayerInventory) AddOrUpdateItem(itemID int, quantity int, slotID int) {
	// Call the new method with default values for bank information
	pi.AddOrUpdateItemWithBank(itemID, quantity, slotID, false, "")
}

// RemoveItem removes an item from the inventory
func (pi *PlayerInventory) RemoveItem(itemID int) {
	// Call the new method with default values for bank information
	pi.RemoveItemWithBank(itemID, false, "")
}

// RemoveItemWithBank removes an item from the inventory with bank information
func (pi *PlayerInventory) RemoveItemWithBank(itemID int, isBank bool, locationID string) {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	var oldQuantity int
	var slotID int
	if item, exists := pi.Items[itemID]; exists {
		if pi.verboseOutput {
			// Calculate delta (always negative for removals)
			delta := -item.Quantity
			
			// Create message
			removeMsg := fmt.Sprintf("[Inventory] Removed item: ID=%d, Quantity was %d (Delta: %d)", 
				itemID, item.Quantity, delta)
			
			// Add bank information if applicable
			if isBank {
				removeMsg += fmt.Sprintf(" [Bank: %s]", locationID)
			}
			
			fmt.Println(removeMsg)
		}
		oldQuantity = item.Quantity
		slotID = item.SlotID
		
		// Queue webhook update
		if pi.webhookURL != "" {
			pi.queueWebhookEventWithBank("Removed", itemID, 0, oldQuantity, slotID, isBank, locationID)
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

// AddOrUpdateItemWithBank adds or updates an item in the inventory with bank information
func (pi *PlayerInventory) AddOrUpdateItemWithBank(itemID int, quantity int, slotID int, isBank bool, locationID string) {
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
	
	// Queue webhook update
	if pi.webhookURL != "" {
		// For webhook, we'll send the corrected quantity but include the original in the debug message
		if originalQuantity < 0 && quantity != originalQuantity {
			log.Debugf("Converting negative quantity %d to %d for item %d", 
				originalQuantity, quantity, itemID)
		}
		pi.queueWebhookEventWithBank(action, itemID, quantity, oldQuantity, slotID, isBank, locationID)
	}
	
	// Only save to file if an output path is provided
	if pi.outputPath != "" {
		pi.SaveToFile()
	}
}

func (pi *PlayerInventory) queueWebhookEventWithBank(action string, itemID int, quantity int, oldQuantity int, slotID int, isBank bool, locationID string) {
	if pi.webhookURL == "" {
		return
	}

	// Calculate the delta (change in quantity)
	delta := quantity - oldQuantity

	event := WebhookEvent{
		ItemID:     itemID,
		Quantity:   quantity,
		SlotID:     slotID,
		Delta:      delta,
		Action:     action,
		Timestamp:  time.Now().Unix(),
	}

	// Add the event to the queue
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.pendingEvents = append(pi.pendingEvents, event)
	pi.lastEventTime = time.Now() // Track when the last event was added
	
	// If we're within 10 seconds of a cluster change, use a shorter batch timer (2 seconds)
	// Otherwise, use the standard 10 second timer
	var waitTime time.Duration
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	if timeSinceClusterChange <= 10*time.Second {
		waitTime = 2 * time.Second
		fmt.Printf("[WEBHOOK] Batch size: %d events (waiting 2s for more events - after cluster change)\n", len(pi.pendingEvents))
	} else {
		waitTime = 10 * time.Second
		fmt.Printf("[WEBHOOK] Batch size: %d events (waiting 10s for more events)\n", len(pi.pendingEvents))
	}
	
	// Reset the timer with the appropriate wait time
	pi.batchTimer.Reset(waitTime)
}

// sendWebhookUpdate sends an update to the configured webhook URL
func (pi *PlayerInventory) sendWebhookUpdate(action string, itemID int, quantity int, oldQuantity int) {
	if pi.webhookURL == "" {
		return
	}

	// Calculate the delta (change in quantity)
	delta := quantity - oldQuantity

	// Determine if we should set override flag
	// For single item updates (equipment items), we set override=true if they occur within 4 seconds of a cluster change
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	override := timeSinceClusterChange <= 4*time.Second
	
	payload := SingleItemWebhookPayload{
		ItemID:        itemID,
		Quantity:      quantity,
		Delta:         delta,
		Action:        action,
		CharacterID:   pi.CharacterID,
		CharacterName: pi.CharacterName,
		Timestamp:     time.Now().Unix(),
		Override:      override,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Failed to marshal webhook payload: %v", err)
		return
	}
	
	// Send the payload to the webhook URL
	resp, err := http.Post(pi.webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Errorf("Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		log.Errorf("Webhook returned error status: %d", resp.StatusCode)
	} else {
		fmt.Printf("[WEBHOOK] Successfully sent single item update for ID=%d\n", itemID)
	}
}

// NotifyBankVaultAccess updates the bank vault access time and location
func (pi *PlayerInventory) NotifyBankVaultAccess(locationID string) {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.lastBankAccess = time.Now()
	pi.currentBankLocation = locationID
	pi.isBank = true
	log.Debugf("Inventory tracker notified of bank vault access at location %s at timestamp %v", locationID, pi.lastBankAccess)
}

// NotifyLeaveBankVault updates the state when leaving a bank vault
func (pi *PlayerInventory) NotifyLeaveBankVault() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.isBank = false
	pi.currentBankLocation = ""
	log.Debugf("Inventory tracker notified of leaving bank vault at timestamp %v", time.Now())
}

// PrintDebugState prints the current state of the inventory tracker for debugging
func (pi *PlayerInventory) PrintDebugState() {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()
	
	fmt.Printf("\n[DEBUG] Inventory Tracker State:\n")
	fmt.Printf("  Character: %s (%s)\n", pi.CharacterName, pi.CharacterID)
	fmt.Printf("  Items: %d\n", len(pi.Items))
	fmt.Printf("  Last Updated: %v\n", time.Unix(pi.LastUpdated, 0))
	fmt.Printf("  Webhook URL: %s\n", pi.webhookURL)
	fmt.Printf("  Webhook Enabled: %v\n", pi.webhookURL != "")
	fmt.Printf("  Pending Events: %d\n", len(pi.pendingEvents))
	fmt.Printf("  Last Cluster Change: %v (%v ago)\n", pi.lastClusterChange, time.Since(pi.lastClusterChange))
	fmt.Printf("  Last Join Time: %v (%v ago)\n", pi.lastJoinTime, time.Since(pi.lastJoinTime))
	fmt.Printf("  Last Leave Time: %v (%v ago)\n", pi.lastLeaveTime, time.Since(pi.lastLeaveTime))
	fmt.Printf("  Last Bank Access: %v (%v ago)\n", pi.lastBankAccess, time.Since(pi.lastBankAccess))
	fmt.Printf("  Is Bank: %v\n", pi.isBank)
	if pi.isBank {
		fmt.Printf("  Current Bank Location: %s\n", pi.currentBankLocation)
	}
}

// NotifyClusterChange updates the cluster change timestamp
func (pi *PlayerInventory) NotifyClusterChange() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.lastClusterChange = time.Now()
	fmt.Printf("[CLUSTER] Changed at %v\n", pi.lastClusterChange)
	
	// Print debug state after cluster change
	pi.PrintDebugState()
} 
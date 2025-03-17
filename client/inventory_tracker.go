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
	Equipped   bool   `json:"equipped,omitempty"` // Whether the equipment item is equipped
	Timestamp  int64  `json:"timestamp"`
	LocationID string `json:"location_id,omitempty"` // ID of the bank location for this item
}

// WebhookPayload represents the data sent to the webhook
type WebhookPayload struct {
	Events        []WebhookEvent `json:"events"`
	CharacterID   string         `json:"character_id"`
	CharacterName string         `json:"character_name"`
	BatchTimestamp int64         `json:"batch_timestamp"`
	Override      bool           `json:"override"`      // Flag to indicate if this batch should override previous data
	Bank          bool           `json:"bank"`          // Flag to indicate if the events are related to a bank
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
	Equipped      bool   `json:"equipped,omitempty"` // Whether the equipment item is equipped
	IsBank        bool   `json:"is_bank,omitempty"` // Whether this item is in a bank
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
	// Add a retry counter
	webhookRetryCount int
}

// NewPlayerInventory creates a new player inventory tracker
func NewPlayerInventory(outputPath string, webhookURL string) *PlayerInventory {
	pi := &PlayerInventory{
		Items:          make(map[int]*InventoryItem),
		outputPath:     outputPath,
		verboseOutput:  true, // Enable verbose output by default when inventory tracking is enabled
		webhookURL:     webhookURL,
		pendingEvents:  make([]WebhookEvent, 0),
		webhookRetryCount: 0, // Initialize retry counter
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
	fmt.Printf("\n[WEBHOOK BATCH] ===== ATTEMPTING TO SEND BATCHED EVENTS =====\n")
	log.Infof("[WEBHOOK BATCH] Starting batched events send with %d events", len(pi.pendingEvents))
	
	if pi.webhookURL == "" {
		fmt.Printf("[WEBHOOK ERROR] Cannot send batched events: webhook URL is empty\n")
		log.Error("[WEBHOOK ERROR] Cannot send batched events: webhook URL is empty")
		fmt.Printf("[WEBHOOK BATCH] ===== END BATCHED EVENTS (FAILED) =====\n\n")
		return
	}
	
	if len(pi.pendingEvents) == 0 {
		fmt.Printf("[WEBHOOK ERROR] Cannot send batched events: no pending events\n")
		log.Error("[WEBHOOK ERROR] Cannot send batched events: no pending events")
		fmt.Printf("[WEBHOOK BATCH] ===== END BATCHED EVENTS (FAILED) =====\n\n")
		return
	}

	// Check if we're within 4 seconds of a cluster change
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	if timeSinceClusterChange <= 4*time.Second {
		fmt.Printf("[WEBHOOK BATCH] Skipping batch send - too close to cluster change (%v)\n", timeSinceClusterChange)
		log.Infof("[WEBHOOK BATCH] Skipping batch send - too close to cluster change (%v)", timeSinceClusterChange)
		fmt.Printf("[WEBHOOK BATCH] ===== END BATCHED EVENTS (SKIPPED) =====\n\n")
		return
	}

	// Check if it's been more than 3 seconds since the last bank access
	// If so, automatically reset the bank state
	if pi.isBank {
		timeSinceBankAccess := time.Since(pi.lastBankAccess)
		if timeSinceBankAccess > 3*time.Second {
			fmt.Printf("[WEBHOOK BATCH] Auto-resetting bank state - it's been %v since last bank access\n", timeSinceBankAccess)
			log.Infof("[WEBHOOK BATCH] Auto-resetting bank state - it's been %v since last bank access", timeSinceBankAccess)
			pi.isBank = false
			pi.currentBankLocation = ""
		}
	}

	fmt.Printf("[WEBHOOK BATCH] Sending batch of %d events\n", len(pi.pendingEvents))
	fmt.Printf("[WEBHOOK BATCH] Character: %s (%s)\n", pi.CharacterName, pi.CharacterID)
	fmt.Printf("[WEBHOOK BATCH] Bank state: isBank=%v, location=%s\n", pi.isBank, pi.currentBankLocation)
	
	log.Infof("[WEBHOOK BATCH] Sending batch of %d events for character %s (%s)", 
		len(pi.pendingEvents), pi.CharacterName, pi.CharacterID)
	log.Infof("[WEBHOOK BATCH] Bank state: isBank=%v, location=%s", 
		pi.isBank, pi.currentBankLocation)

	// Check for equipment items in the batch
	hasEquipmentItems := false
	equippedItems := 0
	for _, event := range pi.pendingEvents {
		if event.SlotID >= 1000000 { // Equipment items typically have slot IDs over 1,000,000
			hasEquipmentItems = true
			if event.Equipped {
				equippedItems++
			}
		}
	}
	
	if hasEquipmentItems {
		fmt.Printf("[WEBHOOK BATCH] Batch includes %d equipment items (%d equipped, slots > 1,000,000)\n", 
			countEquipmentItems(pi.pendingEvents), equippedItems)
		log.Infof("[WEBHOOK BATCH] Batch includes %d equipment items (%d equipped)", 
			countEquipmentItems(pi.pendingEvents), equippedItems)
	} else {
		fmt.Printf("[WEBHOOK BATCH] Batch contains only inventory items (no equipment)\n")
		log.Infof("[WEBHOOK BATCH] Batch contains only inventory items (no equipment)")
	}

	// Determine if we should set override flag
	// Use the time between cluster change and the last event added to the batch
	timeBetweenClusterAndLastEvent := pi.lastEventTime.Sub(pi.lastClusterChange)
	
	// Updated override logic: true if last event occurred within 4 sec of cluster change, false otherwise
	override := timeBetweenClusterAndLastEvent <= 4*time.Second
	
	if override {
		fmt.Printf("[WEBHOOK BATCH] Override=true: Last event occurred within 4s of cluster change (%v after)\n", 
			timeBetweenClusterAndLastEvent)
		log.Infof("[WEBHOOK BATCH] Override=true: Last event occurred within 4s of cluster change (%v after)", 
			timeBetweenClusterAndLastEvent)
	} else {
		fmt.Printf("[WEBHOOK BATCH] Override=false: Last event occurred more than 4s after cluster change (%v after)\n", 
			timeBetweenClusterAndLastEvent)
		log.Infof("[WEBHOOK BATCH] Override=false: Last event occurred more than 4s after cluster change (%v after)", 
			timeBetweenClusterAndLastEvent)
	}

	payload := WebhookPayload{
		Events:         pi.pendingEvents,
		CharacterID:    pi.CharacterID,
		CharacterName:  pi.CharacterName,
		BatchTimestamp: time.Now().Unix(),
		Override:       override,
		Bank:           pi.isBank,
	}

	if pi.isBank && pi.currentBankLocation != "" {
		fmt.Printf("[WEBHOOK BATCH] Including location ID in events: %s\n", pi.currentBankLocation)
		log.Infof("[WEBHOOK BATCH] Including location ID in events: %s", pi.currentBankLocation)
		
		// Update the location ID in each event
		for i := range pi.pendingEvents {
			if pi.pendingEvents[i].LocationID == "" {
				pi.pendingEvents[i].LocationID = pi.currentBankLocation
			}
		}
	}

	// Don't print the full payload to reduce log verbosity
	fmt.Printf("[WEBHOOK PAYLOAD] Batched events payload contains %d events (not shown to reduce verbosity)\n", len(pi.pendingEvents))
	log.Debugf("[WEBHOOK PAYLOAD] Batched events payload contains %d events", len(pi.pendingEvents))

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Failed to marshal batched webhook payload: %v", err)
		fmt.Printf("[WEBHOOK ERROR] Failed to marshal batched webhook payload: %v\n", err)
		fmt.Printf("[WEBHOOK BATCH] ===== END BATCHED EVENTS (FAILED) =====\n\n")
		
		// Increment retry counter and check if we should keep retrying
		pi.webhookRetryCount++
		if pi.webhookRetryCount > 3 {
			log.Warnf("[WEBHOOK RETRY] Exceeded maximum retry count (%d), clearing %d pending events", 
				pi.webhookRetryCount, len(pi.pendingEvents))
			fmt.Printf("[WEBHOOK RETRY] Exceeded maximum retry count (%d), clearing %d pending events\n", 
				pi.webhookRetryCount, len(pi.pendingEvents))
			pi.pendingEvents = nil
			pi.webhookRetryCount = 0
		} else {
			// Log that we're keeping the events in the queue for retry
			log.Infof("[WEBHOOK RETRY] Keeping %d events in queue for retry after timeout error (attempt %d/3)", 
				len(pi.pendingEvents), pi.webhookRetryCount)
			fmt.Printf("[WEBHOOK RETRY] Keeping %d events in queue for retry after timeout error (attempt %d/3)\n", 
				len(pi.pendingEvents), pi.webhookRetryCount)
		}
		
		return
	}

	fmt.Printf("[WEBHOOK REQUEST] Sending POST request to: %s\n", pi.webhookURL)
	log.Infof("[WEBHOOK REQUEST] Sending POST request to: %s with %d bytes", pi.webhookURL, len(jsonPayload))
	
	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Create the request
	req, err := http.NewRequest("POST", pi.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Errorf("Failed to create request for batched events: %v", err)
		fmt.Printf("[WEBHOOK ERROR] Failed to create request for batched events: %v\n", err)
		fmt.Printf("[WEBHOOK BATCH] ===== END BATCHED EVENTS (FAILED) =====\n\n")
		return
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlbionData-Client")
	
	// Send the request
	startTime := time.Now()
	resp, err := client.Do(req)
	requestDuration := time.Since(startTime)
	
	if err != nil {
		log.Errorf("Failed to send batched webhook: %v", err)
		fmt.Printf("[WEBHOOK ERROR] Failed to send batched webhook: %v\n", err)
		fmt.Printf("[WEBHOOK BATCH] ===== END BATCHED EVENTS (FAILED) =====\n\n")
		
		// Increment retry counter and check if we should keep retrying
		pi.webhookRetryCount++
		if pi.webhookRetryCount > 3 {
			log.Warnf("[WEBHOOK RETRY] Exceeded maximum retry count (%d), clearing %d pending events", 
				pi.webhookRetryCount, len(pi.pendingEvents))
			fmt.Printf("[WEBHOOK RETRY] Exceeded maximum retry count (%d), clearing %d pending events\n", 
				pi.webhookRetryCount, len(pi.pendingEvents))
			pi.pendingEvents = nil
			pi.webhookRetryCount = 0
		} else {
			// Log that we're keeping the events in the queue for retry
			log.Infof("[WEBHOOK RETRY] Keeping %d events in queue for retry after timeout error (attempt %d/3)", 
				len(pi.pendingEvents), pi.webhookRetryCount)
			fmt.Printf("[WEBHOOK RETRY] Keeping %d events in queue for retry after timeout error (attempt %d/3)\n", 
				len(pi.pendingEvents), pi.webhookRetryCount)
		}
		
		return
	}
	defer resp.Body.Close()

	// Read and print the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
		fmt.Printf("[WEBHOOK ERROR] Failed to read response body: %v\n", err)
	}
	
	// Log request timing
	fmt.Printf("[WEBHOOK TIMING] Request completed in %v\n", requestDuration)
	log.Infof("[WEBHOOK TIMING] Request completed in %v", requestDuration)
	
	// Check response status
	if resp.StatusCode >= 400 {
		log.Errorf("Batched webhook returned error status: %d", resp.StatusCode)
		fmt.Printf("[WEBHOOK ERROR] Batched webhook returned error status: %d\n", resp.StatusCode)
		fmt.Printf("[WEBHOOK RESPONSE] Error response body: %s\n", string(respBody))
		log.Errorf("[WEBHOOK RESPONSE] Error response body: %s", string(respBody))
		fmt.Printf("[WEBHOOK BATCH] ===== END BATCHED EVENTS (FAILED) =====\n\n")
	} else {
		fmt.Printf("[WEBHOOK SUCCESS] Successfully sent batch of %d events\n", len(pi.pendingEvents))
		fmt.Printf("[WEBHOOK RESPONSE] Response status: %d\n", resp.StatusCode)
		// Don't print the full response body to reduce log verbosity
		fmt.Printf("[WEBHOOK RESPONSE] Response received (not shown to reduce verbosity)\n")
		log.Infof("[WEBHOOK SUCCESS] Successfully sent batch of %d events", len(pi.pendingEvents))
		log.Debugf("[WEBHOOK RESPONSE] Response status: %d, body: %s", resp.StatusCode, string(respBody))
		fmt.Printf("[WEBHOOK BATCH] ===== END BATCHED EVENTS (SUCCESS) =====\n\n")
	}

	// Clear pending events after send
	pi.pendingEvents = nil
	log.Infof("[WEBHOOK BATCH] Cleared pending events queue after sending batch")

	// Reset retry counter on success
	pi.webhookRetryCount = 0
}

// countEquipmentItems counts the number of equipment items in a slice of WebhookEvents
func countEquipmentItems(events []WebhookEvent) int {
	count := 0
	for _, event := range events {
		if event.SlotID >= 1000000 { // Equipment items typically have slot IDs over 1,000,000
			count++
		}
	}
	return count
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
		
		// Check if a bank vault info event occurred within 2 seconds
		bankVaultMutex.Lock()
		currentBankVaultTime := lastBankVaultTime
		bankVaultMutex.Unlock()
		
		timeSinceLastBankVault := now - currentBankVaultTime
		if timeSinceLastBankVault <= 2 && !isBank {
			log.Debugf("Skipping webhook queue for item %d: occurred within 2 seconds of evBankVaultInfo event (time diff: %d sec)", 
				itemID, timeSinceLastBankVault)
		} else {
			pi.queueWebhookEventWithBank(action, itemID, quantity, oldQuantity, slotID, isBank, locationID)
		}
	}
	
	// Only save to file if an output path is provided
	if pi.outputPath != "" {
		pi.SaveToFile()
	}
}

func (pi *PlayerInventory) queueWebhookEventWithBank(action string, itemID int, quantity int, oldQuantity int, slotID int, isBank bool, locationID string) {
	if pi.webhookURL == "" {
		fmt.Printf("[WEBHOOK ERROR] Cannot queue webhook event: webhook URL is empty\n")
		return
	}

	// Calculate the delta (change in quantity)
	delta := quantity - oldQuantity
	
	fmt.Printf("[WEBHOOK QUEUE] Queueing event: Action=%s, ItemID=%d, Quantity=%d, Delta=%d, SlotID=%d, IsBank=%v, LocationID=%s\n",
		action, itemID, quantity, delta, slotID, isBank, locationID)

	event := WebhookEvent{
		ItemID:     itemID,
		Quantity:   quantity,
		SlotID:     slotID,
		Delta:      delta,
		Action:     action,
		Equipped:   false,
		Timestamp:  time.Now().Unix(),
		LocationID: locationID,
	}

	// Add the event to the queue
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.pendingEvents = append(pi.pendingEvents, event)
	pi.lastEventTime = time.Now() // Track when the last event was added
	
	fmt.Printf("[WEBHOOK QUEUE] Added event to queue. Queue size now: %d events\n", len(pi.pendingEvents))
	
	// If we're within 10 seconds of a cluster change, use a shorter batch timer (2 seconds)
	// Otherwise, use the standard 10 second timer
	var waitTime time.Duration
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	if timeSinceClusterChange <= 10*time.Second {
		waitTime = 2 * time.Second
		fmt.Printf("[WEBHOOK TIMER] Setting batch timer to 2s (within 10s of cluster change)\n")
	} else {
		waitTime = 10 * time.Second
		fmt.Printf("[WEBHOOK TIMER] Setting batch timer to 10s (normal operation)\n")
	}
	
	// Reset the timer with the appropriate wait time
	oldTimerActive := pi.batchTimer.Stop()
	pi.batchTimer.Reset(waitTime)
	
	fmt.Printf("[WEBHOOK TIMER] Timer reset: waitTime=%v, oldTimerActive=%v\n", waitTime, oldTimerActive)
}

// sendWebhookUpdate sends an update to the configured webhook URL
func (pi *PlayerInventory) sendWebhookUpdate(action string, itemID int, quantity int, oldQuantity int, slotID int, equipped bool) {
	if pi.webhookURL == "" {
		return
	}

	// Calculate the delta (change in quantity)
	delta := quantity - oldQuantity

	// Determine if we should set override flag
	// For single item updates (equipment items), we set override=true if they occur within 4 seconds of a cluster change
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	override := timeSinceClusterChange <= 4*time.Second
	
	// Check if we're in a bank
	isBank := pi.isBank
	locationID := pi.currentBankLocation
	
	payload := SingleItemWebhookPayload{
		ItemID:        itemID,
		Quantity:      quantity,
		Delta:         delta,
		Action:        action,
		CharacterID:   pi.CharacterID,
		CharacterName: pi.CharacterName,
		SlotID:        slotID,
		Equipped:      equipped,
		IsBank:        isBank,
		LocationID:    locationID,
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
	
	// Store previous state for logging
	prevIsBank := pi.isBank
	prevLocation := pi.currentBankLocation
	
	// Update state
	pi.lastBankAccess = time.Now()
	pi.currentBankLocation = locationID
	pi.isBank = true
	
	fmt.Printf("\n[BANK ACCESS] ===== DETECTED BANK VAULT ACCESS =====\n")
	fmt.Printf("[BANK ACCESS] Location ID: %s\n", locationID)
	fmt.Printf("[BANK ACCESS] Timestamp: %v\n", pi.lastBankAccess)
	fmt.Printf("[BANK ACCESS] Previous bank state: isBank=%v, location=%s\n", 
		prevIsBank, prevLocation)
	fmt.Printf("[BANK ACCESS] Webhook URL: %s\n", pi.webhookURL)
	fmt.Printf("[BANK ACCESS] Webhook Enabled: %v\n", pi.webhookURL != "")
	fmt.Printf("[BANK ACCESS] Pending Events: %d\n", len(pi.pendingEvents))
	fmt.Printf("[BANK ACCESS] ===== END BANK VAULT ACCESS =====\n\n")
	
	log.Infof("[BANK ACCESS] Detected bank vault access at location %s", locationID)
	log.Infof("[BANK ACCESS] Previous bank state: isBank=%v, location=%s", prevIsBank, prevLocation)
	log.Infof("[BANK ACCESS] Current state: Webhook URL=%s, Enabled=%v, Pending Events=%d", 
		pi.webhookURL, pi.webhookURL != "", len(pi.pendingEvents))
	
	// Update global bank vault information
	bankVaultMutex.Lock()
	defer bankVaultMutex.Unlock()
	
	lastBankVaultTime = time.Now().Unix()
	lastBankVaultLocationID = locationID
	
	log.Infof("[BANK ACCESS] Updated global bank vault information: time=%d, location=%s", 
		lastBankVaultTime, lastBankVaultLocationID)
	
	// Check if we need to reset the batch timer for faster processing
	if pi.webhookURL != "" && len(pi.pendingEvents) > 0 {
		// For bank access, we want to process events quickly
		oldTimerActive := pi.batchTimer.Stop()
		pi.batchTimer.Reset(2 * time.Second)
		log.Infof("[BANK ACCESS] Reset batch timer to 2s (oldTimerActive=%v)", oldTimerActive)
		fmt.Printf("[BANK ACCESS] Reset batch timer to 2s for faster processing\n")
	}
}

// NotifyLeaveBankVault updates the state when leaving a bank vault
func (pi *PlayerInventory) NotifyLeaveBankVault() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.isBank = false
	pi.currentBankLocation = ""
	log.Infof("Inventory tracker notified of leaving bank vault at timestamp %v", time.Now())
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

// AddOrUpdateEquipmentItem adds or updates an equipment item in the inventory
func (pi *PlayerInventory) AddOrUpdateEquipmentItem(itemID int, quantity int, slotID int, equipped bool, isBank bool, locationID string) {
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
	
	// Queue webhook update with equipped information
	if pi.webhookURL != "" {
		// For webhook, we'll send the corrected quantity but include the original in the debug message
		if originalQuantity < 0 && quantity != originalQuantity {
			log.Debugf("Converting negative quantity %d to %d for item %d", 
				originalQuantity, quantity, itemID)
		}
		pi.queueWebhookEventWithEquipped(action, itemID, quantity, oldQuantity, slotID, equipped, isBank, locationID)
	}
	
	// Only save to file if an output path is provided
	if pi.outputPath != "" {
		pi.SaveToFile()
	}
}

// queueWebhookEventWithEquipped adds an event to the pending events queue with equipped information
func (pi *PlayerInventory) queueWebhookEventWithEquipped(action string, itemID int, quantity int, oldQuantity int, slotID int, equipped bool, isBank bool, locationID string) {
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
		Equipped:   equipped,
		Timestamp:  time.Now().Unix(),
		LocationID: locationID,
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

// AddOrUpdateBankItem adds or updates a bank item in the inventory
func (pi *PlayerInventory) AddOrUpdateBankItem(itemID int, quantity int, slotID int, tabName string, locationID string) {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	fmt.Printf("[BANK ITEM] Adding/updating bank item: ID=%d, Quantity=%d, SlotID=%d, Tab=%s, Location=%s\n", 
		itemID, quantity, slotID, tabName, locationID)

	now := time.Now().Unix()
	var action string
	var oldQuantity int
	
	if item, exists := pi.Items[itemID]; exists {
		oldQuantity = item.Quantity
		action = "Updated"
		fmt.Printf("[BANK ITEM] Updating existing item: ID=%d, Old Quantity=%d, New Quantity=%d, Delta=%d\n", 
			itemID, oldQuantity, quantity, quantity-oldQuantity)
		item.Quantity = quantity
		item.SlotID = slotID
		item.LastSeen = now
	} else {
		oldQuantity = 0
		action = "Added"
		fmt.Printf("[BANK ITEM] Adding new item: ID=%d, Quantity=%d\n", itemID, quantity)
		pi.Items[itemID] = &InventoryItem{
			ItemID:   itemID,
			Quantity: quantity,
			SlotID:   slotID,
			LastSeen: now,
		}
	}

	pi.LastUpdated = now
	
	// Queue webhook update with bank information
	// Note: Bank items are always sent with override=true in the final webhook payload
	if pi.webhookURL != "" {
		fmt.Printf("[BANK ITEM] Queueing webhook event for item ID=%d, Action=%s, LocationID=%s\n", itemID, action, locationID)
		pi.queueWebhookEventWithBank(action, itemID, quantity, oldQuantity, slotID, true, locationID)
	} else {
		fmt.Printf("[BANK ITEM] Webhook URL is empty, not queueing event for item ID=%d\n", itemID)
	}
	
	// Only save to file if an output path is provided
	if pi.outputPath != "" {
		pi.SaveToFile()
		fmt.Printf("[BANK ITEM] Saved inventory to file: %s\n", pi.outputPath)
	} else {
		fmt.Printf("[BANK ITEM] Output path is empty, not saving to file\n")
	}
}

// SendBankItemsBatch sends a batch update with all bank items
func (pi *PlayerInventory) SendBankItemsBatch(tabName string, locationID string) {
	fmt.Printf("\n[WEBHOOK BANK BATCH] ===== ATTEMPTING TO SEND BANK ITEMS BATCH =====\n")
	log.Infof("[WEBHOOK BANK BATCH] Starting bank items batch send for tab: %s, location: %s", tabName, locationID)
	
	if pi.webhookURL == "" {
		fmt.Printf("[WEBHOOK ERROR] Cannot send bank items batch: webhook URL is empty\n")
		log.Error("[WEBHOOK ERROR] Cannot send bank items batch: webhook URL is empty")
		fmt.Printf("[WEBHOOK BANK BATCH] ===== END BANK ITEMS BATCH (FAILED) =====\n\n")
		return
	}
	
	if len(pi.pendingEvents) == 0 {
		fmt.Printf("[WEBHOOK ERROR] Cannot send bank items batch: no pending events\n")
		log.Error("[WEBHOOK ERROR] Cannot send bank items batch: no pending events")
		fmt.Printf("[WEBHOOK BANK BATCH] ===== END BANK ITEMS BATCH (FAILED) =====\n\n")
		return
	}

	fmt.Printf("[WEBHOOK BANK BATCH] Sending batch of %d bank items from tab: %s\n", len(pi.pendingEvents), tabName)
	log.Infof("[WEBHOOK BANK BATCH] Sending batch of %d bank items from tab: %s", len(pi.pendingEvents), tabName)

	// Check for equipment items in the batch
	hasEquipmentItems := false
	equippedItems := 0
	for _, event := range pi.pendingEvents {
		if event.SlotID >= 1000000 { // Equipment items typically have slot IDs over 1,000,000
			hasEquipmentItems = true
			if event.Equipped {
				equippedItems++
			}
		}
	}
	
	if hasEquipmentItems {
		fmt.Printf("[WEBHOOK BANK BATCH] Batch includes %d equipment items (%d equipped, slots > 1,000,000)\n", 
			countEquipmentItems(pi.pendingEvents), equippedItems)
		log.Infof("[WEBHOOK BANK BATCH] Batch includes %d equipment items (%d equipped)", 
			countEquipmentItems(pi.pendingEvents), equippedItems)
	} else {
		fmt.Printf("[WEBHOOK BANK BATCH] Batch contains only inventory items (no equipment)\n")
		log.Infof("[WEBHOOK BANK BATCH] Batch contains only inventory items (no equipment)")
	}

	// For bank items, we always set override=true
	// This ensures that bank tab content operations always override previous data
	override := true
	fmt.Printf("[WEBHOOK BANK BATCH] Override=true: Bank items always override previous data\n")
	log.Infof("[WEBHOOK BANK BATCH] Override=true: Bank items always override previous data")

	payload := WebhookPayload{
		Events:         pi.pendingEvents,
		CharacterID:    pi.CharacterID,
		CharacterName:  pi.CharacterName,
		BatchTimestamp:   time.Now().Unix(),
		Override:         override, // Always true for bank items
		Bank:             true,
	}

	if pi.isBank && pi.currentBankLocation != "" {
		fmt.Printf("[WEBHOOK BANK BATCH] Including location ID in events: %s\n", pi.currentBankLocation)
		log.Infof("[WEBHOOK BANK BATCH] Including location ID in events: %s", pi.currentBankLocation)
		
		// Update the location ID in each event
		for i := range pi.pendingEvents {
			if pi.pendingEvents[i].LocationID == "" {
				pi.pendingEvents[i].LocationID = pi.currentBankLocation
			}
		}
	}

	// Don't print the full payload to reduce log verbosity
	fmt.Printf("[WEBHOOK PAYLOAD] Bank items batch payload contains %d events (not shown to reduce verbosity)\n", len(pi.pendingEvents))
	log.Debugf("[WEBHOOK PAYLOAD] Bank items batch payload contains %d events", len(pi.pendingEvents))

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Failed to marshal batched webhook payload: %v", err)
		fmt.Printf("[WEBHOOK ERROR] Failed to marshal batched webhook payload: %v\n", err)
		fmt.Printf("[WEBHOOK BANK BATCH] ===== END BANK ITEMS BATCH (FAILED) =====\n\n")
		
		// Increment retry counter and check if we should keep retrying
		pi.webhookRetryCount++
		if pi.webhookRetryCount > 3 {
			log.Warnf("[WEBHOOK RETRY] Exceeded maximum retry count (%d), clearing %d pending events", 
				pi.webhookRetryCount, len(pi.pendingEvents))
			fmt.Printf("[WEBHOOK RETRY] Exceeded maximum retry count (%d), clearing %d pending events\n", 
				pi.webhookRetryCount, len(pi.pendingEvents))
			pi.pendingEvents = nil
			pi.webhookRetryCount = 0
		} else {
			// Log that we're keeping the events in the queue for retry
			log.Infof("[WEBHOOK RETRY] Keeping %d events in queue for retry after timeout error (attempt %d/3)", 
				len(pi.pendingEvents), pi.webhookRetryCount)
			fmt.Printf("[WEBHOOK RETRY] Keeping %d events in queue for retry after timeout error (attempt %d/3)\n", 
				len(pi.pendingEvents), pi.webhookRetryCount)
		}
		
		return
	}

	fmt.Printf("[WEBHOOK REQUEST] Sending POST request to: %s\n", pi.webhookURL)
	log.Infof("[WEBHOOK REQUEST] Sending POST request to: %s with %d bytes", pi.webhookURL, len(jsonPayload))
	
	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Create the request
	req, err := http.NewRequest("POST", pi.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Errorf("Failed to create request for bank items batch: %v", err)
		fmt.Printf("[WEBHOOK ERROR] Failed to create request for bank items batch: %v\n", err)
		fmt.Printf("[WEBHOOK BANK BATCH] ===== END BANK ITEMS BATCH (FAILED) =====\n\n")
		return
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlbionData-Client")
	
	// Send the request
	startTime := time.Now()
	resp, err := client.Do(req)
	requestDuration := time.Since(startTime)
	
	if err != nil {
		log.Errorf("Failed to send batched webhook: %v", err)
		fmt.Printf("[WEBHOOK ERROR] Failed to send batched webhook: %v\n", err)
		fmt.Printf("[WEBHOOK BANK BATCH] ===== END BANK ITEMS BATCH (FAILED) =====\n\n")
		
		// Increment retry counter and check if we should keep retrying
		pi.webhookRetryCount++
		if pi.webhookRetryCount > 3 {
			log.Warnf("[WEBHOOK RETRY] Exceeded maximum retry count (%d), clearing %d pending events", 
				pi.webhookRetryCount, len(pi.pendingEvents))
			fmt.Printf("[WEBHOOK RETRY] Exceeded maximum retry count (%d), clearing %d pending events\n", 
				pi.webhookRetryCount, len(pi.pendingEvents))
			pi.pendingEvents = nil
			pi.webhookRetryCount = 0
		} else {
			// Log that we're keeping the events in the queue for retry
			log.Infof("[WEBHOOK RETRY] Keeping %d events in queue for retry after timeout error (attempt %d/3)", 
				len(pi.pendingEvents), pi.webhookRetryCount)
			fmt.Printf("[WEBHOOK RETRY] Keeping %d events in queue for retry after timeout error (attempt %d/3)\n", 
				len(pi.pendingEvents), pi.webhookRetryCount)
		}
		
		return
	}
	defer resp.Body.Close()

	// Read and print the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
		fmt.Printf("[WEBHOOK ERROR] Failed to read response body: %v\n", err)
	}
	
	// Log request timing
	fmt.Printf("[WEBHOOK TIMING] Request completed in %v\n", requestDuration)
	log.Infof("[WEBHOOK TIMING] Request completed in %v", requestDuration)
	
	// Check response status
	if resp.StatusCode >= 400 {
		log.Errorf("Batched webhook returned error status: %d", resp.StatusCode)
		fmt.Printf("[WEBHOOK ERROR] Batched webhook returned error status: %d\n", resp.StatusCode)
		fmt.Printf("[WEBHOOK RESPONSE] Error response body: %s\n", string(respBody))
		log.Errorf("[WEBHOOK RESPONSE] Error response body: %s", string(respBody))
		fmt.Printf("[WEBHOOK BANK BATCH] ===== END BANK ITEMS BATCH (FAILED) =====\n\n")
	} else {
		fmt.Printf("[WEBHOOK SUCCESS] Successfully sent batch of %d bank items\n", len(pi.pendingEvents))
		fmt.Printf("[WEBHOOK RESPONSE] Response status: %d\n", resp.StatusCode)
		// Don't print the full response body to reduce log verbosity
		fmt.Printf("[WEBHOOK RESPONSE] Response received (not shown to reduce verbosity)\n")
		log.Infof("[WEBHOOK SUCCESS] Successfully sent batch of %d bank items", len(pi.pendingEvents))
		log.Debugf("[WEBHOOK RESPONSE] Response status: %d, body: %s", resp.StatusCode, string(respBody))
		fmt.Printf("[WEBHOOK BANK BATCH] ===== END BANK ITEMS BATCH (SUCCESS) =====\n\n")
	}

	// Clear pending events after send
	pi.pendingEvents = nil
	log.Infof("[WEBHOOK BANK BATCH] Cleared pending events queue after sending batch")

	// Reset retry counter on success
	pi.webhookRetryCount = 0
}

// SetBankBatchTimer sets a timer to send the bank batch after the specified duration
func (pi *PlayerInventory) SetBankBatchTimer(duration time.Duration, tabName string, locationID string) {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	// Stop any existing timer
	if pi.batchTimer != nil {
		pi.batchTimer.Stop()
	}
	
	// Create a new timer that will send the bank batch after the specified duration
	pi.batchTimer = time.AfterFunc(duration, func() {
		fmt.Printf("\n[BANK BATCH TIMER] Timer expired after %v, sending bank items batch\n", duration)
		log.Infof("[BANK BATCH TIMER] Timer expired after %v, sending bank items batch", duration)
		pi.SendBankItemsBatch(tabName, locationID)
	})
	
	fmt.Printf("[BANK BATCH TIMER] Set timer to send bank items batch after %v\n", duration)
	log.Infof("[BANK BATCH TIMER] Set timer to send bank items batch after %v", duration)
} 
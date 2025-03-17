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
	SlotID     int    `json:"slot_id"`
	Delta      int    `json:"delta"`
	Action     string `json:"action"`
	Equipped   bool   `json:"equipped,omitempty"`
	Timestamp  int64  `json:"timestamp"`
	LocationID string `json:"location_id,omitempty"`
	IsBank     bool   `json:"-"` // Not sent to webhook, used internally
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
	// Timer for bank batch events
	bankBatchTimer *time.Timer
}

// NewPlayerInventory creates a new player inventory tracker
func NewPlayerInventory(outputPath string, webhookURL string) *PlayerInventory {
	// Set initial time to be well in the past (30 minutes ago)
	// This ensures we don't start with short timers thinking we're close to a cluster change
	initialTime := time.Now().Add(-30 * time.Minute)
	
	pi := &PlayerInventory{
		Items:          make(map[int]*InventoryItem),
		outputPath:     outputPath,
		verboseOutput:  true, // Enable verbose output by default when inventory tracking is enabled
		webhookURL:     webhookURL,
		pendingEvents:  make([]WebhookEvent, 0),
		webhookRetryCount: 0, // Initialize retry counter
		lastClusterChange: initialTime,
		lastJoinTime:      initialTime,
		lastLeaveTime:     initialTime,
		lastBankAccess:    initialTime,
		lastEventTime:     initialTime,
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
	log.Infof("[INVENTORY] Queue opened: Player inventory sync queue (cluster change trigger)")
}

// OnJoin is called when a join operation is detected
func (pi *PlayerInventory) OnJoin() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.lastJoinTime = time.Now()
	
	// Use appropriate prefix based on cluster change timing
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	prefix := "[INVENTORY]"
	if timeSinceClusterChange <= 10*time.Second {
		prefix = "[INVENTORY-OVERRIDE]"
	}
	
	log.Debugf("%s Join operation detected at %v", prefix, pi.lastJoinTime)
}

// OnLeave is called when a leave event is detected
func (pi *PlayerInventory) OnLeave() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.lastLeaveTime = time.Now()
	
	// Use appropriate prefix based on cluster change timing
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	prefix := "[INVENTORY]"
	if timeSinceClusterChange <= 10*time.Second {
		prefix = "[INVENTORY-OVERRIDE]"
	}
	
	log.Debugf("%s Leave event detected at %v", prefix, pi.lastLeaveTime)
}

// UpdateCharacterInfo updates the character information
func (pi *PlayerInventory) UpdateCharacterInfo(characterID, characterName string) {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	pi.CharacterID = characterID
	pi.CharacterName = characterName
	
	// Use appropriate prefix based on cluster change timing
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	prefix := "[INVENTORY]"
	if timeSinceClusterChange <= 10*time.Second {
		prefix = "[INVENTORY-OVERRIDE]"
	}
	
	log.Infof("%s Character info updated: %s (%s)", prefix, characterName, characterID)
}

// queueWebhookEvent adds an event to the pending events queue and resets the timer
func (pi *PlayerInventory) queueWebhookEvent(action string, itemID int, quantity int, oldQuantity int, slotID int) {
	// Call the new method with default values for bank information
	pi.queueWebhookEventWithBank(action, itemID, quantity, oldQuantity, slotID, false, "")
}

// sendBatchedEvents sends all pending events in a single webhook request
func (pi *PlayerInventory) sendBatchedEvents() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	// Determine if we're in override mode (within 10s of cluster change)
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	isOverride := timeSinceClusterChange <= 10*time.Second
	prefix := "[INVENTORY]"
	if isOverride {
		prefix = "[INVENTORY-OVERRIDE]"
	}
	
	log.Infof("%s Attempting to send batched events (%d events)", prefix, len(pi.pendingEvents))
	
	if pi.webhookURL == "" {
		log.Error("%s Failed to send batch: webhook URL is empty", prefix)
		return
	}
	
	if len(pi.pendingEvents) == 0 {
		log.Error("%s Failed to send batch: no pending events", prefix)
		return
	}

	// Check if we're within 4 seconds of a cluster change
	if timeSinceClusterChange <= 4*time.Second {
		log.Infof("%s Skipping batch send - too close to cluster change (%v)", prefix, timeSinceClusterChange)
		
		// Reset the timer to try again in 2 seconds
		pi.batchTimer.Reset(2 * time.Second)
		
		return
	}

	// Filter out bank events - they should only be sent via SendBankItemsBatch
	nonBankEvents := make([]WebhookEvent, 0)
	for _, event := range pi.pendingEvents {
		if !event.IsBank {
			nonBankEvents = append(nonBankEvents, event)
		}
	}
	
	// If there are no non-bank events, don't send anything
	if len(nonBankEvents) == 0 {
		log.Infof("%s No non-bank events to send in default queue", prefix)
		return
	}

	log.Infof("%s Sending batch of %d events for character %s (%s)", 
		prefix, len(nonBankEvents), pi.CharacterName, pi.CharacterID)
	
	// Determine if we should set override flag
	override := isOverride
	
	log.Infof("%s Batch details: Override=%v, Bank=false, LocationID=0", prefix, override)

	payload := WebhookPayload{
		Events:         nonBankEvents,
		CharacterID:    pi.CharacterID,
		CharacterName:  pi.CharacterName,
		BatchTimestamp: time.Now().Unix(),
		Override:       override,
		Bank:           false,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("%s Failed to marshal webhook payload: %v", prefix, err)
		return
	}
	
	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Create a new request
	req, err := http.NewRequest("POST", pi.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Errorf("%s Failed to create webhook request: %v", prefix, err)
		return
	}
	
	// Set the content type header
	req.Header.Set("Content-Type", "application/json")
	
	// Add a timestamp header for debugging
	req.Header.Set("X-Albion-Inventory-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	
	// Add a batch ID header for tracking
	batchID := fmt.Sprintf("%d-%d", time.Now().Unix(), len(nonBankEvents))
	req.Header.Set("X-Albion-Inventory-Batch-ID", batchID)
	
	// Record the start time for timing
	startTime := time.Now()
	
	// Send the request
	log.Infof("%s Sending webhook request to %s", prefix, pi.webhookURL)
	
	resp, err := client.Do(req)
	requestDuration := time.Since(startTime)
	
	if err != nil {
		log.Errorf("%s Failed to send webhook: %v", prefix, err)
		
		// Increment retry counter and check if we should keep retrying
		pi.webhookRetryCount++
		if pi.webhookRetryCount > 3 {
			log.Warnf("%s Exceeded maximum retry count (%d), clearing %d pending events", 
				prefix, pi.webhookRetryCount, len(nonBankEvents))
			// Only remove non-bank events
			pi.removeNonBankEvents()
			pi.webhookRetryCount = 0
		} else {
			// Log that we're keeping the events in the queue for retry
			log.Infof("%s Keeping %d events in queue for retry after error (attempt %d/3)", 
				prefix, len(nonBankEvents), pi.webhookRetryCount)
		}
		
		return
	}
	defer resp.Body.Close()

	// Read the response body
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("%s Failed to read response body: %v", prefix, err)
	}
	
	// Log request timing
	log.Infof("%s Request completed in %v", prefix, requestDuration)
	
	// Check response status
	if resp.StatusCode >= 400 {
		log.Errorf("%s Webhook returned error status: %d", prefix, resp.StatusCode)
		
		// Increment retry counter and check if we should keep retrying
		pi.webhookRetryCount++
		if pi.webhookRetryCount > 3 {
			log.Warnf("%s Exceeded maximum retry count (%d), clearing %d pending events", 
				prefix, pi.webhookRetryCount, len(nonBankEvents))
			// Only remove non-bank events
			pi.removeNonBankEvents()
			pi.webhookRetryCount = 0
		} else {
			// Log that we're keeping the events in the queue for retry
			log.Infof("%s Keeping %d events in queue for retry after HTTP error (attempt %d/3)", 
				prefix, len(nonBankEvents), pi.webhookRetryCount)
		}
	} else {
		log.Infof("%s Webhook request successful: status %d", prefix, resp.StatusCode)
		
		// Remove non-bank events after successful send
		pi.removeNonBankEvents()
		
		// Reset retry counter on success
		pi.webhookRetryCount = 0
		
		log.Infof("%s Queue cleared: Default inventory queue (batch sent successfully)", prefix)
	}
}

// removeNonBankEvents removes all non-bank events from the pending events queue
func (pi *PlayerInventory) removeNonBankEvents() {
	bankEvents := make([]WebhookEvent, 0)
	for _, event := range pi.pendingEvents {
		if event.IsBank {
			bankEvents = append(bankEvents, event)
		}
	}
	pi.pendingEvents = bankEvents
}

// queueWebhookEventWithBank adds an event to the pending events queue with bank information
func (pi *PlayerInventory) queueWebhookEventWithBank(action string, itemID int, quantity int, oldQuantity int, slotID int, isBank bool, locationID string) {
	if pi.webhookURL == "" {
		log.Error("[INVENTORY] Cannot queue webhook event: webhook URL is empty")
		return
	}

	// Calculate the delta (change in quantity)
	delta := quantity - oldQuantity
	
	// If this is not a tab content event (locationID is empty or not a number 1-5), use "0"
	if locationID == "" || !isValidTabContentLocationID(locationID) {
		locationID = "0"
	}
	
	// Log the event being queued with appropriate prefix
	if isBank {
		log.Infof("[BANK-OVERRIDE] Adding to queue: Item=%d, Action=%s, Quantity=%d, Delta=%d, LocationID=%s", 
			itemID, action, quantity, delta, locationID)
	} else {
		// Check if we're within 10 seconds of cluster change for prefix
		timeSinceClusterChange := time.Since(pi.lastClusterChange)
		if timeSinceClusterChange <= 10*time.Second {
			log.Infof("[INVENTORY-OVERRIDE] Adding to queue: Item=%d, Action=%s, Quantity=%d, Delta=%d", 
				itemID, action, quantity, delta)
		} else {
			log.Infof("[INVENTORY] Adding to queue: Item=%d, Action=%s, Quantity=%d, Delta=%d", 
				itemID, action, quantity, delta)
		}
	}

	event := WebhookEvent{
		ItemID:     itemID,
		Quantity:   quantity,
		SlotID:     slotID,
		Delta:      delta,
		Action:     action,
		Equipped:   false,
		Timestamp:  time.Now().Unix(),
		LocationID: locationID,
		IsBank:     isBank, // Store whether this is a bank event
	}

	// Add the event to the queue
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.pendingEvents = append(pi.pendingEvents, event)
	pi.lastEventTime = time.Now() // Track when the last event was added
	
	// Only set the regular batch timer for non-bank events
	// Bank events should only be sent via the explicit SendBankItemsBatch call
	if !isBank {
		// If we're within 10 seconds of a cluster change, use a shorter batch timer (2 seconds)
		// Otherwise, use the standard 10 second timer
		var waitTime time.Duration
		timeSinceClusterChange := time.Since(pi.lastClusterChange)
		if timeSinceClusterChange <= 10*time.Second {
			waitTime = 2 * time.Second
			log.Infof("[INVENTORY-OVERRIDE] Setting batch timer: 2s (within 10s of cluster change)")
		} else {
			waitTime = 10 * time.Second
			log.Infof("[INVENTORY] Setting batch timer: 10s (normal operation)")
		}
		
		// Reset the timer with the appropriate wait time
		oldTimerActive := pi.batchTimer.Stop()
		pi.batchTimer.Reset(waitTime)
		log.Debugf("[INVENTORY] Timer reset: waitTime=%v, oldTimerActive=%v", waitTime, oldTimerActive)
	}
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
	
	// If this is not a tab content event (locationID is empty or not a number 1-5), use "0"
	if locationID == "" || !isValidTabContentLocationID(locationID) {
		locationID = "0"
	}
	
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
		log.Errorf("[INVENTORY] Failed to marshal webhook payload: %v", err)
		return
	}
	
	// Send the payload to the webhook URL
	resp, err := http.Post(pi.webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Errorf("[INVENTORY] Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		log.Errorf("[INVENTORY] Webhook returned error status: %d", resp.StatusCode)
	} else {
		log.Infof("[INVENTORY] Single item update sent successfully: Item=%d, Status=%d", itemID, resp.StatusCode)
	}
}

// NotifyBankVaultAccess updates the bank vault access time and location
// This should ONLY be called from operationAssetOverviewTabContent.Process
// NOTE: This method only updates state and does NOT trigger bank queues.
// Bank queues are ONLY triggered by operationAssetOverviewTabContent.Process
func (pi *PlayerInventory) NotifyBankVaultAccess(locationID string) {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	// Validate that this is a bank operation with a valid location ID
	if !isValidTabContentLocationID(locationID) {
		log.Warnf("[BANK-OVERRIDE] Invalid bank location ID: %s", locationID)
		return
	}
	
	pi.lastBankAccess = time.Now()
	pi.isBank = true
	pi.currentBankLocation = locationID
	
	log.Infof("[BANK-OVERRIDE] Bank access detected: LocationID=%s", locationID)
}

// PrintDebugState prints the current state of the inventory tracker
func (pi *PlayerInventory) PrintDebugState() {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()
	
	// Use appropriate prefix based on cluster change timing
	timeSinceClusterChange := time.Since(pi.lastClusterChange)
	prefix := "[INVENTORY]"
	if timeSinceClusterChange <= 10*time.Second {
		prefix = "[INVENTORY-OVERRIDE]"
	}
	
	log.Infof("%s State: Character=%s, Items=%d, PendingEvents=%d", 
		prefix, pi.CharacterName, len(pi.Items), len(pi.pendingEvents))
	log.Debugf("%s Last cluster change: %v (%v ago)", 
		prefix, pi.lastClusterChange, time.Since(pi.lastClusterChange))
	log.Debugf("%s Bank state: IsBank=%v, LocationID=%s, LastAccess=%v (%v ago)", 
		prefix, pi.isBank, pi.currentBankLocation, pi.lastBankAccess, time.Since(pi.lastBankAccess))
}

// NotifyClusterChange updates the cluster change timestamp
func (pi *PlayerInventory) NotifyClusterChange() {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()
	
	pi.lastClusterChange = time.Now()
	log.Infof("[INVENTORY-OVERRIDE] Queue opened: Player inventory sync queue (cluster change trigger)")
	
	// Print debug state after cluster change
	pi.PrintDebugState()
}

// queueWebhookEventWithEquipped adds an event to the pending events queue with equipped information
func (pi *PlayerInventory) queueWebhookEventWithEquipped(action string, itemID int, quantity int, oldQuantity int, slotID int, equipped bool, isBank bool, locationID string) {
	if pi.webhookURL == "" {
		return
	}

	// Calculate the delta (change in quantity)
	delta := quantity - oldQuantity

	// If this is not a tab content event (locationID is empty or not a number 1-5), use "0"
	if locationID == "" || !isValidTabContentLocationID(locationID) {
		locationID = "0"
	}

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
	isOverride := timeSinceClusterChange <= 10*time.Second
	prefix := "[INVENTORY]"
	if isOverride {
		prefix = "[INVENTORY-OVERRIDE]"
	}
	
	if timeSinceClusterChange <= 10*time.Second {
		waitTime = 2 * time.Second
		log.Infof("%s Queue update: %d events (2s timer - after cluster change)", prefix, len(pi.pendingEvents))
	} else {
		waitTime = 10 * time.Second
		log.Infof("%s Queue update: %d events (10s timer - normal operation)", prefix, len(pi.pendingEvents))
	}
	
	// Reset the timer with the appropriate wait time
	pi.batchTimer.Reset(waitTime)
}

// AddOrUpdateBankItem adds or updates a bank item in the inventory
func (pi *PlayerInventory) AddOrUpdateBankItem(itemID int, quantity int, slotID int, tabName string, locationID string) {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	log.Infof("[BANK-OVERRIDE] Processing bank item: ID=%d, Quantity=%d, Tab=%s, LocationID=%s", 
		itemID, quantity, tabName, locationID)

	now := time.Now().Unix()
	var action string
	var oldQuantity int
	
	if item, exists := pi.Items[itemID]; exists {
		oldQuantity = item.Quantity
		action = "Updated"
		log.Debugf("[BANK-OVERRIDE] Updating bank item: ID=%d, Old=%d, New=%d, Delta=%d", 
			itemID, oldQuantity, quantity, quantity-oldQuantity)
		item.Quantity = quantity
		item.SlotID = slotID
		item.LastSeen = now
	} else {
		oldQuantity = 0
		action = "Added"
		log.Debugf("[BANK-OVERRIDE] Adding new bank item: ID=%d, Quantity=%d", itemID, quantity)
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
		pi.queueWebhookEventWithBank(action, itemID, quantity, oldQuantity, slotID, true, locationID)
	}
	
	// Only save to file if an output path is provided
	if pi.outputPath != "" {
		pi.SaveToFile()
	}
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
			log.Errorf("[INVENTORY] Failed to create directory for inventory file: %v", err)
			return
		}
	}

	data, err := json.MarshalIndent(pi, "", "  ")
	if err != nil {
		log.Errorf("[INVENTORY] Failed to marshal inventory data: %v", err)
		return
	}

	err = os.WriteFile(pi.outputPath, data, 0644)
	if err != nil {
		log.Errorf("[INVENTORY] Failed to write inventory file: %v", err)
		return
	}

	log.Debugf("[INVENTORY] Saved inventory to %s", pi.outputPath)
}

// SendBankItemsBatch sends all pending bank items as a batch
func (pi *PlayerInventory) SendBankItemsBatch(tabName string) {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()

	// Check if there are any pending bank events
	bankEvents := make([]WebhookEvent, 0)
	for _, event := range pi.pendingEvents {
		if event.IsBank {
			bankEvents = append(bankEvents, event)
		}
	}
	
	if len(bankEvents) == 0 {
		log.Infof("[BANK-OVERRIDE] No pending events to send")
		// Reset bank state even if there are no events
		pi.isBank = false
		pi.bankBatchTimer = nil
		return
	}

	log.Infof("[BANK-OVERRIDE] Attempting to send batch: Tab=%s, LocationID=%s, Events=%d", 
		tabName, pi.currentBankLocation, len(bankEvents))
	
	// Determine if we should set override flag (always true for bank batches)
	override := true
	
	log.Infof("[BANK-OVERRIDE] Batch details: Override=%v, Bank=true, LocationID=%s", 
		override, pi.currentBankLocation)
	
	log.Infof("[BANK-OVERRIDE] Sending batch with events from various location IDs")

	payload := WebhookPayload{
		Events:         bankEvents,
		CharacterID:    pi.CharacterID,
		CharacterName:  pi.CharacterName,
		BatchTimestamp: time.Now().Unix(),
		Override:       override,
		Bank:           true,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("[BANK-OVERRIDE] Failed to marshal webhook payload: %v", err)
		return
	}
	
	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Create a new request
	req, err := http.NewRequest("POST", pi.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Errorf("[BANK-OVERRIDE] Failed to create webhook request: %v", err)
		return
	}
	
	// Set the content type header
	req.Header.Set("Content-Type", "application/json")
	
	// Add a timestamp header for debugging
	req.Header.Set("X-Albion-Inventory-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	
	// Add a batch ID header for tracking
	batchID := fmt.Sprintf("bank-%d-%d", time.Now().Unix(), len(bankEvents))
	req.Header.Set("X-Albion-Inventory-Batch-ID", batchID)
	
	// Record the start time for timing
	startTime := time.Now()
	
	// Send the request
	log.Infof("[BANK-OVERRIDE] Sending webhook request to %s", pi.webhookURL)
	
	resp, err := client.Do(req)
	requestDuration := time.Since(startTime)
	
	if err != nil {
		log.Errorf("[BANK-OVERRIDE] Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[BANK-OVERRIDE] Failed to read response body: %v", err)
	}
	
	// Log request timing
	log.Infof("[BANK-OVERRIDE] Request completed in %v", requestDuration)
	
	// Check response status
	if resp.StatusCode >= 400 {
		log.Errorf("[BANK-OVERRIDE] Webhook returned error status: %d", resp.StatusCode)
	} else {
		log.Infof("[BANK-OVERRIDE] Webhook request successful: status %d", resp.StatusCode)
		
		// Remove bank events after successful send
		pi.removeBankEvents()
		
		log.Infof("[BANK-OVERRIDE] Queue cleared: Bank inventory queue (batch sent successfully)")
	}
	
	// Reset bank state
	pi.isBank = false
	pi.bankBatchTimer = nil
	
	log.Infof("[BANK-OVERRIDE] Bank state reset after sending batch")
	
	// Check if we need to reset the location ID
	if pi.currentBankLocation == "5" {
		log.Infof("[BANK-OVERRIDE] Sent batch with locationId=5, should reset location ID")
	}
}

// removeBankEvents removes all bank events from the pending events queue
func (pi *PlayerInventory) removeBankEvents() {
	nonBankEvents := make([]WebhookEvent, 0)
	for _, event := range pi.pendingEvents {
		if !event.IsBank {
			nonBankEvents = append(nonBankEvents, event)
		}
	}
	pi.pendingEvents = nonBankEvents
}

// SetBankBatchTimer sets a timer to send the bank batch after the specified duration
func (pi *PlayerInventory) SetBankBatchTimer(duration time.Duration, tabName string) {
	pi.eventMutex.Lock()
	defer pi.eventMutex.Unlock()

	// Cancel any existing timer
	if pi.bankBatchTimer != nil {
		pi.bankBatchTimer.Stop()
	}

	log.Infof("[BANK-OVERRIDE] Bank batch timer set: %v for location ID %s", duration, pi.currentBankLocation)

	// Set a new timer
	pi.bankBatchTimer = time.AfterFunc(duration, func() {
		log.Infof("[BANK-OVERRIDE] Bank batch timer expired after %v, sending bank items batch", duration)
		pi.SendBankItemsBatch(tabName)
	})
}

// isValidTabContentLocationID checks if a locationID is a valid tab content ID (1-5)
func isValidTabContentLocationID(locationID string) bool {
	return locationID == "1" || locationID == "2" || locationID == "3" || locationID == "4" || locationID == "5"
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
			log.Debugf("[INVENTORY] Converting negative quantity %d to %d for item %d", 
				originalQuantity, quantity, itemID)
		} else {
			// For very negative values, we're not sure how to interpret them
			// Log a warning but keep the original value
			log.Warnf("[INVENTORY] Received very negative quantity %d for item %d", 
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
		pi.queueWebhookEventWithBank(action, itemID, quantity, oldQuantity, slotID, isBank, locationID)
	}
	
	// Only save to file if an output path is provided
	if pi.outputPath != "" {
		pi.SaveToFile()
	}
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
			log.Debugf("[INVENTORY] Converting negative quantity %d to %d for item %d", 
				originalQuantity, quantity, itemID)
		} else {
			// For very negative values, we're not sure how to interpret them
			// Log a warning but keep the original value
			log.Warnf("[INVENTORY] Received very negative quantity %d for item %d", 
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
		pi.queueWebhookEventWithEquipped(action, itemID, quantity, oldQuantity, slotID, equipped, isBank, locationID)
	}
	
	// Only save to file if an output path is provided
	if pi.outputPath != "" {
		pi.SaveToFile()
	}
} 
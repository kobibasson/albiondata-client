# Albion Online Inventory Tracking

This feature allows you to track your character's inventory in Albion Online and send updates to a webhook endpoint.

## How to Use

### Webhook Integration

You can send inventory updates to a webhook endpoint by using the `-inventory-webhook` flag:

```
albiondata-client.exe -inventory-webhook "http://localhost:3000/api/inventory/webhook"
```

To secure your webhook endpoint, you can add an authorization header with the `-inventory-webhook-auth` flag:

```
albiondata-client.exe -inventory-webhook "http://localhost:3000/api/inventory/webhook" -inventory-webhook-auth "Bearer your-token-here"
```

### Debug Information

You can combine these flags with debug flags to see more information:

```
albiondata-client.exe -debug -events 32,26,27 -inventory-webhook "http://localhost:3000/api/inventory/webhook"
```

This will show debug information for the following events:
- `evNewSimpleItem` (32): When you gather or receive new items
- `evInventoryPutItem` (26): When items are added to your inventory
- `evInventoryDeleteItem` (27): When items are removed from your inventory

## Logging System

The inventory tracking feature uses a clear, consistent logging system that explains what's happening with the queues and webhooks. All logs are prefixed with `[INVENTORY]` for easy filtering. The logging system provides information about:

- Queue opening and closing events
- Why queues are opened or closed
- What information is being sent
- HTTP response status codes
- Retry attempts for failed webhook requests

Example log messages:

```
[INVENTORY] Queue opened: Player inventory sync queue (cluster change trigger)
[INVENTORY] Adding to default queue: Item=123, Action=Added, Quantity=10, Delta=10
[INVENTORY] Setting batch timer: 10s (normal operation)
[INVENTORY] Sending webhook request to http://localhost:3000/api/inventory/webhook
[INVENTORY] Webhook request successful: status 200
[INVENTORY] Queue cleared: Default inventory queue (batch sent successfully)
```

## How It Works

The inventory tracking feature works by monitoring game events and operations, then sending webhook updates with inventory changes. There are three main tracking mechanisms, each with different behavior:

### 1. Bank Inventory Tracking

**Trigger Event**: `opAssetOverviewTabContent` operation

When the player opens a bank vault tab, this operation is detected and triggers the bank inventory tracking mode. The system:

- Opens a queue for collecting subsequent inventory events
- Assigns a location ID (1-5) to identify different bank tabs
  - Location IDs increment each time a new tab is viewed (1→2→3→4→5→1)
  - IDs reset to 1 after reaching 5 or after 10 seconds of inactivity
- Collects all `evNewSimpleItem` and `evNewEquipmentItem` events that occur while the queue is open
- Each event is tagged with the location ID that was active when the event was added to the queue
- Sends a webhook payload with:
  - `override: true` - indicating this data should replace previous bank inventory data
  - `bank: true` - indicating these are bank items
  - Each event retains its original `location_id` (1-5) from when it was added to the queue

The bank queue closes after 10 seconds of inactivity or immediately if it's been more than 10 seconds since the last tab content event. The queue will only reopen when another `opAssetOverviewTabContent` operation is detected.

### 2. Player Inventory Synchronization

**Trigger Event**: Cluster change event (`opChangeCluster`)

When the player changes clusters (zones), this triggers a full inventory synchronization:

- Opens a queue for collecting subsequent inventory events
- Keeps the queue open for 4 seconds
- Collects all `evNewSimpleItem` and `evNewEquipmentItem` events that occur while the queue is open
- Sends a webhook payload with:
  - `override: true` - indicating this data should replace previous player inventory data
  - `bank: false` - indicating these are player inventory items
  - `location_id: "0"` - indicating player inventory (not bank)

The queue closes after 4 seconds and will only reopen when another cluster change event is detected.

### 3. Default Inventory Updates

**Trigger**: Any inventory event without a preceding trigger event

For inventory changes that occur outside of bank access or cluster changes:

- Processes individual `evNewSimpleItem` and `evNewEquipmentItem` events
- Sends a webhook payload with:
  - `override: false` - indicating these are incremental updates
  - `bank: false` - indicating these are player inventory items
  - `location_id: "0"` - indicating player inventory (not bank)

## Webhook Payload Format

### Player Inventory Example:

```json
{
  "events": [
    {
      "item_id": 909,
      "quantity": 122,
      "slot_id": 0,
      "delta": 122,
      "action": "Added",
      "timestamp": 1709971612
    },
    {
      "item_id": 912,
      "quantity": 946,
      "slot_id": 1,
      "delta": 5,
      "action": "Updated",
      "timestamp": 1709971615
    }
  ],
  "character_id": "your-character-uuid",
  "character_name": "YourCharacterName",
  "batch_timestamp": 1709971618,
  "override": false,
  "bank": false
}
```

### Bank Inventory Example:

```json
{
  "events": [
    {
      "item_id": 909,
      "quantity": 122,
      "slot_id": 0,
      "delta": 122,
      "action": "Added",
      "timestamp": 1709971612,
      "location_id": "3"
    },
    {
      "item_id": 910,
      "quantity": 50,
      "slot_id": 1,
      "delta": 50,
      "action": "Added",
      "timestamp": 1709971615,
      "location_id": "4"
    }
  ],
  "character_id": "your-character-uuid",
  "character_name": "YourCharacterName",
  "batch_timestamp": 1709971618,
  "override": true,
  "bank": true
}
```

Note that in the bank inventory example, events can have different location IDs within the same batch, as each event retains the location ID from when it was added to the queue.

## Payload Fields Explained

- `events`: Array of inventory events
  - `item_id`: The ID of the item
  - `quantity`: The current quantity of the item
  - `slot_id`: The slot ID of the item in the inventory
  - `delta`: The change in quantity (positive for increase, negative for decrease)
  - `action`: The action that triggered the event (Added, Updated, Removed)
  - `timestamp`: The Unix timestamp when the event occurred
  - `location_id`: For bank items, the bank tab index (1-5); for player inventory, "0"
  - `equipped`: For equipment items, whether the item is equipped (boolean)

- `character_id`: The UUID of the character
- `character_name`: The name of the character
- `batch_timestamp`: The Unix timestamp when the batch was sent
- `override`: Whether this batch should override previous inventory data
  - `true` for bank operations and cluster changes
  - `false` for incremental updates
- `bank`: Whether this batch contains bank items
  - `true` for bank operations
  - `false` for player inventory

## Implementation Details

The system uses different queues and timers to manage the three tracking mechanisms:

1. **Bank Inventory**: Triggered by `opAssetOverviewTabContent`, uses location IDs 1-5, always sets `override: true` and `bank: true`

2. **Player Inventory Sync**: Triggered by cluster changes, uses a 4-second queue, sets `override: true` and `bank: false`

3. **Default Updates**: No trigger event, processes individual events, sets `override: false` and `bank: false`

All three mechanisms use the same underlying event handlers for `evNewSimpleItem` and `evNewEquipmentItem`, but with different metadata based on the trigger event.

## Troubleshooting

If you're not receiving inventory updates:

1. Make sure you're running the Albion Data Client with the `-inventory-webhook` flag
2. Verify that your webhook server is running and accessible
3. Look for webhook error messages in the Albion Data Client logs (prefixed with `[INVENTORY]`)
4. Make sure your webhook server is properly handling the POST requests
5. If you're using authentication, ensure the `-inventory-webhook-auth` flag contains the correct authorization value expected by your server
   - The client will display a masked version of your authorization header at startup and in logs before each request for verification
   - Look for messages like `Adding Authorization header: Beare...123` in the logs

If you encounter a "flag redefined" error when running the application, it means there's a conflict in the command-line flags. This has been fixed in the latest version, but if you're still experiencing this issue, please report it on the GitHub repository.

If you're still having issues, please report them on the GitHub repository.

## Technical Implementation Notes

The following technical details document recent fixes and implementation specifics that may be helpful for troubleshooting or future development:

### Authentication Header Support

All webhook requests can include an Authorization header for improved security. To use this feature:

1. Use the `-inventory-webhook-auth` flag when starting the client
2. The value provided will be added as an `Authorization` header to all webhook requests
3. Common formats include `Bearer your-token-here` or `Basic base64-encoded-credentials`
4. The authorization header is applied to all types of webhook requests (bank updates, inventory sync, and individual updates)

### Bank Location ID Handling

1. **Location ID Preservation**: Each bank event is tagged with the location ID (1-5) that was active when the event was added to the queue. Events retain their original location ID even when multiple bank tabs are accessed in sequence.

2. **Location ID Reset**: The location ID counter (`LastTabContentLocationID`) cycles from 1 to 5 and resets to 0 after reaching 5. This reset ensures the next tab content operation starts with location ID 1.

3. **Bank State Reset**: After sending a bank batch, the bank state (`isBank` flag and `bankBatchTimer`) is properly reset to prevent sending additional bank batches without a new trigger event.

### Timer Initialization

1. **Initial Time Values**: All time-related fields (`lastClusterChange`, `lastJoinTime`, `lastLeaveTime`, `lastBankAccess`, `lastEventTime`) are initialized to 30 minutes in the past rather than the current time. This prevents the system from incorrectly using short timers (2s) on startup, ensuring it begins with the normal 10-second timer for batching events.

2. **Batch Timer Retry**: If a batch send is skipped due to being too close to a cluster change (within 4 seconds), the timer is reset to try again in 2 seconds, ensuring events aren't lost.

### Queue Management

1. **Bank Queue Behavior**: Bank queues are only triggered by `opAssetOverviewTabContent` operations, not by bank vault access notifications. The `NotifyBankVaultAccess` method only updates state and does not trigger queues.

2. **Auto-Reset Logic**: If it's been more than 3 seconds since the last bank access and the bank state is still active, the system automatically resets the bank state before sending a batch.

3. **Timer Management**: After a bank batch is sent via timer, a new timer is set up for any pending default events to ensure they aren't lost.

### Deviations from Documentation

The implementation includes some technical details that differ slightly from the documentation:

1. **Bank State Reset**: The documentation states that "The bank queue closes after 10 seconds of inactivity," but the implementation also explicitly resets the bank state after sending a batch, regardless of inactivity time.

2. **Location ID Handling**: While the documentation mentions that location IDs cycle from 1-5, the implementation specifically resets to 0 (which becomes 1 on the next event) after reaching 5 or after sending a batch with locationId=5.

3. **Timer Behavior**: The implementation includes additional logic for retrying batch sends that are skipped due to being too close to cluster changes, which isn't explicitly mentioned in the documentation.

These implementation details ensure more reliable tracking of inventory changes, particularly when switching between bank tabs or during cluster changes. 
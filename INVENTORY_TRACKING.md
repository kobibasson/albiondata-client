# Albion Online Inventory Tracking

This feature allows you to track your character's inventory in Albion Online and send updates to a webhook endpoint.

## How to Use

### Webhook Integration

You can send inventory updates to a webhook endpoint by using the `-inventory-webhook` flag:

```
albiondata-client.exe -inventory-webhook "http://localhost:3000/api/inventory/webhook"
```

### Debug Information

You can combine this flag with debug flags to see more information:

```
albiondata-client.exe -debug -events 32,26,27 -inventory-webhook "http://localhost:3000/api/inventory/webhook"
```

This will show debug information for the following events:
- `evNewSimpleItem` (32): When you gather or receive new items
- `evInventoryPutItem` (26): When items are added to your inventory
- `evInventoryDeleteItem` (27): When items are removed from your inventory

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
- Sends a webhook payload with:
  - `override: true` - indicating this data should replace previous bank inventory data
  - `bank: true` - indicating these are bank items
  - `location_id` - the assigned bank tab index (1-5)

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
    }
  ],
  "character_id": "your-character-uuid",
  "character_name": "YourCharacterName",
  "batch_timestamp": 1709971618,
  "override": true,
  "bank": true
}
```

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
3. Look for webhook error messages in the Albion Data Client console
4. Make sure your webhook server is properly handling the POST requests

If you encounter a "flag redefined" error when running the application, it means there's a conflict in the command-line flags. This has been fixed in the latest version, but if you're still experiencing this issue, please report it on the GitHub repository.

If you're still having issues, please report them on the GitHub repository. 
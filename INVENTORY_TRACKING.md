# Albion Online Inventory Tracking

This feature allows you to track your character's inventory in Albion Online and save it to a local JSON file or send updates to a webhook.

## How to Use

### Local JSON File Tracking

Run the Albion Data Client with the `-inventory-output` flag to specify the path where you want to save the inventory JSON file:

```
albiondata-client.exe -inventory-output "C:\path\to\inventory.json"
```

### Webhook Integration

You can send inventory updates to a webhook endpoint by using the `-inventory-webhook` flag:

```
albiondata-client.exe -inventory-webhook "http://localhost:3000/api/inventory/webhook"
```

You can use the webhook without saving to a JSON file, or you can combine both features:

```
albiondata-client.exe -inventory-output "C:\path\to\inventory.json" -inventory-webhook "http://localhost:3000/api/inventory/webhook"
```

### Debug Information

You can also combine these flags with debug flags to see more information:

```
albiondata-client.exe -debug -events 32,26,27 -inventory-webhook "http://localhost:3000/api/inventory/webhook"
```

This will show debug information for the following events:
- `evNewSimpleItem` (32): When you gather or receive new items
- `evInventoryPutItem` (26): When items are added to your inventory
- `evInventoryDeleteItem` (27): When items are removed from your inventory

## Real-time Inventory Change Messages

When you run the application with the `-inventory-output` or `-inventory-webhook` flag, you'll see real-time messages about inventory changes in the console, even without the `-debug` flag. These messages include:

- When items are added to your inventory
- When item quantities are updated (with the change amount)
- When items are removed from your inventory
- When your character information is updated

Example messages:
```
[Inventory] Added item: ID=909, Quantity=122
[Inventory] Updated item: ID=909, Quantity=244 (+122)
[Inventory] Updated item: ID=912, Quantity=911 (+3)
[Inventory] Removed item: ID=909, Quantity was 244
```

## Webhook Debug Information

When using the `-inventory-webhook` flag, you'll see detailed debug information about the webhook requests and responses:

```
[Webhook] Sending: Added for item 909 (Qty: 122, Old: 0)
[Webhook] SUCCESS: Server responded with status code 200
[Webhook] Response: {"success":true}

[Webhook] Sending: Updated for item 912 (Qty: 946, Old: 941)
[Webhook] SUCCESS: Server responded with status code 200
[Webhook] Response: {"success":true}
```

If there are any errors, you'll see detailed error messages:

```
[Webhook] Sending: Updated for item 912 (Qty: 946, Old: 941)
[Webhook] ERROR: Failed to send webhook: dial tcp [::1]:3000: connectex: No connection could be made because the target machine actively refused it.
```

This makes it easy to debug webhook integration issues without having to enable the full debug mode.

## Webhook Integration

The webhook feature sends a JSON payload to the specified URL whenever your inventory changes. The payload includes:

```json
{
  "action": "Added",  // or "Updated", "Removed", "CharacterInfo"
  "item_id": 909,
  "quantity": 122,
  "old_quantity": 0,
  "character_id": "your-character-uuid",
  "character_name": "YourCharacterName",
  "timestamp": 1709971612
}
```

You can use this webhook to integrate with other applications, such as:
- A custom web dashboard to monitor your inventory
- A Discord bot to notify you of important inventory changes
- A database to track your gathering efficiency over time
- Any other application that can receive HTTP POST requests

## JSON File Format

The inventory JSON file will have the following format:

```json
{
  "character_id": "your-character-uuid",
  "character_name": "YourCharacterName",
  "items": {
    "909": {
      "item_id": 909,
      "quantity": 122,
      "last_seen": 1709971612
    },
    "912": {
      "item_id": 912,
      "quantity": 911,
      "last_seen": 1709971678
    }
  },
  "last_updated": 1709971678
}
```

- `character_id`: Your character's unique identifier
- `character_name`: Your character's name
- `items`: A map of items in your inventory, where the key is the item ID
  - `item_id`: The ID of the item
  - `quantity`: The quantity of the item
  - `last_seen`: The timestamp when the item was last seen (Unix timestamp)
- `last_updated`: The timestamp when the inventory was last updated (Unix timestamp)

## How It Works

The inventory tracking feature works by monitoring the following game events:

1. `evNewSimpleItem`: When you gather or receive new items
2. `evInventoryPutItem`: When items are added to your inventory
3. `evInventoryDeleteItem`: When items are removed from your inventory

When any of these events occur, the inventory tracker updates the JSON file (if enabled) and sends a webhook update (if enabled).

## Item IDs

The item IDs in the inventory data correspond to the item IDs in Albion Online. You can use these IDs to identify the items in your inventory.

For example, from the debug logs you provided:
- Item ID 909: Logs (T4 resource)
- Item ID 912: Rough Logs (T5 resource)

You can find more information about item IDs in the Albion Online Data Project or other community resources.

## Troubleshooting

If you're not seeing any updates to your inventory file:

1. Make sure you're running the Albion Data Client with the `-inventory-output` flag
2. Check that you have the correct path specified for the output file
3. Try running with the `-debug` flag to see more information
4. Make sure you're performing actions in the game that would trigger inventory updates (gathering, trading, etc.)

If you're using the webhook feature and not receiving updates:

1. Check that your webhook server is running and accessible
2. Verify that the URL is correct and includes the protocol (http:// or https://)
3. Look for webhook error messages in the Albion Data Client console
4. Make sure your webhook server is properly handling the POST requests

If you encounter a "flag redefined" error when running the application, it means there's a conflict in the command-line flags. This has been fixed in the latest version, but if you're still experiencing this issue, please report it on the GitHub repository.

If you're still having issues, please report them on the GitHub repository. 
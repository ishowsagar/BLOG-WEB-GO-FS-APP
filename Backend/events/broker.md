# Workflow with broker

> Now when there is "REQUEST" comes on the ws handler path -> executes function which upgrades the underlying connection into a webSocket connection.
> The client is created and stored in the hub as active, then reader and writer goroutines start.
> If the payload is a direct message, the reader publishes it with routing key `user.<id>`.
> If the payload is a room message, the reader publishes it with routing key `room.<id>`.
> The broker uses one topic exchange for both flows, and the consumer routes `user.*` deliveries to the targeted client path and `room.*` deliveries to the room path.

<!-- & Reason behind using RabbitMqBroker -->

> direct pushes needed more complex logic for p2p.
> The broker stores websocket events in one topic exchange and marks deliveries with routing keys.
> The publisher publishes the event to the exchange and stamps the message with `user.<id>` or `room.<id>`.
> The consumer receives deliveries for both routing patterns and redirects them to the hub.
> By doing this, events are marked first in the exchange and then routed to the correct client or room.

<!-- & Deductions -->

#1 - There should be one declared topic exchange where websocket deliveries are stored with routing keys
#2 - Need rabbit's conn and ch both for doing all these works, conn for networking and ch for all redirects and routing works
#3 - Need to declare publisher logic that stores the notification payload in the exchange with the correct routing key
#4 - Need to declare a consumer which consumes marked deliveries and redirects them to the hub
#5 - Routing keys are now `user.ID` for DM deliveries and `room.ID` for room deliveries
#6 - The consumer runs in the background and checks for deliveries matching both `user.*` and `room.*`

Conclusion -> invoke publish method to mark the delivery of the event and consumer to recieve and redirect ( publish to exchange where noti should/been be created ) and consume it

<!-- & Results -->

> Actually since reader is reading if there is payload coming and if it is -> stores in exchange for the delivery and consumer checks for those deliveries marked with routing key being the `userID` or `roomID` depending on the payload. If `reciever_id` is `0`, that is broadcast to all active clients. Room chat uses `room_id` and room membership logic instead of the DM receiver path.

> A ws request made on path -> with attached payload invokes broker and the rest of the operations.
> We publish to the topic exchange and the consumer is always running to check for deliveries related to `user.*` and `room.*` routing keys.
> That redirects into the hub's channels and then back to the correct active client or room.

> Dm payload

    ```json

{
"sender_id": 41,
"reciever_id": 16,
"sender_name": "edge Client",
"reciever_name": "brave Client",
"type": "dm",
"content": "hello from RabbitMQ UI",
"post_id": 0
}

> Room chat payload

```json
{
  "sender_id": 41,
  "reciever_id": 16,
  "sender_name": "Brave Client",
  "reciever_name": "Edge Client",
  "room_id": 1,
  "room_status": false,
  "type": "room_msg",
  "content": "hello in the group chat",
  "post_id": 0
}
```

> Use `room_status: true` when joining the room, and `room_status: false` with non-empty `content` when sending a room message.

case currentRoomPayloadRequest := <- h.RoomClientsPayloads :
slog.Info("Succesfully recieved payload in roomCLientsPayload chan", "room_id", currentRoomPayloadRequest.RoomID, "sender_id", currentRoomPayloadRequest.SenderID, "room_status", currentRoomPayloadRequest.RoomStatus)

    			roomClients, roomExists := h.ChatRoomClients[currentRoomPayloadRequest.RoomID]
    			if !roomExists && currentRoomPayloadRequest.RoomStatus {
    				roomClients = make(map[*Client]bool)
    				h.ChatRoomClients[currentRoomPayloadRequest.RoomID] = roomClients
    			}

    			var senderClient *Client // store msg sender in room
    			for activeClient := range h.Clients {
    				if activeClient.ID == currentRoomPayloadRequest.SenderID {
    					senderClient = activeClient
    					break
    				}
    			}

    			if senderClient == nil {
    				slog.Warn("HUB room payload sender not found among active clients", "room_id", currentRoomPayloadRequest.RoomID, "sender_id", currentRoomPayloadRequest.SenderID)
    				break
    			}

    			if currentRoomPayloadRequest.RoomStatus {
    				if !roomExists {
    					slog.Info("HUB created room map", "room_id", currentRoomPayloadRequest.RoomID)
    				}
    				roomClients[senderClient] = true //* stores sender Client in room
    				slog.Info("client is added to the room", "roomID", currentRoomPayloadRequest.RoomID, "clientID", currentRoomPayloadRequest.SenderID)

    				if currentRoomPayloadRequest.RecieverID != 0 && currentRoomPayloadRequest.RecieverID != currentRoomPayloadRequest.SenderID {
    					for activeClient := range h.Clients {
    						if activeClient.ID == currentRoomPayloadRequest.RecieverID {
    							roomClients[activeClient] = true
    							slog.Info("receiver added to the room", "roomID", currentRoomPayloadRequest.RoomID, "clientID", currentRoomPayloadRequest.RecieverID)
    							break
    						}
    					}
    				}
    				break
    			}

    			if currentRoomPayloadRequest.Content == "" {
    				delete(roomClients, senderClient)
    				slog.Info("client is disconnected from the room", "roomID", currentRoomPayloadRequest.RoomID, "clientID", currentRoomPayloadRequest.SenderID)
    				if len(roomClients) == 0 {
    					delete(h.ChatRoomClients, currentRoomPayloadRequest.RoomID)
    					slog.Info("HUB: room deleted (empty)", "room_id", currentRoomPayloadRequest.RoomID)
    				}
    				break
    			}

    			if !roomExists {
    				slog.Warn("HUB room message received but room does not exist locally", "room_id", currentRoomPayloadRequest.RoomID)
    				break
    			}

    			// * loop over each client in the roomClients -> rediect payload to eachClient.Send chan so it writes response to the those client in the writers
    			for currentRoomClient := range roomClients {
    				select {
    				case currentRoomClient.Send <- currentRoomPayloadRequest:
    					slog.Info("message sent in the room", "roomID", currentRoomPayloadRequest.RoomID, "message", currentRoomPayloadRequest.Content)
    				default:
    					slog.Warn("HUB failed to send room message, client.Send full", "client_id", currentRoomClient.ID, "room_id", currentRoomPayloadRequest.RoomID)
    				}
    			}

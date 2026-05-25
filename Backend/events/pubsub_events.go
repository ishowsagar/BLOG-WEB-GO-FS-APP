package events

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/rabbitmq/amqp091-go"
)

const notificationsExchangeName = "notifications_topic"

// PubSubBroker holds RabbitMQ connection and forwards messages to hub
type PubSubBroker struct {
	conn   *amqp091.Connection // stores that rabbitMQ connection
	ch     *amqp091.Channel // chan which holds all those events operations
	qName  string // instance queue name
	hub *services.Hub
}

// NewPubSubBroker returns a broker instance (connection not yet established)
func NewPubSubBroker(hub *services.Hub) *PubSubBroker {
	return &PubSubBroker{
		hub: hub,
	}
}

// Connect establishes persistent RabbitMQ connection and declares exchange/queue
func (p *PubSubBroker) Connect(rabbitURL string) error {
	
	// get connection
	conn, err := amqp091.Dial(rabbitURL)
	if err != nil {
		slog.Error("failed to connect to RabbitMQ", "error", err)
		return err
	}

	// retrieve concurrent channel from the conn 
	ch, err := conn.Channel()
	if err != nil {
		slog.Error("failed to create channel", "error", err)
		conn.Close()
		return err
	}

	//! declares exchange - where all events are stored and where all deliveries are stamped for delivery to consumers 
	err = ch.ExchangeDeclare(notificationsExchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		slog.Error("failed to declare exchange", "error", err)
		ch.Close()
		conn.Close()
		return err
	}

	// declare instance queue (auto-delete, exclusive to this worker)
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		slog.Error("failed to declare queue", "error", err)
		ch.Close()
		conn.Close()
		return err
	}

	//& all these config get assigned to the pubsubbroker instance -> stores all things now
	p.conn = conn
	p.ch = ch
	p.qName = q.Name
	slog.Info("RabbitMQ connected", "queue", p.qName)
	return nil
}

// Publish sends a message to a specific user (routing key = user.<id>)
func (p *PubSubBroker) PublishEvents(userID uint, payload *services.ClientNotifyPayload) error {
	
	// chan nil check first cause it needs to be connected to the rabbit
	if p.ch == nil {
		return fmt.Errorf("broker not connected")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal payload", "error", err)
		return err
	}

	routingKey := fmt.Sprintf("user.%d", userID)
	if payload.RoomID != 0 {
		routingKey = fmt.Sprintf("room.%d", payload.RoomID)
	}
	
	// publish notification for this user in -> format of user.ID - like key-val pair in "noti..." already declared event
	//& publishing an event in "notifications" exchange in such a way that -> we are stamping delivery of bodyPayload for this routingKey(userID)
	err = p.ch.Publish(notificationsExchangeName, routingKey, false, false, amqp091.Publishing{
		// now we have stored this marked delivery that this things belongs to this userID
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		slog.Error("failed to publish message", "error", err, "routing_key", routingKey)
		return err
	}

	slog.Debug("published message", "routing_key", routingKey, "user_id", userID)
	return nil
}

// BindUser binds this instance's queue to receive messages for a specific user
func (p *PubSubBroker) BindUserToTheExchange(userID uint) error {

	// & always first checks for connection, if conn is nil -> don't move furthur
	if p.ch == nil {
		return fmt.Errorf("broker not connected")
	}

	routingKey := fmt.Sprintf("user.%d", userID)
	
	// & binding user to the event
	err := p.ch.QueueBind(p.qName, routingKey, notificationsExchangeName, false, nil)
	if err != nil {
		slog.Error("failed to bind user queue", "error", err, "user_id", userID)
		return err
	}

	slog.Debug("bound user queue", "user_id", userID, "routing_key", routingKey)
	return nil
}

// UnbindUser unbinds this instance's queue from a specific user
func (p *PubSubBroker) UnbindUserFromExchange(userID uint) error {
	if p.ch == nil {
		return fmt.Errorf("broker not connected")
	}

	routingKey := fmt.Sprintf("user.%d", userID)
	err := p.ch.QueueUnbind(p.qName, routingKey, notificationsExchangeName, nil)
	if err != nil {
		slog.Error("failed to unbind user queue", "error", err, "user_id", userID)
		return err
	}

	slog.Debug("unbound user queue", "user_id", userID)
	return nil
}

// StartConsuming starts the message consumer loop (run as a goroutine) -> recieves all marked deliveries from the publishe
func (p *PubSubBroker) StartConsumingDeliveries() error {
	if p.ch == nil {
		return fmt.Errorf("broker not connected")
	}

	// bind the instance queue to both user and room routing keys

	// # binds the "room./user" to the queue, which always being checked by the ch.Consume -> when these queus are active it check for those
	err := p.ch.QueueBind(p.qName, "user.*", notificationsExchangeName, false, nil)
	if err != nil {
		slog.Error("failed to bind wildcard", "error", err)
		return err
	}

	// binding check for room related deliveries inthe notifications exchange
	err = p.ch.QueueBind(p.qName, "room.*", notificationsExchangeName, false, nil)
	if err != nil {
		slog.Error("failed to bind room wildcard", "error", err)
		return err
	}

	msgs, err := p.ch.Consume(p.qName, "", true, false, false, false, nil)
	if err != nil {
		slog.Error("failed to start consuming", "error", err)
		return err
	}

	slog.Info("started consuming from RabbitMQ", "queue", p.qName)

	go func() {
		for eachMsg := range msgs {
			var payload services.ClientNotifyPayload
			err := json.Unmarshal(eachMsg.Body, &payload)
			if err != nil {
				slog.Error("failed to unmarshal message", "error", err)
				continue
			}

			// consumer forward to hub for delivery to websocket clients
			slog.Info("CONSUMER received message", "sender", payload.SenderID, "receiver", payload.RecieverID, "type", payload.Type, "content", payload.Content)


			
			//& if incoming delivery payload has roomID <- recieving delivieries related to room
			if payload.RoomID != 0 {
				slog.Info("CONSUMER routing to RoomClientsPayloads", "room_id", payload.RoomID, "sender_id", payload.SenderID)
				select {
				case p.hub.RoomClientsPayloads <- &payload:
					slog.Info("CONSUMER successfully sent to RoomClientsPayloads", "room_id", payload.RoomID)
				default:
					slog.Error("CONSUMER unable to send to RoomClientsPayloads", "error", "channel full or blocked", "room_id", payload.RoomID)
				}
			} else if payload.RecieverID == 0 && payload.RoomID == 0  {
				//& each recieved delivery from publisher get decoded here and set to hub for delivery to the client
				slog.Info("CONSUMER routing to Broadcast (RecieverID=0)")
				select {
				case p.hub.Broadcast <- &payload:
				default:
					slog.Error("CONSUMER unable to send to Broadcast", "error", "channel full or blocked")
				}
			} else if payload.RoomID == 0 && payload.RecieverID != 0 {
				slog.Info("CONSUMER routing to TargettedBrokerMessages", "target_user", payload.RecieverID)
				select {
				case p.hub.TargettedBrokerMessages <- &payload:
					slog.Info("CONSUMER successfully sent to TargettedBrokerMessages")
				default:
					slog.Error("CONSUMER unable to send to TargettedBrokerMessages", "error", "channel full or blocked", "receiver", payload.RecieverID)
				}
				}else if payload.Type == "post_created" && payload.RecieverID != 0 && payload.RoomID == 0{
					// means incoming delivery is related to post_created and target reciever id is also there
					//* send to a dedicated hub chan which only redirects post type notification to the targetted user only
					p.hub.TargettedClientNotificationTypeOnly <- &payload // sending to broadcast's targettedNotfication where we will filter it out by checking payload type
					slog.Info("notification payload is successfully redirected to hub's TargettedClientNotificationTypeOnly chan")
					
				}else if payload.Type == "comment_posted" && payload.RecieverID !=0 && payload.RoomID == 0 {
				// * if these conditions are satisfied we have recieved intended commennt notification payload on the exchange via publisher
				select {
				case p.hub.TargettedClientNotificationTypeOnly <- &payload :
					slog.Info("comment notification is successfully sent to hub's TargettedClientNotificationTypeOnly chan")
				default :
						slog.Error("CONSUMER unable to send to TargettedBrokerMessages", "error", "channel full or blocked", "receiver", payload.RecieverID)	
				}
			} 

			slog.Debug("forwarded to hub", "sender_id", payload.SenderID, "receiver_id", payload.RecieverID)
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection
func (p *PubSubBroker) Close() error {
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	slog.Info("RabbitMQ connection closed")
	return nil
}

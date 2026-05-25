package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/rabbitmq/amqp091-go"
)

const notificationsExchangeName = "notifications_topic"

func main() {
	var (
		rabbitURL = flag.String("rabbit", "amqp://guest:guest@localhost:5672/", "RabbitMQ URL")
		userID    = flag.Uint("user", 0, "target user id (routing key user.<id>)")
		senderID  = flag.Uint("sender", 0, "sender id")
		typ       = flag.String("type", "test", "notification type")
		content   = flag.String("content", "hello from publish_test", "notification content")
		postID    = flag.Uint("post", 0, "post id")
	)
	flag.Parse()

	if *userID == 0 {
		fmt.Fprintln(os.Stderr, "--user is required and must be > 0")
		os.Exit(2)
	}

	conn, err := amqp091.Dial(*rabbitURL)
	if err != nil {
		slog.Error("failed to dial rabbitmq", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("failed to open channel", "error", err)
		os.Exit(1)
	}
	defer ch.Close()

	// Ensure exchange exists (idempotent)
	if err := ch.ExchangeDeclare(notificationsExchangeName, "topic", true, false, false, false, nil); err != nil {
		slog.Error("failed to declare exchange", "error", err)
		os.Exit(1)
	}

	payload := &services.ClientNotifyPayload{
		SenderID:   uint(*senderID),
		RecieverID: uint(*userID),
		Type:       *typ,
		Content:    *content,
		PostID:     uint(*postID),
		CreatedAt:  time.Now(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal payload", "error", err)
		os.Exit(1)
	}

	routingKey := fmt.Sprintf("user.%d", *userID)
	if err := ch.Publish(notificationsExchangeName, routingKey, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        body,
	}); err != nil {
		slog.Error("failed to publish", "error", err)
		os.Exit(1)
	}

	slog.Info("published test notification", "routing_key", routingKey)
}

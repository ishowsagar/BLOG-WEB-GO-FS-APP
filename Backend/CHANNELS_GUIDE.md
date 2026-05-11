# Go Channels - Learning Guide for Your Blog App

## What are Channels?

Channels are pipes that allow goroutines (lightweight threads) to communicate safely. Think of them as **message queues between concurrent tasks**.

---

## 1. BASICS - Creating & Using Channels

### Unbuffered Channel (waits for receiver)

```go
// Create a channel
messageChan := make(chan string)

// Send data (blocks until someone receives)
go func() {
    messageChan <- "Hello from goroutine"
}()

// Receive data (blocks until something arrives)
message := <-messageChan
fmt.Println(message) // Output: Hello from goroutine
```

### Buffered Channel (queue with size limit)

```go
// Create channel that can hold 2 messages
logChan := make(chan string, 2)

// Send 2 messages without waiting for receivers
logChan <- "Log message 1"
logChan <- "Log message 2"

// Now receive
fmt.Println(<-logChan) // Log message 1
fmt.Println(<-logChan) // Log message 2
```

---

## 2. USE CASE 1: Async Database Operations (Your Posts Repo)

**Problem**: When creating a post, if DB is slow, user waits.

**Solution**: Use goroutine + channel to handle DB write in background.

```go
// In services/posts_repo.go

// Add this function for async creation
func (p *PostDBModel) CreatePostAsync(post models.Post) chan *models.Post {
    // Create buffered channel (1 slot)
    resultChan := make(chan *models.Post, 1)

    // Do DB work in background
    go func() {
        defer close(resultChan) // Important: close when done

        ctx, timeout := context.WithTimeout(context.Background(), utils.DbTimeoutDuration)
        defer timeout()

        query := `
            insert into posts(user_id,title,content,created_at,updated_at)
            values($1,$2,$3,$4,$5)
            returning id,user_id,title,content,created_at,updated_at
        `

        resRow := p.db.QueryRowContext(ctx, query,
            post.UserID, post.Title, post.Content, time.Now(), time.Now(),
        )

        var postVar models.Post
        err := resRow.Scan(
            &postVar.ID, &postVar.UserID, &postVar.Title,
            &postVar.Content, &postVar.CreatedAt, &postVar.UpdatedAt,
        )

        if err != nil {
            slog.Error("CreatePost failed", "error", err)
            return
        }

        // Send result through channel
        resultChan <- &postVar
    }()

    return resultChan
}

// Usage in your controller:
// postChan := postRepo.CreatePostAsync(newPost)
// userPost := <-postChan  // Wait for result when needed
// fmt.Println(userPost.ID)
```

---

## 3. USE CASE 2: Batch Notifications (When Post is Created)

**Scenario**: After a post is created, notify followers via email/websocket.

```go
package services

// NotificationService handles async notifications
type NotificationService struct {
    notifChan chan Notification
}

type Notification struct {
    UserID    uint
    PostID    uint
    Message   string
}

func NewNotificationService() *NotificationService {
    ns := &NotificationService{
        notifChan: make(chan Notification, 10), // Queue 10 notifications
    }

    // Start background worker that processes notifications
    go ns.processNotifications()

    return ns
}

// Worker that processes notifications one by one
func (ns *NotificationService) processNotifications() {
    for notif := range ns.notifChan { // Loop until channel closes
        // Send email, update DB, etc (slow operations)
        sendEmail(notif.UserID, notif.Message)

        slog.Info("Notification sent", "userID", notif.UserID)
    }
}

// Called from your post controller
func (ns *NotificationService) NotifyFollowers(post models.Post) {
    // Non-blocking send (happens immediately)
    ns.notifChan <- Notification{
        UserID:  post.UserID,
        PostID:  post.ID,
        Message: "New post created!",
    }
}

// Usage:
// notifService := NewNotificationService()
// notifService.NotifyFollowers(newPost)  // Returns immediately
```

---

## 4. USE CASE 3: Fan-Out Pattern (Multiple DB Writes Concurrently)

**Scenario**: When user follows someone, update multiple tables at once.

```go
// In services/user_interactions.go

func (s *UserService) FollowUserConcurrent(followerID, followeeID uint) error {
    // Create channels for each operation
    errorChan := make(chan error, 3) // 3 concurrent operations

    // Operation 1: Insert into follows table
    go func() {
        query := "INSERT INTO follows(follower_id, followee_id) VALUES($1, $2)"
        _, err := s.db.Exec(query, followerID, followeeID)
        errorChan <- err
    }()

    // Operation 2: Increment follower count
    go func() {
        query := "UPDATE users SET follower_count = follower_count + 1 WHERE id = $1"
        _, err := s.db.Exec(query, followeeID)
        errorChan <- err
    }()

    // Operation 3: Increment following count
    go func() {
        query := "UPDATE users SET following_count = following_count + 1 WHERE id = $1"
        _, err := s.db.Exec(query, followerID)
        errorChan <- err
    }()

    // Wait for all 3 to complete and check for errors
    for i := 0; i < 3; i++ {
        if err := <-errorChan; err != nil {
            return err // If any fails, return immediately
        }
    }

    return nil // All succeeded
}
```

---

## 5. USE CASE 4: Timeout Pattern (Prevent Hanging)

```go
// In your post repo - prevent infinite waiting
func (p *PostDBModel) LoadFeedWithTimeout(limit int) ([]*models.Post, error) {
    resultChan := make(chan []*models.Post)
    errorChan := make(chan error)

    go func() {
        // Do expensive DB query
        posts, err := p.loadFeedFromDB(limit)

        if err != nil {
            errorChan <- err
        } else {
            resultChan <- posts
        }
    }()

    // Set 5-second timeout
    select {
    case posts := <-resultChan:
        return posts, nil
    case err := <-errorChan:
        return nil, err
    case <-time.After(5 * time.Second):
        return nil, errors.New("feed load timeout")
    }
}
```

---

## 6. CRITICAL RULES

### Rule 1: Always Close Channels

```go
// WRONG - goroutine panic if you send after close
close(resultChan) // Do this after goroutine finishes

// RIGHT - close only from sender side
go func() {
    // Do work
    resultChan <- data
    close(resultChan) // Close when done
}()
```

### Rule 2: Only One Sender Closes

```go
// WRONG - Multiple senders closing causes panic
go func() { close(ch) }()
go func() { close(ch) }() // PANIC!

// RIGHT - One place closes
resultChan := make(chan string)
go func() {
    sender1 := <-resultChan
    sender2 := <-resultChan
    close(resultChan) // Only one closer
}()
```

### Rule 3: Reading from Closed Channel

```go
// Reading from closed channel returns zero value, no error
ch := make(chan int)
close(ch)

val := <-ch     // Returns 0 (zero int)
val, ok := <-ch // Returns 0, false (ok is false = channel closed)
```

---

## 7. WHERE TO USE IN YOUR BLOG APP

### ✅ DO USE CHANNELS FOR:

1. **Email/Notification sending** - After post created, send async notification
2. **Bulk DB operations** - Concurrent inserts, updates with fan-out
3. **File uploads** - Async image processing after post created
4. **Rate limiting** - Queue requests through channel
5. **Event logging** - Centralized logging via channel worker

### ❌ DON'T USE CHANNELS FOR:

1. **Simple sequential DB queries** - Just call them directly
2. **Request/response in controller** - Use function returns
3. **Storing state** - Use variables, not channels

---

## 8. PRACTICAL EXAMPLE FOR YOUR BLOG APP

Let me show you how to add **async post creation with notifications**:

### Step 1: Create notification service

File: `Backend/services/notification_service.go`

```go
package services

import (
	"log/slog"
	"github.com/ishowsagar/go-blog-web-application/models"
)

type PostNotificationService struct {
	notifChan chan models.Post
}

func NewPostNotificationService() *PostNotificationService {
	pns := &PostNotificationService{
		notifChan: make(chan models.Post, 20), // Buffer 20 notifications
	}

	// Start background worker
	go pns.worker()

	return pns
}

func (pns *PostNotificationService) NotifyPostCreated(post models.Post) {
	// Non-blocking send
	select {
	case pns.notifChan <- post:
		slog.Info("Post notification queued", "postID", post.ID)
	default:
		slog.Warn("Notification queue full")
	}
}

func (pns *PostNotificationService) worker() {
	for post := range pns.notifChan {
		// Do slow work here without blocking caller
		slog.Info("Processing notification for post", "postID", post.ID)

		// Example: Save to DB, send email, etc
		// sendEmailToFollowers(post.UserID)
		// logPostCreationEvent(post.ID)
	}
}
```

### Step 2: Use in controller

File: `Backend/controller/posts_handler.go` (modify CreatePost)

```go
func (pc *PostController) CreatePost(c *gin.Context) {
	userID := c.GetUint("user_id")
	var postReq PostRequest

	if err := c.BindJSON(&postReq); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// Create post
	post := models.Post{
		UserID:  userID,
		Title:   postReq.Title,
		Content: postReq.Content,
	}

	createdPost, err := pc.postRepo.CreatePost(post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// ✨ NEW: Send notification async (doesn't block response)
	pc.notificationService.NotifyPostCreated(*createdPost)

	// Return response immediately
	c.JSON(http.StatusCreated, createdPost)
}
```

---

## 9. FLOW DIAGRAM

```
User Request
    |
    v
CreatePost Controller
    |
    +-> DB Insert (fast)
    |
    +-> Notification Channel Send (non-blocking)
    |       |
    |       v (in background)
    |       Worker: Email to followers
    |       Worker: Update cache
    |       Worker: Log event
    |
    +-> Return Response to User (immediately)
```

---

## Summary

- **Channels** = communication between goroutines
- **Goroutines** = lightweight concurrency
- **Buffered channels** = queue, don't wait for receiver
- **Use for**: async work (notifications, bulk ops, timeouts)
- **Remember**: close from sender, only one closer

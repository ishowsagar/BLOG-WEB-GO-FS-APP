# Database Schema Documentation
## Social Media Application (Instagram-like)

---

## Overview
This is a PostgreSQL database for a social media application with users, posts, comments, likes, follows, and direct messaging functionality.

---

## Tables Summary

### 1. users
**Purpose**: Store user account information and profile data

**Columns**:
- `id` (int8, PRIMARY KEY) - Unique user identifier
- `name` (text) - User's full name
- `email` (text) - User's email address
- `password` (text) - Hashed password
- `username` (text) - Unique username (default: 'insta_user12345')
- `nickname` (text) - Display name (default: 'User')
- `bio` (text) - User biography (default: 'New to instagram, follow me')
- `followers_count` (int8) - Number of followers (default: 0)
- `following_count` (int8) - Number of users being followed (default: 0)
- `created_at` (timestamptz) - Account creation timestamp

**Relationships**:
- Has many: posts, comments, likes, tokens
- Has many follows (as follower)
- Has many follows (as followee)
- Sends messages (as sender)
- Receives messages (as receiver)

---

### 2. posts
**Purpose**: Store user-generated content/posts

**Columns**:
- `id` (int8, PRIMARY KEY) - Unique post identifier
- `user_id` (int8, FOREIGN KEY → users.id) - Post author
- `title` (text) - Post title
- `content` (text) - Post content/body
- `created_at` (timestamptz) - Post creation timestamp
- `updated_at` (timestamptz) - Last update timestamp

**Relationships**:
- Belongs to: users (via user_id)
- Has many: comments, likes

**Constraints**:
- ON DELETE CASCADE - deleting a user deletes their posts
- ON UPDATE CASCADE

---

### 3. comments
**Purpose**: Store comments on posts

**Columns**:
- `id` (int8, PRIMARY KEY) - Unique comment identifier
- `user_id` (int8, FOREIGN KEY → users.id) - Comment author
- `post_id` (int8, FOREIGN KEY → posts.id) - Post being commented on
- `content` (text) - Comment text
- `created_at` (timestamptz) - Comment creation timestamp
- `updated_at` (timestamptz) - Last update timestamp

**Relationships**:
- Belongs to: users (via user_id)
- Belongs to: posts (via post_id)

**Constraints**:
- ON DELETE CASCADE - deleting a user or post deletes associated comments
- ON UPDATE CASCADE

---

### 4. likes
**Purpose**: Track post likes by users

**Columns**:
- `id` (int8, PRIMARY KEY) - Unique like identifier
- `user_id` (int8, FOREIGN KEY → users.id) - User who liked
- `post_id` (int8, FOREIGN KEY → posts.id) - Post being liked
- `like_count` (int8) - Like count (Note: unusual design, typically 1 row = 1 like)
- `liked_at` (timestamptz) - Like timestamp

**Relationships**:
- Belongs to: users (via user_id)
- Belongs to: posts (via post_id)

**Constraints**:
- ON DELETE CASCADE for user (deleting user removes their likes)

**Note**: The `like_count` column suggests this might be tracking cumulative likes rather than individual like records. Typical design would be one row per user per post.

---

### 5. follows
**Purpose**: Track follower/following relationships between users

**Columns**:
- `id` (int8, PRIMARY KEY) - Unique follow relationship identifier
- `follower_id` (int8, FOREIGN KEY → users.id) - User who follows
- `followee_id` (int8, FOREIGN KEY → users.id) - User being followed
- `followed_at` (timestamptz) - Follow timestamp

**Relationships**:
- Belongs to: users (via follower_id) - the follower
- Belongs to: users (via followee_id) - the person being followed

**Constraints**:
- ON DELETE CASCADE - deleting a user removes all follow relationships
- ON UPDATE CASCADE
- UNIQUE INDEX on (follower_id, followee_id) - prevents duplicate follows

**Important**: 
- `follower_id` = the person doing the following
- `followee_id` = the person being followed

---

### 6. messages
**Purpose**: Store direct messages between users

**Columns**:
- `id` (int8, PRIMARY KEY) - Unique message identifier
- `sender_id` (int8) - User who sent the message
- `reciever_id` (int8) - User who receives the message (Note: typo in column name)
- `content` (text) - Message text
- `created_at` (timestamptz) - Message timestamp

**Relationships**:
- Belongs to: users (via sender_id)
- Belongs to: users (via reciever_id)

**Note**: No foreign key constraints defined. Column name has typo: "reciever_id" should be "receiver_id"

---

### 7. tokens
**Purpose**: Store authentication tokens for user sessions

**Columns**:
- `id` (int8, PRIMARY KEY) - Unique token identifier
- `user_id` (int8, FOREIGN KEY → users.id) - Token owner
- `hash` (text) - Token hash
- `expiry` (timestamptz) - Token expiration time

**Relationships**:
- Belongs to: users (via user_id)

**Constraints**:
- ON DELETE CASCADE - deleting user removes their tokens
- ON UPDATE CASCADE

---

## Common Query Patterns

### User Queries
- Get user profile by username/email
- Get user's followers/following lists
- Update follower/following counts
- Search users by name/username

### Post Queries
- Get all posts by a user
- Get posts from users I follow (feed)
- Get recent posts (sorted by created_at)
- Get post with all comments and likes

### Social Interactions
- Check if user A follows user B
- Get mutual followers
- Get users who liked a post
- Get comment count per post
- Get like count per post

### Feed & Timeline
- Get feed (posts from followed users, sorted by time)
- Get trending posts (most liked/commented)
- Get user's own posts timeline

### Messaging
- Get conversation between two users
- Get all conversations for a user
- Mark messages as read (requires additional column)

---

## Database Design Notes

### Potential Issues:
1. **likes.like_count**: Unusual design - typically each like is one row with like_count always = 1
2. **messages**: Missing foreign key constraints
3. **messages.reciever_id**: Typo in column name
4. **users.followers_count/following_count**: Denormalized data - needs to be kept in sync with follows table
5. **No indexes**: Missing indexes on foreign keys which could slow down queries
6. **messages**: No read/unread status tracking

### Recommended Indexes (if not already created):
```sql
-- Foreign key indexes for better join performance
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_likes_post_id ON likes(post_id);
CREATE INDEX idx_likes_user_id ON likes(user_id);
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_receiver ON messages(reciever_id);

-- Query optimization indexes
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
```

---

## Sample Relationships Visualization

```
users (1) ─────< posts (many)
  │              │
  │              ├─< comments (many)
  │              └─< likes (many)
  │
  ├─< follows (many as follower)
  ├─< follows (many as followee)
  ├─< tokens (many)
  ├─< messages (many as sender)
  └─< messages (many as receiver)
```

---

## When Using This Schema with AI

**To get help with queries, provide:**
1. This documentation file
2. What you're trying to achieve (e.g., "get user feed")
3. Any specific filtering/sorting requirements
4. Performance concerns if dealing with large datasets

**Common requests:**
- "Show me all posts from users I follow"
- "Get the 10 most liked posts this week"
- "Find mutual followers between two users"
- "Get a user's complete profile with stats"
- "Retrieve a conversation between two users"

---

Generated: 2026-05-15
Database Type: PostgreSQL
Application Type: Social Media Platform

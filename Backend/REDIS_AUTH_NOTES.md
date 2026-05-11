# Redis-backed auth cache guide

This note explains how to implement Redis for login/auth in this backend, using the current code structure as the example.

## What the current app does

Right now the flow is:

1. `LoginUser` checks email and password.
2. It generates a JWT token.
3. It stores a hashed token in the `tokens` table.
4. `AuthMiddleware` parses the JWT on every protected request.
5. The middleware then looks up the token row in Postgres and checks hash + expiry.

That means the database is the source of truth for auth state, not a cache.

## What Redis should do here

Redis should store short-lived auth session data so protected requests do not hit Postgres every time.

Use Redis for:

- fast token/session lookup in middleware
- automatic expiry with TTL
- quick session invalidation after password reset or logout

Keep Postgres for:

- durable source of truth
- fallback when Redis is empty or restarted

## General Redis example

This is the smallest mental model for Redis in Go:

```go
ctx := context.Background()
client := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

err := client.Set(ctx, "hello", "world", 10*time.Minute).Err()
if err != nil {
    return err
}

value, err := client.Get(ctx, "hello").Result()
if err != nil {
    return err
}

fmt.Println(value)
```

The theory behind it is simple:

- `Set` writes data into Redis.
- `Get` reads it back.
- `time.Minute` or `time.Hour` gives it a TTL.
- When TTL ends, Redis removes it automatically.

## Simple mental model

Think of it like this:

- JWT proves the request came from your server.
- Redis says whether the session is still active.
- Postgres is the backup record if Redis misses.

## Minimal data shape

Store one cache entry per user session.

Suggested Redis key:

- `auth:token:user:<user_id>`

Suggested value:

- token hash
- expiry time

You can store the value as JSON or a simple string format. JSON is easier to read while learning.

## Cache helper example

If you make a small service for auth sessions, it can look like this:

```go
type TokenCacheModel struct {
    Client *redis.Client
}

func NewTokenCacheModel(addr, password string, db int) *TokenCacheModel {
    return &TokenCacheModel{
        Client: redis.NewClient(&redis.Options{
            Addr:     addr,
            Password: password,
            DB:       db,
        }),
    }
}

func (t *TokenCacheModel) SetToken(userID uint, tokenHash string, expiry time.Time) error {
    key := fmt.Sprintf("auth:token:user:%d", userID)
    payload := map[string]any{
        "hash":   tokenHash,
        "expiry": expiry,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    ttl := time.Until(expiry)
    return t.Client.Set(context.Background(), key, body, ttl).Err()
}

func (t *TokenCacheModel) GetTokenByUserID(userID uint) (string, time.Time, error) {
    key := fmt.Sprintf("auth:token:user:%d", userID)

    result, err := t.Client.Get(context.Background(), key).Result()
    if err != nil {
        return "", time.Time{}, err
    }

    var payload struct {
        Hash   string    `json:"hash"`
        Expiry time.Time `json:"expiry"`
    }

    if err := json.Unmarshal([]byte(result), &payload); err != nil {
        return "", time.Time{}, err
    }

    return payload.Hash, payload.Expiry, nil
}

func (t *TokenCacheModel) DeleteTokenByUserID(userID uint) error {
    key := fmt.Sprintf("auth:token:user:%d", userID)
    return t.Client.Del(context.Background(), key).Err()
}
```

That is the core shape you want to understand. The important thing is not the exact syntax, but the flow:

- create client once
- write token data on login
- read token data in middleware
- delete token data on logout or password reset

## Request flow after Redis is added

### Login

1. Read email and password.
2. Fetch the user from the database.
3. Verify the password hash.
4. Generate the JWT.
5. Hash the token string.
6. Save the session to Postgres.
7. Save the same session to Redis with TTL matching token expiry.
8. Return the plain JWT to the client.

Example login write-through logic:

```go
plainTokenString, err := utils.GenerateToken(userFound.ID)
if err != nil {
    return err
}

hashedTokenByte, err := utils.HashToken(*plainTokenString)
if err != nil {
    return err
}

token := models.Token{
    UserID: userFound.ID,
    Hash:   string(hashedTokenByte),
    Expiry: time.Now().UTC().Add(24 * time.Hour),
}

err = tokenDB.InsertToken(token)
if err != nil {
    return err
}

err = tokenCache.SetToken(token)
if err != nil {
    slog.Warn("cache write failed", "error", err)
}
```

### Protected request

1. Read the `Authorization` header.
2. Parse and verify the JWT signature.
3. Extract `user_id` from claims.
4. Look up `auth:token:user:<user_id>` in Redis.
5. If found, compare hash and expiry.
6. If Redis misses, fall back to Postgres.
7. If Postgres succeeds, refresh Redis.
8. Set `user_id` in context and continue.

Example middleware flow:

```go
tokenStr := c.GetHeader("Authorization")
if tokenStr == "" {
    c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
    return
}

token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
    return []byte(jwtSecretSigningKey), nil
})
if err != nil || token == nil || !token.Valid {
    c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
    return
}

claims := token.Claims.(jwt.MapClaims)
userID := uint(claims["user_id"].(float64))

cachedHash, cachedExpiry, err := tokenCache.GetTokenByUserID(userID)
if err == nil {
    incomingHash, _ := utils.HashToken(tokenStr)
    if string(incomingHash) == cachedHash && time.Now().Before(cachedExpiry) {
        c.Set("user_id", userID)
        c.Next()
        return
    }
}

dbToken, err := tokenDB.GetTokenByUserID(userID)
if err != nil {
    c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session not found"})
    return
}

_ = tokenCache.SetToken(*dbToken)
c.Set("user_id", userID)
c.Next()
```

### Password reset or logout

1. Update the password or end the session.
2. Delete the row from Postgres.
3. Delete the Redis key.

That prevents old sessions from staying valid.

Example invalidation:

```go
err := tokenDB.DeleteTokenByUserID(userID)
if err != nil {
    return err
}

err = tokenCache.DeleteTokenByUserID(userID)
if err != nil {
    slog.Warn("cache delete failed", "error", err)
}
```

## Where each change belongs in this repo

### 1. Startup wiring

Add Redis client creation in `cmd/server/main.go`.

This is where the app already creates database models and controllers, so Redis should be created there too and passed down like the DB models.

### 2. Config loading

Add Redis settings in `utils/env.go`.

Typical env values:

- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `REDIS_DB`

### 3. Cache service

Create a small service in `services/` for Redis token operations.

That service should expose:

- `SetToken(...)`
- `GetTokenByUserID(...)`
- `DeleteTokenByUserID(...)`

### 4. Login handler

After token creation in `controller/users_handler.go`, write the session to Redis.

The DB insert still stays, because Redis should not be your only auth record.

### 5. Auth middleware

Change `middleware/auth_middleware.go` so it checks Redis first.

Only hit the database if Redis does not have the session.

### 6. Password reset

When a password changes, remove the old token from both Postgres and Redis.

That is important because password reset should invalidate existing sessions.

## What to learn from this pattern

This is a read-through cache pattern:

- read from cache first
- fall back to database
- fill cache on miss
- invalidate on write

For login auth, that pattern usually gives the best balance between speed and correctness.

## Good learning sequence

If you want to implement this yourself, do it in this order:

1. Add Redis env config.
2. Connect to Redis in app startup.
3. Create a Redis token service.
4. Save tokens to Redis during login.
5. Read tokens from Redis in middleware.
6. Fall back to Postgres on cache miss.
7. Delete cache on password reset/logout.

## Important tradeoffs

- Redis is fast, but ephemeral.
- Postgres is durable, but slower for high-frequency auth checks.
- If Redis restarts, your app should still work because Postgres remains available.
- Do not store plain passwords in Redis.
- Keep token TTL aligned with expiry so stale sessions disappear automatically.

## Short version

If you remember only one thing, remember this:

> JWT validates identity, Redis stores active session state, Postgres backs it up.

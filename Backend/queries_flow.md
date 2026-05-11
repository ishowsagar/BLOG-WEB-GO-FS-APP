# SQL Query Methods with Context

## Methods

### `QueryContext(ctx, query, args...)`

- **Use for:** SELECT queries returning **multiple rows**
- **Returns:** `*sql.Rows` (iterate with `.Next()`)
- **Example:** Get all posts by a user

### `QueryRowContext(ctx, query, args...)`

- **Use for:** SELECT queries returning **single row**
- **Returns:** `*sql.Row` (scan with `.Scan()`)
- **Example:** Get user by ID

### `ExecContext(ctx, query, args...)`

- **Use for:** INSERT, UPDATE, DELETE
- **Returns:** `sql.Result` (LastInsertId, RowsAffected)
- **Example:** Create user, update post

## For Your CreateUser:

Use **`ExecContext()`** since you're inserting:

```go
result, err := u.db.ExecContext(ctx, query, user.Name, user.Email, user.Password, time.Now())
if err != nil {
    return nil, err
}
id, _ := result.LastInsertId()
user.ID = uint(id)
return &user, nil


psql= `psql -U blog_admin -h localhost -p 5432 -d blogdb`

// added likeCount for models and also non-migrated, just rendr field and fetch from query to store for response, not for db by using "-"
```

# 🗄️ Phase 9: Data Persistence

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 8](../Phase_8/Phase_8.md)

---

**Objective:** Efficient data storage and retrieval patterns in Go.

**Reference:** [database/sql Package](https://pkg.go.dev/database/sql), [pgx Documentation](https://github.com/jackc/pgx)

**Prerequisites:** Phase 0-8

**Estimated Duration:** 2-3 weeks

---

## 📋 Table of Contents

1. [The `database/sql` Package](#91-the-databasesql-package)
2. [Connection Pool Management](#92-connection-pool-management)
3. [Querying Data](#93-querying-data)
4. [PostgreSQL with `pgx`](#94-postgresql-with-pgx)
5. [Transactions](#95-transactions)
6. [Query Building Approaches](#96-query-building-approaches)
7. [NoSQL in Go](#97-nosql-in-go)
8. [Migrations](#98-migrations)
9. [Best Practices](#99-best-practices)
10. [Interview Questions](#interview-questions)

---

## 9.1 The `database/sql` Package

### Driver Architecture

**Interview Question:** *"How does Go's database/sql package work with different databases?"*

```go
// database/sql provides a generic interface
// Drivers implement the actual database communication

import (
    "database/sql"
    _ "github.com/lib/pq"           // PostgreSQL driver
    _ "github.com/go-sql-driver/mysql"  // MySQL driver
    _ "github.com/mattn/go-sqlite3"     // SQLite driver
)

// The blank identifier import registers the driver
// Driver registration happens in init()

func main() {
    // Open database connection
    // sql.Open doesn't actually connect - it prepares the connection pool
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Ping verifies connection is working
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Connected to database")
}
```

### Connection String Formats

```go
// PostgreSQL
"postgres://user:password@localhost:5432/dbname?sslmode=disable"
"host=localhost port=5432 user=user password=pass dbname=mydb sslmode=disable"

// MySQL
"user:password@tcp(localhost:3306)/dbname?parseTime=true"

// SQLite
"file:test.db?cache=shared&mode=memory"
"/path/to/database.db"
```

### sql.DB Is a Pool, Not a Connection

**Interview Question:** *"What is sql.DB and how does it manage connections?"*

```go
// sql.DB is NOT a single connection - it's a connection pool!

db, _ := sql.Open("postgres", connStr)

// These configure the pool:
db.SetMaxOpenConns(25)                  // Max open connections
db.SetMaxIdleConns(5)                   // Max idle connections
db.SetConnMaxLifetime(5 * time.Minute)  // Max connection lifetime
db.SetConnMaxIdleTime(1 * time.Minute)  // Max idle time before close

// sql.DB is safe for concurrent use
// Share one instance across your application
```

---

## 9.2 Connection Pool Management

### Configuring the Pool

**Interview Question:** *"How do you properly configure database connection pool settings?"*

```go
func NewDB(connStr string) (*sql.DB, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }
    
    // Pool configuration
    db.SetMaxOpenConns(25)     // Limit total connections
    db.SetMaxIdleConns(5)      // Keep some connections warm
    db.SetConnMaxLifetime(5 * time.Minute)  // Prevent stale connections
    db.SetConnMaxIdleTime(1 * time.Minute)
    
    // Verify connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("ping database: %w", err)
    }
    
    return db, nil
}
```

### Pool Sizing Guidelines

```
MaxOpenConns:
- Start with: 25 connections
- Formula: (Number of CPUs * 2) + effective_spindle_count
- For cloud: Often limited by database tier

MaxIdleConns:
- Keep a subset of MaxOpenConns
- Too high: Wastes resources
- Too low: Connection churn

ConnMaxLifetime:
- Set to less than database timeout
- Helps with load balancer failover
- 5 minutes is a good default

ConnMaxIdleTime:
- Close idle connections faster
- Helps in variable load scenarios
- 1-5 minutes typical
```

### Monitoring Pool Health

```go
// Get pool statistics
stats := db.Stats()

fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("In use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
fmt.Printf("Wait count: %d\n", stats.WaitCount)
fmt.Printf("Wait duration: %v\n", stats.WaitDuration)
fmt.Printf("Max idle closed: %d\n", stats.MaxIdleClosed)
fmt.Printf("Max lifetime closed: %d\n", stats.MaxLifetimeClosed)
```

---

## 9.3 Querying Data

### Query Methods

**Interview Question:** *"What's the difference between Query, QueryRow, and Exec?"*

```go
// Query - returns multiple rows
rows, err := db.QueryContext(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    return err
}
defer rows.Close()  // IMPORTANT: Always close!

for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        return err
    }
    // Process row...
}
// Check for iteration errors
if err := rows.Err(); err != nil {
    return err
}

// QueryRow - returns single row
var name string
err := db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", 1).Scan(&name)
if err == sql.ErrNoRows {
    // Handle not found
} else if err != nil {
    return err
}

// Exec - no rows returned (INSERT, UPDATE, DELETE)
result, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", 1)
if err != nil {
    return err
}
rowsAffected, _ := result.RowsAffected()
lastInsertID, _ := result.LastInsertId()  // Not supported by all drivers
```

### Scanning Into Structs

```go
type User struct {
    ID        int64
    Name      string
    Email     string
    CreatedAt time.Time
}

func GetUser(ctx context.Context, db *sql.DB, id int64) (*User, error) {
    query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
    
    var user User
    err := db.QueryRowContext(ctx, query, id).Scan(
        &user.ID,
        &user.Name,
        &user.Email,
        &user.CreatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("scan user: %w", err)
    }
    
    return &user, nil
}

func ListUsers(ctx context.Context, db *sql.DB) ([]User, error) {
    query := `SELECT id, name, email, created_at FROM users ORDER BY created_at DESC`
    
    rows, err := db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    
    return users, rows.Err()
}
```

### Handling NULL Values

**Interview Question:** *"How do you handle NULL values from the database in Go?"*

```go
// Option 1: sql.Null* types
type User struct {
    ID       int64
    Name     string
    Nickname sql.NullString  // Can be NULL
    Age      sql.NullInt64   // Can be NULL
}

err := rows.Scan(&u.ID, &u.Name, &u.Nickname, &u.Age)

if u.Nickname.Valid {
    fmt.Println("Nickname:", u.Nickname.String)
}

// Option 2: Pointer types
type User struct {
    ID       int64
    Name     string
    Nickname *string  // nil if NULL
    Age      *int     // nil if NULL
}

// Option 3: COALESCE in SQL
query := `SELECT id, name, COALESCE(nickname, '') as nickname FROM users`
```

### Prepared Statements

```go
// Prepare statement once, use multiple times
stmt, err := db.PrepareContext(ctx, "SELECT id, name FROM users WHERE status = $1")
if err != nil {
    return err
}
defer stmt.Close()

// Execute with different parameters
rows1, _ := stmt.QueryContext(ctx, "active")
rows2, _ := stmt.QueryContext(ctx, "pending")

// Benefits:
// - Parse SQL once
// - Better performance for repeated queries
// - Protection against SQL injection
```

---

## 9.4 PostgreSQL with `pgx`

### Why pgx?

**Interview Question:** *"What are the advantages of pgx over database/sql?"*

```go
// pgx is a PostgreSQL-specific driver with enhanced features

// Benefits:
// - Native PostgreSQL protocol (binary, not text)
// - Better performance
// - Rich type support (arrays, JSON, HSTORE)
// - COPY protocol for bulk operations
// - Listen/Notify support
// - Connection pooling (pgxpool)

import (
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)
```

### Connection Pool with pgxpool

```go
func NewPool(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(connStr)
    if err != nil {
        return nil, err
    }
    
    // Pool configuration
    config.MaxConns = 25
    config.MinConns = 5
    config.MaxConnLifetime = 5 * time.Minute
    config.MaxConnIdleTime = 1 * time.Minute
    config.HealthCheckPeriod = 30 * time.Second
    
    // Create pool
    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, err
    }
    
    // Verify connection
    if err := pool.Ping(ctx); err != nil {
        return nil, err
    }
    
    return pool, nil
}
```

### Querying with pgx

```go
// Single row
var name string
err := pool.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", 1).Scan(&name)

// Multiple rows
rows, err := pool.Query(ctx, "SELECT id, name FROM users")
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    rows.Scan(&id, &name)
}

// Using pgx.CollectRows (cleaner!)
users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
```

### Named Struct Scanning

```go
type User struct {
    ID        int64     `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
}

func GetUser(ctx context.Context, pool *pgxpool.Pool, id int64) (*User, error) {
    rows, err := pool.Query(ctx, 
        `SELECT id, name, email, created_at FROM users WHERE id = $1`, id)
    if err != nil {
        return nil, err
    }
    
    user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound
    }
    return &user, err
}
```

### COPY Protocol for Bulk Operations

**Interview Question:** *"How would you efficiently insert millions of rows?"*

```go
func BulkInsertUsers(ctx context.Context, pool *pgxpool.Pool, users []User) error {
    // COPY is orders of magnitude faster than INSERT for bulk data
    _, err := pool.CopyFrom(
        ctx,
        pgx.Identifier{"users"},  // Table name
        []string{"name", "email", "created_at"},  // Columns
        pgx.CopyFromSlice(len(users), func(i int) ([]any, error) {
            return []any{
                users[i].Name,
                users[i].Email,
                users[i].CreatedAt,
            }, nil
        }),
    )
    return err
}

// Performance comparison:
// INSERT: ~1,000 rows/sec
// Batch INSERT: ~10,000 rows/sec
// COPY: ~100,000+ rows/sec
```

### Listen/Notify

```go
// Real-time notifications from PostgreSQL

func ListenForNotifications(ctx context.Context, pool *pgxpool.Pool) error {
    conn, err := pool.Acquire(ctx)
    if err != nil {
        return err
    }
    defer conn.Release()
    
    // Subscribe to channel
    _, err = conn.Exec(ctx, "LISTEN user_events")
    if err != nil {
        return err
    }
    
    // Wait for notifications
    for {
        notification, err := conn.Conn().WaitForNotification(ctx)
        if err != nil {
            return err
        }
        
        fmt.Printf("Channel: %s, Payload: %s\n", 
            notification.Channel, notification.Payload)
    }
}

// Trigger from another connection:
// NOTIFY user_events, '{"action": "created", "user_id": 123}'
```

---

## 9.5 Transactions

### Basic Transaction Pattern

**Interview Question:** *"How do you properly handle database transactions in Go?"*

```go
func TransferFunds(ctx context.Context, db *sql.DB, from, to int64, amount float64) error {
    // Begin transaction
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    
    // Defer rollback (no-op if committed)
    defer tx.Rollback()
    
    // Debit from account
    _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2",
        amount, from)
    if err != nil {
        return fmt.Errorf("debit: %w", err)
    }
    
    // Credit to account
    _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
        amount, to)
    if err != nil {
        return fmt.Errorf("credit: %w", err)
    }
    
    // Commit transaction
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit: %w", err)
    }
    
    return nil
}
```

### Transaction Helper Function

```go
// Generic transaction helper
func WithTransaction(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)  // Re-throw panic after rollback
        }
    }()
    
    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit()
}

// Usage
err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
    // All operations use tx
    _, err := tx.ExecContext(ctx, "UPDATE ...")
    if err != nil {
        return err
    }
    
    _, err = tx.ExecContext(ctx, "INSERT ...")
    return err
})
```

### Isolation Levels

```go
// Configure isolation level
tx, err := db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
})

// Available levels:
// sql.LevelDefault
// sql.LevelReadUncommitted
// sql.LevelReadCommitted
// sql.LevelWriteCommitted
// sql.LevelRepeatableRead
// sql.LevelSnapshot
// sql.LevelSerializable
// sql.LevelLinearizable
```

### Transaction with pgx

```go
func TransferWithPgx(ctx context.Context, pool *pgxpool.Pool, from, to int64, amount float64) error {
    return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
        // Debit
        _, err := tx.Exec(ctx,
            "UPDATE accounts SET balance = balance - $1 WHERE id = $2",
            amount, from)
        if err != nil {
            return err
        }
        
        // Credit
        _, err = tx.Exec(ctx,
            "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
            amount, to)
        return err
    })
    // Automatically commits on nil error, rolls back on error
}
```

---

## 9.6 Query Building Approaches

### Raw SQL (Recommended for Simple Cases)

```go
// Direct SQL - maximum control and performance
func GetActiveUsers(ctx context.Context, db *sql.DB, limit int) ([]User, error) {
    query := `
        SELECT id, name, email, created_at
        FROM users
        WHERE status = 'active'
        ORDER BY created_at DESC
        LIMIT $1
    `
    
    rows, err := db.QueryContext(ctx, query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    
    return users, rows.Err()
}
```

### SQLC (Code Generation from SQL)

**Interview Question:** *"What is SQLC and why might you choose it?"*

```sql
-- queries.sql
-- name: GetUser :one
SELECT id, name, email, created_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, email, created_at
FROM users
ORDER BY created_at DESC
LIMIT $1;

-- name: CreateUser :one
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at;
```

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries.sql"
    schema: "schema.sql"
    gen:
      go:
        package: "db"
        out: "internal/db"
```

```go
// Generated code - type-safe, compile-time checked
func main() {
    queries := db.New(pool)
    
    // Type-safe queries
    user, err := queries.GetUser(ctx, 1)
    users, err := queries.ListUsers(ctx, 10)
    newUser, err := queries.CreateUser(ctx, db.CreateUserParams{
        Name:  "Alice",
        Email: "alice@example.com",
    })
}
```

### Query Builder (squirrel)

```go
import sq "github.com/Masterminds/squirrel"

// Build queries programmatically
func SearchUsers(ctx context.Context, db *sql.DB, filters UserFilters) ([]User, error) {
    psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
    
    query := psql.Select("id", "name", "email", "created_at").
        From("users").
        OrderBy("created_at DESC")
    
    // Dynamic conditions
    if filters.Status != "" {
        query = query.Where(sq.Eq{"status": filters.Status})
    }
    if filters.Name != "" {
        query = query.Where(sq.Like{"name": "%" + filters.Name + "%"})
    }
    if filters.MinAge > 0 {
        query = query.Where(sq.GtOrEq{"age": filters.MinAge})
    }
    
    query = query.Limit(uint64(filters.Limit))
    
    sql, args, err := query.ToSql()
    if err != nil {
        return nil, err
    }
    
    rows, err := db.QueryContext(ctx, sql, args...)
    // ...
}
```

### ORM (GORM)

```go
import "gorm.io/gorm"

type User struct {
    ID        uint   `gorm:"primaryKey"`
    Name      string
    Email     string `gorm:"uniqueIndex"`
    CreatedAt time.Time
}

func main() {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    
    // Auto migrate
    db.AutoMigrate(&User{})
    
    // Create
    user := User{Name: "Alice", Email: "alice@example.com"}
    db.Create(&user)
    
    // Read
    var users []User
    db.Where("status = ?", "active").Find(&users)
    
    // Update
    db.Model(&user).Update("Name", "Bob")
    
    // Delete
    db.Delete(&user)
}
```

### Comparison

| Approach | Pros | Cons |
|----------|------|------|
| Raw SQL | Full control, best performance | Manual scanning, no type safety |
| SQLC | Type-safe, compile-time checked | Learning curve, code generation |
| Query Builder | Dynamic queries, type-safe | Runtime errors possible |
| ORM | Rapid development | Performance issues, magic |

**Go Community Preference:** Raw SQL or SQLC > Query Builders > ORMs

---

## 9.7 NoSQL in Go

### Redis

```go
import "github.com/redis/go-redis/v9"

func NewRedisClient() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:         "localhost:6379",
        Password:     "",
        DB:           0,
        PoolSize:     10,
        MinIdleConns: 5,
    })
}

func CacheUser(ctx context.Context, client *redis.Client, user *User) error {
    data, err := json.Marshal(user)
    if err != nil {
        return err
    }
    
    return client.Set(ctx, fmt.Sprintf("user:%d", user.ID), data, time.Hour).Err()
}

func GetCachedUser(ctx context.Context, client *redis.Client, id int64) (*User, error) {
    data, err := client.Get(ctx, fmt.Sprintf("user:%d", id)).Bytes()
    if err == redis.Nil {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    
    var user User
    return &user, json.Unmarshal(data, &user)
}
```

### MongoDB

```go
import "go.mongodb.org/mongo-driver/mongo"

func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        return nil, err
    }
    
    if err := client.Ping(ctx, nil); err != nil {
        return nil, err
    }
    
    return client, nil
}

func InsertUser(ctx context.Context, coll *mongo.Collection, user *User) error {
    _, err := coll.InsertOne(ctx, user)
    return err
}

func FindUser(ctx context.Context, coll *mongo.Collection, id string) (*User, error) {
    var user User
    err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
    if err == mongo.ErrNoDocuments {
        return nil, ErrNotFound
    }
    return &user, err
}
```

### Embedded Databases

```go
// SQLite - embedded SQL
import _ "github.com/mattn/go-sqlite3"

db, err := sql.Open("sqlite3", "./data.db")

// BadgerDB - embedded key-value
import "github.com/dgraph-io/badger/v4"

db, err := badger.Open(badger.DefaultOptions("./badger"))
defer db.Close()

// Write
err = db.Update(func(txn *badger.Txn) error {
    return txn.Set([]byte("key"), []byte("value"))
})

// Read
err = db.View(func(txn *badger.Txn) error {
    item, err := txn.Get([]byte("key"))
    if err != nil {
        return err
    }
    return item.Value(func(val []byte) error {
        fmt.Println(string(val))
        return nil
    })
})
```

---

## 9.8 Migrations

### golang-migrate/migrate

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration
migrate create -ext sql -dir migrations -seq create_users_table
```

```sql
-- migrations/000001_create_users_table.up.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

```sql
-- migrations/000001_create_users_table.down.sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

### Running Migrations

```bash
# Run migrations
migrate -database "postgres://user:pass@localhost/db?sslmode=disable" -path migrations up

# Rollback
migrate -database "..." -path migrations down 1

# Force version (for fixing issues)
migrate -database "..." -path migrations force 1
```

### Programmatic Migrations

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbURL string) error {
    m, err := migrate.New(
        "file://migrations",
        dbURL,
    )
    if err != nil {
        return err
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    return nil
}
```

---

## 9.9 Best Practices

### Repository Pattern

```go
// Define interface
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int64) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, limit, offset int) ([]User, error)
}

// Implementation
type PostgresUserRepository struct {
    db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
    return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
    
    var user User
    err := r.db.QueryRow(ctx, query, id).Scan(
        &user.ID, &user.Name, &user.Email, &user.CreatedAt,
    )
    
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound
    }
    
    return &user, err
}
```

### Error Handling

```go
// Domain errors
var (
    ErrNotFound      = errors.New("not found")
    ErrDuplicate     = errors.New("duplicate entry")
    ErrForeignKey    = errors.New("foreign key violation")
)

func (r *PostgresUserRepository) Create(ctx context.Context, user *User) error {
    query := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`
    
    err := r.db.QueryRow(ctx, query, user.Name, user.Email).Scan(&user.ID)
    if err != nil {
        // Check for specific PostgreSQL errors
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            switch pgErr.Code {
            case "23505": // unique_violation
                return ErrDuplicate
            case "23503": // foreign_key_violation
                return ErrForeignKey
            }
        }
        return fmt.Errorf("insert user: %w", err)
    }
    
    return nil
}
```

### Always Use Context

```go
// GOOD: Context for timeouts and cancellation
func (r *Repository) GetUser(ctx context.Context, id int64) (*User, error) {
    return r.db.QueryRowContext(ctx, query, id).Scan(...)
}

// BAD: No context - can't cancel, no timeout
func (r *Repository) GetUser(id int64) (*User, error) {
    return r.db.QueryRow(query, id).Scan(...)
}
```

### Close Rows!

```go
// ALWAYS close rows - prevents connection leaks
rows, err := db.QueryContext(ctx, query)
if err != nil {
    return err
}
defer rows.Close()  // CRITICAL!

for rows.Next() {
    // ...
}
return rows.Err()
```

---

## Interview Questions

### Beginner Level

1. **Q:** What is `sql.DB`?
   **A:** It's a connection pool, not a single connection. It manages multiple connections and is safe for concurrent use.

2. **Q:** Why must you close `rows` from a Query?
   **A:** Failure to close returns the connection to the pool, causing connection leaks and eventually exhausting the pool.

3. **Q:** How do you handle sql.ErrNoRows?
   **A:** Check with `if err == sql.ErrNoRows` or `errors.Is(err, sql.ErrNoRows)` to distinguish "not found" from actual errors.

### Intermediate Level

4. **Q:** How would you configure a connection pool for a high-traffic application?
   **A:** Set MaxOpenConns based on database capacity, MaxIdleConns for warm connections, ConnMaxLifetime for failover, and monitor with db.Stats().

5. **Q:** What's the difference between sql.DB and pgx?
   **A:** pgx is PostgreSQL-specific with binary protocol, better performance, COPY support, Listen/Notify, and richer type support.

6. **Q:** How do you ensure a transaction rollback on error?
   **A:** Defer rollback immediately after BeginTx. Rollback is a no-op after commit.

### Advanced Level

7. **Q:** How would you efficiently insert 1 million rows?
   **A:** Use PostgreSQL COPY protocol (pgx.CopyFrom). It's 100x faster than INSERT.

8. **Q:** SQLC vs ORM - when to use each?
   **A:** SQLC for type-safety with SQL control, compile-time checking. ORM for rapid prototyping but avoid in performance-critical paths.

9. **Q:** How do you handle database migrations in production?
   **A:** Use tools like golang-migrate, version-controlled SQL files, run in CI/CD, test rollbacks, use transactions where supported.

---

## Summary

| Topic | Key Points |
|-------|------------|
| database/sql | Generic interface, sql.DB is pool, drivers register via import |
| Connection Pool | Configure MaxOpenConns, MaxIdleConns, ConnMaxLifetime |
| Querying | Query (many), QueryRow (one), Exec (no return), always close rows |
| pgx | PostgreSQL-specific, binary protocol, COPY, Listen/Notify |
| Transactions | BeginTx, defer Rollback, Commit, isolation levels |
| Query Building | Raw SQL > SQLC > Query Builders > ORMs |
| Migrations | golang-migrate, versioned SQL files, up/down migrations |
| Best Practices | Repository pattern, context always, domain errors |

**Next Phase:** [Phase 10 — Cloud Native & Production](../Phase_10/Phase_10.md)


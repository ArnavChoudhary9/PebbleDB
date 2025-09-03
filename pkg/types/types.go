package types

// ContextKey represents a context key type
type ContextKey string

const (
	// UserContextKey is used to store user information in context
	UserContextKey ContextKey = "user"
	// DatabaseContextKey is used to store database connection in context
	DatabaseContextKey ContextKey = "database"
	// WorkingDirectoryContextKey is used to store working directory in context
	WorkingDirectoryContextKey ContextKey = "working_directory"
)

// Project represents a database project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	Path        string `json:"path,omitempty"`
}

// JSONJoin represents a join operation in JSON
type JSONJoin struct {
	Type      string `json:"type"`      // "INNER", "LEFT", "RIGHT", "FULL"
	Table     string `json:"table"`     // Table to join
	Condition string `json:"condition"` // Join condition (e.g., "users.id = profiles.user_id")
}

// JSONRequest represents a generic JSON request
type JSONRequest struct {
	Action    string                 `json:"action"`
	ProjectID string                 `json:"project_id,omitempty"` // Project identifier
	Table     string                 `json:"table"`
	Tables    []string               `json:"tables,omitempty"`    // For join action
	On        string                 `json:"on,omitempty"`        // For join condition
	JoinType  string                 `json:"join_type,omitempty"` // Optional join type
	Data      map[string]interface{} `json:"data,omitempty"`
	Where     string                 `json:"where,omitempty"`
	WhereArgs []interface{}          `json:"where_args,omitempty"`
	Columns   []string               `json:"columns,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
	Offset    int                    `json:"offset,omitempty"`
	OrderBy   string                 `json:"order_by,omitempty"`
	GroupBy   string                 `json:"group_by,omitempty"`
	Having    string                 `json:"having,omitempty"`
	Schema    map[string]interface{} `json:"schema,omitempty"`
	Joins     []JSONJoin             `json:"joins,omitempty"`
	// Project-specific fields
	ProjectName        string `json:"project_name,omitempty"`
	ProjectDescription string `json:"project_description,omitempty"`
}

// JSONResponse represents a generic JSON response
type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int64       `json:"count,omitempty"`
	ID      int64       `json:"id,omitempty"`
	Query   string      `json:"query,omitempty"` // Optional: show generated query for debugging
}

// RefreshTokenRequest represents the refresh token request payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse represents the refresh token response
type RefreshTokenResponse struct {
	AccessToken  string      `json:"access_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int         `json:"expires_in"`
	ExpiresAt    int64       `json:"expires_at"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
}

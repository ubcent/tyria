package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

var (
	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden is returned when user doesn't have required permissions
	ErrForbidden = errors.New("forbidden")
	// ErrInvalidRole is returned when role is not recognized
	ErrInvalidRole = errors.New("invalid role")
)

// Role represents user roles in the system
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleViewer Role = "viewer"
)

// ValidRoles returns all valid roles
func ValidRoles() []Role {
	return []Role{RoleOwner, RoleAdmin, RoleViewer}
}

// IsValidRole checks if a role string is valid
func IsValidRole(role string) bool {
	for _, r := range ValidRoles() {
		if string(r) == role {
			return true
		}
	}
	return false
}

// CanAccess checks if a role can access a resource based on RBAC
func (r Role) CanAccess(requiredRole Role) bool {
	switch r {
	case RoleOwner:
		return true // Owner can access everything
	case RoleAdmin:
		return requiredRole == RoleAdmin || requiredRole == RoleViewer
	case RoleViewer:
		return requiredRole == RoleViewer
	default:
		return false
	}
}

// UserContext represents the authenticated user context
type UserContext struct {
	UserID   int    `json:"user_id"`
	TenantID int    `json:"tenant_id"`
	Email    string `json:"email"`
	Role     Role   `json:"role"`
}

// ContextKey is used for context values
type ContextKey string

const (
	UserContextKey ContextKey = "user_context"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtManager *JWTManager
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtManager *JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth middleware that requires valid JWT authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header or cookie
		token := m.extractToken(r)
		if token == "" {
			http.Error(w, "Missing authentication token", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
			return
		}

		// Create user context
		userCtx := &UserContext{
			UserID:   claims.UserID,
			TenantID: claims.TenantID,
			Email:    claims.Email,
			Role:     Role(claims.Role),
		}

		// Add to request context
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware that requires a specific role or higher
func (m *AuthMiddleware) RequireRole(requiredRole Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := GetUserContext(r.Context())
			if userCtx == nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !userCtx.Role.CanAccess(requiredRole) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractToken extracts JWT token from request
func (m *AuthMiddleware) extractToken(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Try cookie
	cookie, err := r.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// GetUserContext retrieves user context from request context
func GetUserContext(ctx context.Context) *UserContext {
	userCtx, ok := ctx.Value(UserContextKey).(*UserContext)
	if !ok {
		return nil
	}
	return userCtx
}

// SetUserIDHeader sets X-User-ID header for downstream services
func SetUserIDHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userCtx := GetUserContext(r.Context()); userCtx != nil {
			r.Header.Set("X-User-ID", strconv.Itoa(userCtx.UserID))
			r.Header.Set("X-Tenant-ID", strconv.Itoa(userCtx.TenantID))
			r.Header.Set("X-User-Role", string(userCtx.Role))
		}
		next.ServeHTTP(w, r)
	})
}
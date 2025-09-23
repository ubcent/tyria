package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/ubcent/edge.link/internal/auth"
	"github.com/ubcent/edge.link/internal/models"
	"github.com/ubcent/edge.link/internal/tenant"
	"github.com/ubcent/edge.link/internal/users"
)

// AuthServer handles authentication endpoints
type AuthServer struct {
	db         *sql.DB
	userSvc    *users.Service
	tenantSvc  *tenant.Service
	jwtManager *auth.JWTManager
}

// NewAuthServer creates a new authentication server
func NewAuthServer(db *sql.DB) *AuthServer {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production" // Default for development
	}

	return &AuthServer{
		db:         db,
		userSvc:    users.NewService(db),
		tenantSvc:  tenant.NewService(db),
		jwtManager: auth.NewJWTManager(jwtSecret, "edge.link"),
	}
}

// SignupRequest represents the signup request payload
type SignupRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	CompanyName string `json:"company_name"`
}

// SignupResponse represents the signup response
type SignupResponse struct {
	Message string       `json:"message"`
	User    *models.User `json:"user,omitempty"`
	Token   string       `json:"token,omitempty"`
}

// SigninRequest represents the signin request payload
type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SigninResponse represents the signin response
type SigninResponse struct {
	Message string       `json:"message"`
	User    *models.User `json:"user,omitempty"`
	Token   string       `json:"token"`
}

// writeJSONError writes a JSON formatted error response
func (a *AuthServer) writeJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorResp := struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	}
	json.NewEncoder(w).Encode(errorResp)
}

// HandleSignup handles user registration
func (a *AuthServer) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.CompanyName == "" {
		a.writeJSONError(w, "Email, password, and company name are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	_, err := a.userSvc.GetByEmail(r.Context(), req.Email)
	if err == nil {
		a.writeJSONError(w, "User already exists", http.StatusConflict)
		return
	}

	// Create tenant first
	newTenant := &models.Tenant{
		Name:   req.CompanyName,
		Plan:   "free",
		Status: "active",
	}

	if err := a.tenantSvc.Create(r.Context(), newTenant); err != nil {
		a.writeJSONError(w, "Failed to create tenant", http.StatusInternalServerError)
		return
	}

	// Create user as owner of the tenant
	newUser := &models.User{
		TenantID: newTenant.ID,
		Email:    req.Email,
		Role:     string(auth.RoleOwner),
	}

	if err := a.userSvc.Create(r.Context(), newUser, req.Password); err != nil {
		a.writeJSONError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := a.jwtManager.GenerateToken(newUser.ID, newUser.TenantID, newUser.Email, newUser.Role)
	if err != nil {
		a.writeJSONError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set secure HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		Secure:   true, // Set to false for development over HTTP
		SameSite: http.SameSiteStrictMode,
	})

	// Hide password in response
	newUser.HashedPassword = ""

	response := SignupResponse{
		Message: "User created successfully. Please check your email for verification.",
		User:    newUser,
		Token:   token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// HandleSignin handles user authentication
func (a *AuthServer) HandleSignin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		a.writeJSONError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := a.userSvc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		a.writeJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := a.jwtManager.GenerateToken(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		a.writeJSONError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set secure HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		Secure:   true, // Set to false for development over HTTP
		SameSite: http.SameSiteStrictMode,
	})

	// Hide password in response
	user.HashedPassword = ""

	response := SigninResponse{
		Message: "Authentication successful",
		User:    user,
		Token:   token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSignout handles user logout
func (a *AuthServer) HandleSignout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear the auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Expire immediately
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// HandleProfile returns the current user's profile
func (a *AuthServer) HandleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userCtx := auth.GetUserContext(r.Context())
	if userCtx == nil {
		a.writeJSONError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	user, err := a.userSvc.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		a.writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	// Hide password
	user.HashedPassword = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// SetupAuthRoutes sets up authentication routes
func (a *AuthServer) SetupAuthRoutes(router *http.ServeMux) {
	router.HandleFunc("/api/auth/signup", a.HandleSignup)
	router.HandleFunc("/api/auth/signin", a.HandleSignin)
	router.HandleFunc("/api/auth/signout", a.HandleSignout)

	// Protected routes
	authMiddleware := auth.NewAuthMiddleware(a.jwtManager)
	router.Handle("/api/auth/profile", authMiddleware.RequireAuth(http.HandlerFunc(a.HandleProfile)))
}

// GetAuthMiddleware returns the auth middleware for use in other routes
func (a *AuthServer) GetAuthMiddleware() *auth.AuthMiddleware {
	return auth.NewAuthMiddleware(a.jwtManager)
}

// ConfirmEmailRequest represents email confirmation request
type ConfirmEmailRequest struct {
	Token string `json:"token"`
}

// HandleConfirmEmail handles email confirmation (stub implementation)
func (a *AuthServer) HandleConfirmEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConfirmEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual email verification logic
	// For now, just return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email confirmed successfully",
	})
}

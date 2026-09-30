package services

// Shared error values used by every feature (features/<menu>).
// Handlers map these to HTTP statuses via errors.Is; keep them here
// so cross-feature checks stay comparable (no duplicates).
import "errors"

var (
	ErrInputRequired    = errors.New("username, email, and password are required")
	ErrUsernameTooShort = errors.New("username must be at least 3 characters")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must not exceed 72 bytes")
	ErrUsernameTaken    = errors.New("username already taken")
	ErrEmailTaken       = errors.New("email already registered")
	ErrInvalidLogin     = errors.New("invalid username or password")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidRefresh   = errors.New("invalid or expired refresh token")
	ErrForbidden        = errors.New("forbidden")

	// Role domain.
	ErrRoleNotFound  = errors.New("role not found")
	ErrRoleExists    = errors.New("role code already exists")
	ErrRoleProtected = errors.New("system roles cannot be deleted")
	ErrRoleInUse     = errors.New("role is still assigned to users")
	ErrInvalidRole   = errors.New("invalid role")

	// Notification template domain.
	ErrTemplateNotFound = errors.New("template not found")
	ErrInvalidTemplate  = errors.New("invalid template")
	ErrInvalidChannel   = errors.New("channel must be EMAIL, PUSH, or INAPP")

	// Session domain.
	ErrSessionNotFound = errors.New("session not found")

	// Menu registry domain (CPMENU + CPMATRIX).
	ErrMenuNotFound     = errors.New("menu not found")
	ErrMenuInUse        = errors.New("menu is still assigned to roles")
	ErrInvalidMenu      = errors.New("invalid menu")
	ErrPermissionExists = errors.New("permission code already exists")
	ErrMenuHasChildren  = errors.New("menu still has child menus")

	// Module domain (CPMATRIX tabel).
	ErrModuleNotFound = errors.New("module not found")
	ErrModuleExists   = errors.New("module already exists")
	ErrModuleInUse    = errors.New("module still has menus")
	ErrInvalidModule  = errors.New("invalid module")

	// Dashboard domain (features/organizer, features/seller).
	ErrInvalidFilter = errors.New("invalid dashboard filter")

	// Discovery domain (features/discovery, public read).
	ErrEventNotFound = errors.New("event not found")
	ErrNewsNotFound  = errors.New("news not found")

	// Event domain (features/event, owner-scoped writes).
	ErrEventState    = errors.New("invalid event status transition")
	ErrEventNotReady = errors.New("event is not ready to publish")

	// Ticketing domain (features/ticketing, buyer/officer flows).
	ErrOrderNotFound   = errors.New("order not found")
	ErrTicketNotFound  = errors.New("ticket not found")
	ErrInvalidCheckout = errors.New("invalid checkout")
	ErrNotEligible     = errors.New("not eligible for refund")
	ErrScanDenied      = errors.New("scan not allowed")
)

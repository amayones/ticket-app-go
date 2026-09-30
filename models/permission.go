package models

import "time"

// Permission codes (seed di scripts/migrate2_rbac.sql). Satu permission
// per menu: jika role punya permission ini (baris CPPERMISSION), role dapat
// memakai seluruh fungsi menu tersebut. Tambah menu = tambah baris CPMENU
// (via POST /api/admin/menus). Role tanpa akses apa pun mendapat halaman kosong.
const (
	MenuUsers         = "MENU_USERS"
	MenuRoles         = "MENU_ROLES"
	MenuSessions      = "MENU_SESSIONS"
	MenuAudit         = "MENU_AUDIT"
	MenuSecurity      = "MENU_SECURITY"
	MenuSyslog        = "MENU_SYSLOG"
	MenuNotifications = "MENU_NOTIFICATIONS"
	// MenuModul hanya mengatur tampilnya halaman Modul & Menu di sidebar;
	// API manajemen menu tetap di bawah MENU_ROLES.
	MenuModul = "MENU_MODUL"
	// Dashboard menus (modul DASHBOARD, seed di migrate5_dashboard.sql).
	// Registered with ZERO grants; admin assigns via the Role & Permission
	// matrix and the API enforces the same grant (RequirePermission).
	MenuOrganizer = "MENU_ORGANIZER"
	MenuSeller    = "MENU_SELLER"
	// Discovery menus (modul DISCOVERY, seed di migrate7 + migrate8).
	// Public APIs need no auth; sidebar visibility follows matrix grants.
	// (Event detail is a modal inside Events, not a menu.)
	MenuHome       = "MENU_HOME"
	MenuNews       = "MENU_NEWS"
	MenuEvents     = "MENU_EVENTS"
	MenuPromotions = "MENU_PROMOTIONS"
	// Event menus (modul EVENT, seed di migrate10 + migrate11).
	// Registered with ZERO grants; admin assigns via the matrix.
	// APIs are owner-scoped (organizer only touches own events).
	// (Create Event is the FRM modal of My Events, not a menu.)
	MenuMyEvents    = "MENU_MY_EVENTS"
	MenuTicketTypes = "MENU_TICKET_TYPES"
	MenuAttendees   = "MENU_ATTENDEES"
	// Ticketing menus (modul TICKETING, seed di migrate13_ticketing_menus.sql).
	// Registered with ZERO grants; admin assigns via the matrix.
	// APIs are buyer/officer-scoped (ownership/assignment, not menu rights).
	// (Ticket Detail is a modal inside My Tickets, not a menu.)
	MenuCheckout    = "MENU_CHECKOUT"
	MenuMyTickets   = "MENU_MY_TICKETS"
	MenuRefunds     = "MENU_REFUNDS"
	MenuScanTicket  = "MENU_SCAN_TICKET"
	MenuScanHistory = "MENU_SCAN_HISTORY"
)

// Permission adalah permission tampil satu menu (kode MENU_* dari CPMENU);
// grant per role tercatat di tabel CPPERMISSION. Kind/Parent dibaca dari
// CPMENU supaya UI matriks tahu baris mana yang hanya header PARENT: baris
// PARENT tidak punya centang akses (grant-nya menyusul otomatis dari menu
// anak yang dicentang, lihat MyMenusByRole).
type Permission struct {
	ID          int       `json:"-"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Kind        string    `json:"kind"`
	Parent      string    `json:"parent_code,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// RoleDetail adalah role + daftar kode permission miliknya (untuk matriks RBAC).
type RoleDetail struct {
	Role
	Permissions []string `json:"permissions"`
}

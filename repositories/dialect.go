package repositories

import (
	"fmt"
	"strings"
)

// Dialect memilih sintaks SQL per engine. Semua query ditulis dengan
// placeholder "?" lalu di-rebind saat eksekusi (Bind).
type Dialect string

const (
	DialectMSSQL    Dialect = "mssql"
	DialectPostgres Dialect = "postgres"
	DialectSQLite   Dialect = "sqlite"
)

// ParseDialect memetakan DB_CONNECTION config ke Dialect.
func ParseDialect(raw string) Dialect {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "postgres", "postgresql", "pg":
		return DialectPostgres
	case "sqlite", "sqlite3":
		return DialectSQLite
	default:
		return DialectMSSQL
	}
}

// Bind menulis ulang "?" menjadi @pN (mssql) atau $N (postgres).
// sqlite memakai "?" apa adanya. CATATAN: jangan taruh "?" di dalam
// string literal query karena ikut dihitung.
func (d Dialect) Bind(query string) string {
	if d != DialectMSSQL && d != DialectPostgres {
		return query
	}
	var sb strings.Builder
	n := 0
	for _, r := range query {
		if r != '?' {
			sb.WriteRune(r)
			continue
		}
		n++
		if d == DialectPostgres {
			fmt.Fprintf(&sb, "$%d", n)
		} else {
			fmt.Fprintf(&sb, "@p%d", n)
		}
	}
	return sb.String()
}

// Now returns the current-timestamp function per engine.
func (d Dialect) Now() string {
	switch d {
	case DialectPostgres:
		return "NOW()"
	case DialectSQLite:
		return "CURRENT_TIMESTAMP"
	default:
		return "GETDATE()"
	}
}

// IsTrue returns the boolean-true literal per engine.
func (d Dialect) IsTrue() string {
	if d == DialectPostgres {
		return "TRUE"
	}
	return "1"
}

// pageMSSQL and pageStd render pagination. MSSQL takes (offset, limit),
// postgres/sqlite take (limit, offset) — argument order differs!
func pageMSSQL() string { return "OFFSET ? ROWS FETCH NEXT ? ROWS ONLY" }
func pageStd() string   { return "LIMIT ? OFFSET ?" }

// Table qualifies the table name. MSSQL keeps dbo. (login default schema
// may vary); postgres/sqlite use the bare name (dbo schema does not exist
// there, and sqlite rejects unknown schema prefixes).
func (d Dialect) Table(name string) string {
	if d == DialectMSSQL {
		return "dbo." + name
	}
	return name
}

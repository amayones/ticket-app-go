package repositories

import "testing"

func TestParseDialect(t *testing.T) {
	cases := []struct {
		in   string
		want Dialect
	}{
		{"sqlserver", DialectMSSQL},
		{"mssql", DialectMSSQL},
		{"", DialectMSSQL},
		{"ngawur", DialectMSSQL},
		{"postgres", DialectPostgres},
		{"postgresql", DialectPostgres},
		{"sqlite", DialectSQLite},
		{"sqlite3", DialectSQLite},
	}
	for _, tc := range cases {
		if got := ParseDialect(tc.in); got != tc.want {
			t.Fatalf("ParseDialect(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDialectBind(t *testing.T) {
	q := "SELECT * FROM t WHERE a = ? AND b = ?"
	if got := DialectMSSQL.Bind(q); got != "SELECT * FROM t WHERE a = @p1 AND b = @p2" {
		t.Fatalf("mssql bind: %q", got)
	}
	if got := DialectPostgres.Bind(q); got != "SELECT * FROM t WHERE a = $1 AND b = $2" {
		t.Fatalf("pg bind: %q", got)
	}
	if got := DialectSQLite.Bind(q); got != q {
		t.Fatalf("sqlite bind must be identity: %q", got)
	}
}

func TestDialectHelpers(t *testing.T) {
	if DialectMSSQL.Now() != "GETDATE()" || DialectPostgres.Now() != "NOW()" ||
		DialectSQLite.Now() != "CURRENT_TIMESTAMP" {
		t.Fatal("Now() mismatch")
	}
	if DialectMSSQL.Table("CPUSER") != "dbo.CPUSER" {
		t.Fatal("mssql table must keep dbo.")
	}
	if DialectPostgres.Table("CPUSER") != "CPUSER" || DialectSQLite.Table("CPUSER") != "CPUSER" {
		t.Fatal("pg/sqlite table must be bare")
	}
	if DialectPostgres.IsTrue() != "TRUE" || DialectMSSQL.IsTrue() != "1" || DialectSQLite.IsTrue() != "1" {
		t.Fatal("IsTrue() mismatch")
	}
}

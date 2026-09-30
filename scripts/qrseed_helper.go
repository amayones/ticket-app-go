// QR seed helper: prints UPDATE statements filling T_TICKET.QR_TOKEN with
// signed tokens (utils.GenerateTicketToken). SQL cannot sign JWTs, so seed
// rows are inserted with NULL tokens and backfilled by this helper.
//
// Usage (from repo root):
//
//	go run ./scripts/qrseed_helper.go > /tmp/qrseed.sql
//	sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB> -C -b -i /tmp/qrseed.sql
//
// Reads QR_SECRET (fallback JWT_SECRET) from environment/.env like config.
package main

import (
	"fmt"
	"os"

	"golang-backend/config"
	"golang-backend/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}
	db, err := config.ConnectDatabase(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "database:", err)
		os.Exit(1)
	}
	defer db.Close()

	dialect := "dbo."
	if cfg.DBConnection == config.DBPostgres || cfg.DBConnection == config.DBSQLite {
		dialect = ""
	}
	query := `SELECT t.CODE, o.EVENT_CODE FROM ` + dialect + `T_TICKET t
	  JOIN ` + dialect + `T_ORDER o ON o.CODE = t.ORDER_CODE
	  WHERE t.QR_TOKEN IS NULL`
	rows, err := db.Query(query)
	if err != nil {
		fmt.Fprintln(os.Stderr, "query:", err)
		os.Exit(1)
	}
	defer rows.Close()
	for rows.Next() {
		var tcode, event string
		if err := rows.Scan(&tcode, &event); err != nil {
			fmt.Fprintln(os.Stderr, "scan:", err)
			os.Exit(1)
		}
		token, err := utils.GenerateTicketToken(cfg.QRSecret, tcode, event)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sign:", err)
			os.Exit(1)
		}
		fmt.Printf("UPDATE %sT_TICKET SET QR_TOKEN = '%s' WHERE CODE = '%s';\n", dialect, token, tcode)
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "rows:", err)
		os.Exit(1)
	}
}

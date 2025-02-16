package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/sifaconer/grud/src/core"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	log.Info("Hello World!")

	// get params to dsn from env or string flag
	dsn := flag.String("dsn", "", "Postgres DSN")
	flag.Parse()

	if *dsn == "" {
		log.Error("DSN not provided")
		os.Exit(1)
	}

	database := core.NewPostgresDatabase(core.DatabaseConfig{DSN: *dsn}, log)
	database, err := database.Connect()
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	schemas, err := database.GetSchemas()
	if err != nil {
		log.Error("Failed to get schemas", "error", err)
		os.Exit(1)
	}

	tables, err := database.GetTables()
	if err != nil {
		log.Error("Failed to get tables", "error", err)
		os.Exit(1)
	}

	for _, schema := range schemas {
		log.Info("Schema", "schema", schema.Name)
		for _, table := range tables {
			if table.Schema == schema.Name {
				log.Info("🗃️  Table", "table", table.Name)
				for _, column := range table.Columns {
					log.Info("  📚   Column",
						"column", column.ColumnName,
						"data_type", column.GetUdtName(),
						"nullable", column.GetIsNullable())
					
				}
				for _, primaryKey := range table.PrimaryKeys {
					log.Info("    🔑    Primary Key", "primary_key", primaryKey.Constraint.ColumnName)
						
				}
				for _, foreignKey := range table.ForeignKeys {
					log.Info("    🗝️    Foreign Key", "foreign_key", foreignKey.Constraint.ColumnName)
				}
			}
		}
	}
}

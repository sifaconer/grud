package core

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type postgresDatabase struct {
	config DatabaseConfig
	db     *bun.DB
	log    *slog.Logger
}

// Close implements Database.
func (p *postgresDatabase) Close() error {
	return p.db.Close()
}

// Connect implements Database.
func (p *postgresDatabase) Connect() (Database, error) {
	dsn := p.dsn()
	p.log.Info("Connecting to", "dsn", p.obfuscateDSN())

	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	bunDB := bun.NewDB(db, pgdialect.New())
	err := bunDB.Ping()
	if err != nil {
		p.log.Error("Failed to ping database", "error", err)
		return nil, err
	}
	p.db = bunDB
	return p, nil
}

// GetTableColumns implements Database.
func (p *postgresDatabase) GetTableColumns(tableName string) (Columns, error) {
	result := Columns{}

	query := `
		SELECT 
			table_catalog,
			table_schema,
			table_name,
			column_name,
			ordinal_position,
			column_default,
			is_nullable,
			character_maximum_length,
			character_octet_length,
			numeric_precision,
			udt_name,
			dtd_identifier
		FROM 
			information_schema.columns
		WHERE 
			table_name = ?
			AND table_schema = ?;
	`
	schema := "public"
	if s := strings.Split(tableName, "."); len(s) > 1 {
		schema = strings.Split(tableName, ".")[0]
		tableName = strings.Split(tableName, ".")[1]
	}
	rows, err := p.db.QueryContext(context.Background(), query, tableName, schema)
	if err != nil {
		p.log.Error("Failed to get table columns", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var column Column
		err := rows.Scan(
			&column.TableCatalog,
			&column.TableSchema,
			&column.TableName,
			&column.ColumnName,
			&column.OrdinalPosition,
			&column.ColumnDefault,
			&column.IsNullable,
			&column.CharacterMaximumLength,
			&column.CharacterOctetLength,
			&column.NumericPrecision,
			&column.UdtName,
			&column.DtdIdentifier)
		if err != nil {
			p.log.Error("Failed to scan column", "error", err)
			return nil, err
		}
		result = append(result, column)
	}

	return result, nil
}

// GetTableForeignKeys implements Database.
func (p *postgresDatabase) GetTableForeignKeys(tableName string) (ForeignKeys, error) {
	result := ForeignKeys{}

	query := `
		SELECT 
			tc.table_name,
			tc.table_schema,
			tc.constraint_type, 
			tc.constraint_name,
			tc.constraint_catalog,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name,
			ccu.table_schema  AS foreign_table_schema
		FROM 
			information_schema.key_column_usage AS kcu
		JOIN 
			information_schema.table_constraints AS tc 
			ON kcu.constraint_name = tc.constraint_name
			AND kcu.table_schema = tc.table_schema
		JOIN 
			information_schema.constraint_column_usage AS ccu 
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE 
			tc.constraint_type = 'FOREIGN KEY'
			AND kcu.table_name = ?
			AND kcu.table_schema = ?;
	`
	schema := "public"
	if s := strings.Split(tableName, "."); len(s) > 1 {
		schema = strings.Split(tableName, ".")[0]
		tableName = strings.Split(tableName, ".")[1]
	}
	rows, err := p.db.QueryContext(context.Background(), query, tableName, schema)
	if err != nil {
		p.log.Error("Failed to get table foreign keys", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var foreignKey ForeignKey
		err := rows.Scan(
			&foreignKey.Constraint.TableName,
			&foreignKey.Constraint.TableSchema,
			&foreignKey.Constraint.ConstraintType,
			&foreignKey.Constraint.ConstraintName,
			&foreignKey.Constraint.Catalog,
			&foreignKey.Constraint.ColumnName,
			&foreignKey.ForeignTableName,
			&foreignKey.ForeignColumnName,
			&foreignKey.ForeignSchemaName)
		if err != nil {
			p.log.Error("Failed to scan foreign key", "error", err)
			return nil, err
		}
		result = append(result, foreignKey)
	}

	return result, nil
}

// GetSchemas implements Database.
func (p *postgresDatabase) GetSchemas() (Schemas, error) {
	result := Schemas{}

	// Filter schemas that start with pg_ or information_schema
	query := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT LIKE 'pg_%'
		AND schema_name NOT LIKE 'information_schema'
	`
	rows, err := p.db.QueryContext(context.Background(), query)
	if err != nil {
		p.log.Error("Failed to get schemas", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var schema Schema
		err := rows.Scan(&schema.Name)
		if err != nil {
			p.log.Error("Failed to scan schema", "error", err)
			return nil, err
		}
		result = append(result, schema)
	}
	return result, nil
}

// GetTables implements Database.
func (p *postgresDatabase) GetTables() (Tables, error) {
	result := Tables{}

	query := `
		SELECT 
			table_schema,
			table_name
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
	`
	rows, err := p.db.QueryContext(context.Background(), query)
	if err != nil {
		p.log.Error("Failed to get table names", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table Table
		err := rows.Scan(&table.Schema, &table.Name)
		if err != nil {
			p.log.Error("Failed to scan table", "error", err)
			return nil, err
		}
		err = p.populateTable(&table)
		if err != nil {
			p.log.Error("Failed to populate table", "error", err)
			return nil, err
		}
		result = append(result, table)
	}

	return result, nil
}

// GetTableNamesBySchema implements Database.
func (p *postgresDatabase) GetTableNamesBySchema(schemaName string) (Tables, error) {
	result := Tables{}

	query := `
		SELECT 
			table_schema,
			table_name
		FROM information_schema.tables
		WHERE table_schema = ?
	`
	rows, err := p.db.QueryContext(context.Background(), query, schemaName)
	if err != nil {
		p.log.Error("Failed to get table names by schema", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table Table
		err := rows.Scan(&table.Schema, &table.Name)
		if err != nil {
			p.log.Error("Failed to scan table", "error", err)
			return nil, err
		}
		err = p.populateTable(&table)
		if err != nil {
			p.log.Error("Failed to populate table", "error", err)
			return nil, err
		}
		result = append(result, table)
	}

	return result, nil
}

// GetTablePrimaryKeys implements Database.
func (p *postgresDatabase) GetTablePrimaryKeys(tableName string) (PrimaryKeys, error) {
	result := PrimaryKeys{}

	query := `
		SELECT 
			tc.table_name,
			tc.table_schema,
			tc.constraint_type, 
			tc.constraint_name,
			tc.constraint_catalog,
			kcu.column_name
		FROM 
			information_schema.key_column_usage AS kcu
		JOIN 
			information_schema.table_constraints AS tc 
			ON kcu.constraint_name = tc.constraint_name
			AND kcu.table_schema = tc.table_schema
		JOIN 
			information_schema.constraint_column_usage AS ccu 
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE 
			tc.constraint_type = 'PRIMARY KEY'
			AND kcu.table_name = ?
			AND kcu.table_schema = ?;
	`
	schema := "public"
	if s := strings.Split(tableName, "."); len(s) > 1 {
		schema = strings.Split(tableName, ".")[0]
		tableName = strings.Split(tableName, ".")[1]
	}
	rows, err := p.db.QueryContext(context.Background(), query, tableName, schema)
	if err != nil {
		p.log.Error("Failed to get table primary keys", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var primaryKey PrimaryKey
		err := rows.Scan(
			&primaryKey.Constraint.TableName,
			&primaryKey.Constraint.TableSchema,
			&primaryKey.Constraint.ConstraintType,
			&primaryKey.Constraint.ConstraintName,
			&primaryKey.Constraint.Catalog,
			&primaryKey.Constraint.ColumnName)
		if err != nil {
			p.log.Error("Failed to scan primary key", "error", err)
			return nil, err
		}
		result = append(result, primaryKey)
	}

	return result, nil
}

func (p *postgresDatabase) dsn() string {
	if p.config.DSN != "" {
		return p.config.DSN
	}
	// DSN format: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS
	// Example: postgres://user:password@host:port/database?sslmode=disable&connect_timeout=10&application_name=grud
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=10&application_name=grud",
		p.config.User, p.config.Password, p.config.Host, p.config.Port, p.config.Database, p.config.SSLMode)
}

func (p *postgresDatabase) obfuscateDSN() string {
	dsn := p.dsn()
	parsedURL, err := url.Parse(dsn)
	if err != nil {
		p.log.Error("Failed to parse DSN", "error", err)
		return dsn
	}

	return parsedURL.Redacted()
}

func (p *postgresDatabase) populateTable(table *Table) error {
	cols, err := p.GetTableColumns(fmt.Sprintf("%s.%s", table.Schema, table.Name))
	if err != nil {
		p.log.Error("Failed to get table columns", "error", err)
		return err
	}
	table.Columns = cols
	primaryKeys, err := p.GetTablePrimaryKeys(fmt.Sprintf("%s.%s", table.Schema, table.Name))
	if err != nil {
		p.log.Error("Failed to get table primary keys", "error", err)
		return err
	}
	table.PrimaryKeys = primaryKeys
	foreignKeys, err := p.GetTableForeignKeys(fmt.Sprintf("%s.%s", table.Schema, table.Name))
	if err != nil {
		p.log.Error("Failed to get table foreign keys", "error", err)
		return err
	}
	table.ForeignKeys = foreignKeys

	return nil
}

func NewPostgresDatabase(config DatabaseConfig, log *slog.Logger) Database {
	return &postgresDatabase{
		config: config,
		log:    log,
	}
}

var _ Database = (*postgresDatabase)(nil)

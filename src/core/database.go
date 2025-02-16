package core

type Database interface {
	// Connect establishes a connection to the database.
	Connect() (Database, error)

	// Close closes the connection to the database.
	Close() error

	// GetSchemas returns a list of schema names in the database.
	GetSchemas() (Schemas, error)

	// GetTables returns a list of table names in the database.
	GetTables() (Tables, error)

	// GetTableNamesBySchema returns a list of table names in the specified schema.
	GetTableNamesBySchema(schemaName string) (Tables, error)

	// GetTableColumns returns a list of columns in the specified table.
	GetTableColumns(tableName string) (Columns, error)

	// GetTablePrimaryKeys returns a list of primary keys in the specified table.
	GetTablePrimaryKeys(tableName string) (PrimaryKeys, error)

	// GetTableForeignKeys returns a list of foreign keys in the specified table.
	GetTableForeignKeys(tableName string) (ForeignKeys, error)
}

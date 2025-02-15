package core

type Database interface {
	// GetTableNames returns a list of table names in the database.
	GetTableNames() ([]string, error)

	// GetTableColumns returns a list of columns in the specified table.
	GetTableColumns(tableName string) ([]string, error)

	// GetTablePrimaryKeys returns a list of primary keys in the specified table.
	GetTablePrimaryKeys(tableName string) ([]string, error)

	// GetTableForeignKeys returns a list of foreign keys in the specified table.
	GetTableForeignKeys(tableName string) ([]string, error)
}
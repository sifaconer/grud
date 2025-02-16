package core

import "strconv"

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	DSN      string
}

type Schema struct {
	Name   string `json:"name"`
	Tables Tables `json:"tables"`
}

type Schemas []Schema

type Column struct {
	TableCatalog           string  `json:"table_catalog"`
	TableSchema            string  `json:"table_schema"`
	TableName              string  `json:"table_name"`
	ColumnName             string  `json:"column_name"`
	OrdinalPosition        string `json:"ordinal_position"`
	ColumnDefault          *string `json:"column_default"`
	IsNullable             *string `json:"is_nullable"`
	CharacterMaximumLength *string `json:"character_maximum_length"`
	CharacterOctetLength   *string `json:"character_octet_length"`
	NumericPrecision       *string `json:"numeric_precision"`
	UdtName                *string `json:"udt_name"`
	DtdIdentifier          *string `json:"dtd_identifier"`
}

func (c Column) GetColumnDefault() string {
	if c.ColumnDefault == nil {
		return ""
	}
	return *c.ColumnDefault
}
	
func (c Column) GetIsNullable() bool {
	return c.IsNullable != nil && *c.IsNullable == "YES"
}

func (c Column) GetCharacterMaximumLength() int {
	if c.CharacterMaximumLength == nil {
		return 0
	}
	value, err := strconv.Atoi(*c.CharacterMaximumLength)
	if err != nil {
		return 0
	}
	return value
}

func (c Column) GetCharacterOctetLength() int {
	if c.CharacterOctetLength == nil {
		return 0
	}
	value, err := strconv.Atoi(*c.CharacterOctetLength)
	if err != nil {
		return 0
	}
	return value
}

func (c Column) GetNumericPrecision() int {
	value := 0
	if c.NumericPrecision != nil {
		value, _ = strconv.Atoi(*c.NumericPrecision)
	}
	return value
}

func (c Column) GetUdtName() string {
	if c.UdtName == nil {
		return ""
	}
	return *c.UdtName
}

func (c Column) GetDtdIdentifier() string {
	if c.DtdIdentifier == nil {
		return ""
	}
	return *c.DtdIdentifier
}

type Columns []Column

type Constraint struct {
	ConstraintName string `json:"constraint_name"`
	ConstraintType string `json:"constraint_type"`
	ColumnName     string `json:"column_name"`
	TableName      string `json:"table_name"`
	TableSchema    string `json:"table_schema"`
	Catalog        string `json:"catalog"`
}

type PrimaryKey struct {
	Constraint
}

type PrimaryKeys []PrimaryKey

type ForeignKey struct {
	Constraint
	ForeignTableName  string `json:"foreign_table_name"`
	ForeignColumnName string `json:"foreign_column_name"`
	ForeignSchemaName string `json:"foreign_schema_name"`
}

type ForeignKeys []ForeignKey

type Table struct {
	Schema      string      `json:"schema"`
	Name        string      `json:"name"`
	Columns     Columns     `json:"columns"`
	PrimaryKeys PrimaryKeys `json:"primary_keys"`
	ForeignKeys ForeignKeys `json:"foreign_keys"`
}

type Tables []Table

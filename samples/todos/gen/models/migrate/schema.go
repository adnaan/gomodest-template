// Code generated (@generated) by entc, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// TodosColumns holds the columns for the "todos" table.
	TodosColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "text", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Nullable: true, Enums: []string{"todo", "inprogress", "done"}, Default: "todo"},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
	}
	// TodosTable holds the schema information for the "todos" table.
	TodosTable = &schema.Table{
		Name:        "todos",
		Columns:     TodosColumns,
		PrimaryKey:  []*schema.Column{TodosColumns[0]},
		ForeignKeys: []*schema.ForeignKey{},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		TodosTable,
	}
)

func init() {
	TodosTable.Annotation = &entsql.Annotation{
		Table: "todos",
	}
}

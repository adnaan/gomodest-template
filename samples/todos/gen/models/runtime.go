// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"gomodest-template/samples/todos/gen/models/todo"
	"gomodest-template/samples/todos/schema"
	"time"

	"github.com/google/uuid"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	todoFields := schema.Todo{}.Fields()
	_ = todoFields
	// todoDescCreatedAt is the schema descriptor for created_at field.
	todoDescCreatedAt := todoFields[3].Descriptor()
	// todo.DefaultCreatedAt holds the default value on creation for the created_at field.
	todo.DefaultCreatedAt = todoDescCreatedAt.Default.(func() time.Time)
	// todoDescUpdatedAt is the schema descriptor for updated_at field.
	todoDescUpdatedAt := todoFields[4].Descriptor()
	// todo.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	todo.DefaultUpdatedAt = todoDescUpdatedAt.Default.(func() time.Time)
	// todo.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	todo.UpdateDefaultUpdatedAt = todoDescUpdatedAt.UpdateDefault.(func() time.Time)
	// todoDescID is the schema descriptor for id field.
	todoDescID := todoFields[0].Descriptor()
	// todo.DefaultID holds the default value on creation for the id field.
	todo.DefaultID = todoDescID.Default.(func() uuid.UUID)
}

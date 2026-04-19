package schema

import (
	"time"

	columnx "github.com/DaiYuANg/arcgo/dbx/column"
	"github.com/DaiYuANg/arcgo/dbx/idgen"
	schemax "github.com/DaiYuANg/arcgo/dbx/schema"
)

type UserRow struct {
	ID        int64     `dbx:"id"`
	Name      string    `dbx:"name"`
	Email     string    `dbx:"email"`
	Age       int       `dbx:"age"`
	CreatedAt time.Time `dbx:"created_at,codec=rfc3339_time"`
	UpdatedAt time.Time `dbx:"updated_at,codec=rfc3339_time"`
}

type UserSchema struct {
	schemax.Schema[UserRow]
	ID        columnx.IDColumn[UserRow, int64, idgen.IDSnowflake] `dbx:"id"`
	Name      columnx.Column[UserRow, string]                     `dbx:"name"`
	Email     columnx.Column[UserRow, string]                     `dbx:"email,unique"`
	Age       columnx.Column[UserRow, int]                        `dbx:"age"`
	CreatedAt columnx.Column[UserRow, time.Time]                  `dbx:"created_at,codec=rfc3339_time"`
	UpdatedAt columnx.Column[UserRow, time.Time]                  `dbx:"updated_at,codec=rfc3339_time"`
}

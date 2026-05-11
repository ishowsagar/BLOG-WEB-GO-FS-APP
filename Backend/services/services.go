package services

import "database/sql"

// @ Methods for enabling models instances

// func that creates intance of type UserDbModel -> which stores all the methods for querying related to user <- need sql db for it
func NewUserDbModel(db *sql.DB) *UserDBModel{
	return &UserDBModel{
		db: db,
	}
}
package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

//  type that stores db of type *sql.DB
type TokenDBModel struct {
	DB *sql.DB
}

//  returns the instance of type TokenDbModel -> which stores db and also methods on it
func NewTokenDbModel(db *sql.DB) *TokenDBModel {
	return &TokenDBModel{
		DB: db,
	}
}

// method that belongs to the tokenDBModel which -> stores this method to insert token in the db
func(t *TokenDBModel) InsertToken(token models.Token) error {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		insert into 
			tokens(user_id,expiry,hash)
		values ($1,$2,$3)
		returning id,user_id,expiry,hash
	`

	resRow := t.DB.QueryRowContext(ctx,query,token.UserID,token.Expiry,token.Hash)
	
	var tokenToSend models.Token
	err := resRow.Scan(
		&tokenToSend.ID,
		&tokenToSend.UserID,
		&tokenToSend.Expiry,
		&tokenToSend.Hash,
	)

	if err != nil {
		return err
	}

	return nil

}


//  method that retrieve token from the database, must pass uint id
func(t *TokenDBModel) GetTokenByUserID(userID uint) (*models.Token,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		select 
			id,user_id,expiry,hash
		from
			tokens
		where 
			user_id=$1
	`

	resRow := t.DB.QueryRowContext(ctx,query,userID) // passed user ID
	
	var tokenToSend models.Token
	err := resRow.Scan(
		&tokenToSend.ID,
		&tokenToSend.UserID,
		&tokenToSend.Expiry,
		&tokenToSend.Hash,
	)

	if err != nil {
		//  if query was successfull but does not return any token
		if err == sql.ErrNoRows {
			return nil,sql.ErrNoRows
		}
		return nil,err
	}

	return &tokenToSend,nil
}



func(t *TokenDBModel) UpdateTokenIfExists(expiry time.Time,userID uint,hash string) error {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()


	query := `
		update
			tokens
		set
			expiry=$1,hash=$2
		where
			user_id=$3
	` 	

	resRow,err :=t.DB.ExecContext(ctx,query,expiry,hash,userID)
	if err!= nil {
		return err
	}

	// no of rows effected by it
	rows,err := resRow.RowsAffected()
	if err != nil {
		return err
	}

	// if no updated
	if rows == 0 {
		return sql.ErrNoRows
	}
	
	return nil

}
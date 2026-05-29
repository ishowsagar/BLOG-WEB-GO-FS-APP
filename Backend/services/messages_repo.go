package services

import (
	"context"
	"database/sql"

	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// types
type MessagesDBModel struct {
	DB *sql.DB
}

// func that returns the instance of type -> messagesDbModel which -> stores db pf type *sql.DB
func NewMessagesDBModel(db *sql.DB) *MessagesDBModel {
	return &MessagesDBModel{
		DB: db,
	}
}

// pointer in reciever and args cuz we want actual val, not val copy
func(m *MessagesDBModel) StoreDmMessage(msgPayload *models.DirectMessage) (error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		Insert into
			direct_messages(sender_id,reciever_id,message)
		values
			($1,$2,$3)
	`	

	res,err := m.DB.ExecContext(ctx,query,msgPayload.SenderID,msgPayload.RecieverID,msgPayload.Message)
	if err!= nil {
		return err
	}
	insertedRows,err := res.RowsAffected();
	if  err != nil {
		return err
	}
	if insertedRows == 0 {
		return sql.ErrNoRows	
	}

	return nil
}
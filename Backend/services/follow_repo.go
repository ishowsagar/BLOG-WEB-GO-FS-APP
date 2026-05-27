package services

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/utils"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrFollowAlreadyExists = errors.New("follow relationship already exists")

// @types
type FollowDBModel struct {
	DB *sql.DB
}

//  func that returns the instance of type FollowDbModel which -> stores db of type *sql.DB
func NewFollowDbModel(db *sql.DB) *FollowDBModel {
	return &FollowDBModel{
		DB: db,
	}
}


// method that belongs to the type FollowDBModel & add data into follows table who is following who
func(f *FollowDBModel) FollowUser(followerID,followingID uint) (bool,error) {
	// todo - add query approach in go chan~routine way
	// todo - need concurrent approach to follow user
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		Insert into
			follows(follower_id,followee_id)
		values
			($1,$2)
	`

	res,err := f.DB.ExecContext(ctx,query)
	if err!= nil {
		return false,err
	}
	rowNum,err := res.RowsAffected() ; 
	if err != nil {
		return false,err
	}

	//  checking if query was a successfull job but was there any actual insertion
	if rowNum == 0 {
		return false,sql.ErrNoRows
	}
	return true,nil

}

// ✅ Added named return parameter 'err' so the defer block can safely catch errors and rollback
func (f *FollowDBModel) FollowUserTransaction(followerID, followingID uint) (entryID uint,err error) {

	ctx, timeout := context.WithTimeout(context.Background(), utils.DbTimeoutDuration)
	defer timeout()

	// Begin transaction
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0,err
	}

	// ✅ This now works perfectly because 'err' is a named return value!
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var followEntryID uint

	followingQuery := `
		INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
		RETURNING id
	`
	// ✅ Switched to QueryRowContext safely without touching uninitialized 'res'
	err = tx.QueryRowContext(ctx, followingQuery, followerID, followingID).Scan(&followEntryID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			slog.Error("duplicate follow relationship", "error", err, "followerID", followerID, "followingID", followingID)
			return 0,ErrFollowAlreadyExists
		}
		slog.Error("failed to insert follow", "error", err, "followerID", followerID, "followingID", followingID)
		return 0,err
	}
	slog.Info("follow insert successful", "inserted_id", followEntryID, "followerID", followerID, "followingID", followingID)

	// Update followers count for the person being followed (followee)
	followerCountUpdateQuery := `
		UPDATE users
		SET followers_count = COALESCE(followers_count, 0) + 1
		WHERE id = $1
	`
	res, err := tx.ExecContext(ctx, followerCountUpdateQuery, followingID)
	if err != nil {
		slog.Error("failed to update followers_count", "error", err, "followeeID", followingID)
		return 0,err
	}
	rows, _ := res.RowsAffected()
	slog.Info("followers_count updated", "rows_affected", rows, "followeeID", followingID)

	// Update following count of the person who clicked follow (follower)
	followingCountUpdateQuery := `
		UPDATE users
		SET following_count = COALESCE(following_count, 0) + 1
		WHERE id = $1
	`
	res, err = tx.ExecContext(ctx, followingCountUpdateQuery, followerID)
	if err != nil {
		slog.Error("failed to update following_count", "error", err, "followerID", followerID)
		return 0,err
	}
	rows, _ = res.RowsAffected()
	slog.Info("following_count updated", "rows_affected", rows, "followerID", followerID)

	// Commit all transaction steps atomically
	err = tx.Commit()
	if err != nil {
		return 0,err
	}

	return followEntryID,nil
}

// func that updates follower count
func(f *FollowDBModel) UpdateFollowerCount(userID uint) (bool,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		Update 
			users
		set
			follower_count = follower_count + 1
		where
			id=$1 
	`

	res,err := f.DB.ExecContext(ctx,query)
	if err!= nil {
		return false,err
	}
	rowNum,err := res.RowsAffected() ; 
	if err != nil {
		return false,err
	}

	//  checking if query was a successfull job but was there any actual insertion
	if rowNum == 0 {
		return false,sql.ErrNoRows
	}
	return true,nil

}

// func that updates follower count - passing client id of who is follower
func(f *FollowDBModel) UpdateFollowingCount(userID uint) (bool,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		Update 
			users
		set
			following_count = following_count + 1
		where
			id=$1 
	`

	res,err := f.DB.ExecContext(ctx,query)
	if err!= nil {
		return false,err
	}
	rowNum,err := res.RowsAffected() ; 
	if err != nil {
		return false,err
	}

	//  checking if query was a successfull job but was there any actual insertion
	if rowNum == 0 {
		return false,sql.ErrNoRows
	}
	return true,nil

}

// get follow details of any follow happened by its entryID <- handler assigns
func(f *FollowDBModel) GetFollowDetailsByFollowerUserID	(followSenderID uint) (*models.FollowPayload,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		Select
			follower_id as follow_sender_id,followee_id as follow_reciever_id
		from
			follows
		where
			id=$1
	`

	resRow := f.DB.QueryRowContext(ctx,query,followSenderID)
	var payload models.FollowPayload	
	err := resRow.Scan(
		&payload.FollowSenderID,
		&payload.FollowRecieverID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil,sql.ErrNoRows
		}
		return nil,err
	}

	return &payload,nil
}
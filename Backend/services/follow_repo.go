package services

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

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

// add followers,folllowee id into followes,updates users both flwr/wing count for them, if followed- following count,and who followed his flwr count +1
func(f *FollowDBModel) FollowUserTransaction(followerID,followingID uint) error {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// flow -  begin transaction,commit, or rollback
	tx,err := f.DB.BeginTx(ctx,nil) //* tx now holds all db operations
	if err!= nil {
		return err
	}

	defer func() {
		// ! if at the end, still hit any err, rollback changes
		if err != nil {
			tx.Rollback()
		}
	}()



	//# Single atomic transaction -
	
	//  for following a user,relationship added

		followingQuery := `
			Insert into
				follows(follower_id,followee_id)
			Values
				($1,$2)

		`
		res,err := tx.ExecContext(ctx,followingQuery,followerID,followingID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				slog.Error("duplicate follow relationship","error",err,"followerID",followerID,"followingID",followingID)
				return ErrFollowAlreadyExists
			}
			slog.Error("failed to insert follow","error",err,"followerID",followerID,"followingID",followingID)
			return err
		}
		rows,err := res.RowsAffected() // how many rows were changed
		if err != nil {
			slog.Error("failed to read follow insert rows affected","error",err,"followerID",followerID,"followingID",followingID)
			return err
		}
		if rows == 0 {
			slog.Error("no follow row inserted","followerID",followerID,"followingID",followingID)
			return sql.ErrNoRows // if 0 -> return sql.ErrNoRows
		}
		slog.Info("follow insert successful","rows_affected",rows,"followerID",followerID,"followingID",followingID)

	// update followers count for whom client might be following(followee)
	// bug - had update with no prior inserted value
	// fix - fixed with adding coalesce(whichField,setDefaultifNotThere) - fix null values
		followerCountUpdateQuery := `
			update
				users
			set
				followers_count = COALESCE(followers_count, 0) + 1
			where 
				id=$1 
		`
		res,err = tx.ExecContext(ctx,followerCountUpdateQuery,followingID)
		if err != nil {
			slog.Error("failed to update followers_count","error",err,"followeeID",followingID)
			return err
		}
		rows,_ = res.RowsAffected()
		slog.Info("followers_count updated","rows_affected",rows,"followeeID",followingID)

		
	// update following count of whom who is following someone
	followingCountUpdateQuery := `
			update
				users
			set
				following_count = COALESCE(following_count, 0) + 1
			where 
				id=$1 
		`
		res,err = tx.ExecContext(ctx,followingCountUpdateQuery,followerID)
		if err != nil {
			slog.Error("failed to update following_count","error",err,"followerID",followerID)
			return err
		}
		rows,_ = res.RowsAffected()
		slog.Info("following_count updated","rows_affected",rows,"followerID",followerID)

	
		// & commiting all transaction at once
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return err
		} 

		return nil
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
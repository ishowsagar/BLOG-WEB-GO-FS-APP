package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// type struct that stores db of type *sql.DB
type LikeDBModel struct {
	DB *sql.DB
}

// func that returns the instance of type LikeDbModel - which stores method belongs to this type related to like db query calls
func NewLikeDbModel(dB *sql.DB) *LikeDBModel {
	return &LikeDBModel{
		DB: dB,
	}
}

// @ interface -> stores all method related to LikeDBModel type struct <- implemented by LikeDBModel type
type LikeStore interface {
	CheckIfAlreadyLiked(like models.Like) (*models.Like,error)
	PostFirstLike(like models.Like) (*models.Like,error) 
	UpdateLikeByOneCount(like models.Like) (*models.Like,error)
}

// func that checks if user has already liked post
func(l *LikeDBModel) CheckIfAlreadyLiked(like models.Like) (*models.Like,error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		select
			id,post_id,user_id,like_count,liked_at
		from
			likes
		where
			post_id=$1 and user_id=$2
	`	

	resRow := l.DB.QueryRowContext(ctx,query,like.PostID,like.UserID)
	var likeVar models.Like
	
	err := resRow.Scan(
		&likeVar.ID,
		&likeVar.PostID,
		&likeVar.UserID,
		&likeVar.LikeCount,
		&likeVar.LikedAt,
	)

	// * we needed check for no rows,because no err is returned implicitly by qrC, so we had to check we knew query ran but was there any row
	if err != nil {
		if err == sql.ErrNoRows {
			return nil,sql.ErrNoRows
		}
		return nil,err
	}

	// no res returned but query ran
	

	return &likeVar,nil
}

// method that belongs to the LikeDBModel which -> updates like of a post
func(l *LikeDBModel) PostFirstLike(like models.Like) (*models.Like,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// Insert a new like record for the first like by this user on the post.
	// bug - due to this is new - had to insert values first,can't directly update cuase it needs value already which would be inserted by insert query firstly
	// fixed - changed to insert query
	query := `
		insert into likes(post_id,user_id,like_count,liked_at)
		values($1,$2,$3,$4)
		returning id,post_id,user_id,like_count,liked_at
	`

	resRow := l.DB.QueryRowContext(ctx, query, like.PostID, like.UserID,1,time.Now())
	var likeVar models.Like
	
	err := resRow.Scan(
		&likeVar.ID,
		&likeVar.PostID,
		&likeVar.UserID,
		&likeVar.LikeCount,
		&likeVar.LikedAt,
	)

	if err != nil {
		return nil,err
	}


	return &likeVar,nil
}


// func that belongs to the LikeDbModel which -> updates already Liked post like by +1
func(l *LikeDBModel) UpdateLikeByOneCount(like models.Like) (*models.Like,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// Increment like_count for a specific user's like on a post.
	query := `
		UPDATE likes
		SET like_count = like_count + 1,
		    liked_at = $1
		WHERE post_id = $2 AND user_id = $3
		RETURNING id, post_id, user_id, like_count, liked_at
	`

	resRow := l.DB.QueryRowContext(ctx, query, time.Now(), like.PostID, like.UserID)
	var likeVar models.Like
	
	err := resRow.Scan(
		&likeVar.ID,
		&likeVar.PostID,
		&likeVar.UserID,
		&likeVar.LikeCount,
		&likeVar.LikedAt, 
	)

	if err != nil {
		return nil,err
	}


	return &likeVar,nil
}
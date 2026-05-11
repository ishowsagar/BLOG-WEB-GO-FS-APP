package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @types

//  type CommentDBModel that stores DB of type *sql.DB
type CommentDBModel struct {
	DB *sql.DB
}

//  returns the instance of type CommentDBModel
func NewCommentDbModel(db *sql.DB) *CommentDBModel {
	return &CommentDBModel{
		DB: db,
	}
}

// @ Interface -> Stores all the methods which belongs to the CommentDbModel type
type CommentStore interface {
	PostComment(Comment models.Comment) (*models.Comment,error)
	LoadAllComments() ([]*models.Comment,error)
} 

// @ methods belongs to the type -> CommentDBModel which stores db of type *sql.DB
func(c *CommentDBModel) PostComment(comment models.Comment) (*models.Comment,error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		insert into comments(user_id,post_id,content,created_at,updated_at)
		values($1,$2,$3,$4,$5)
		returning user_id,post_id,content,created_at,updated_at
	`

	resRow := c.DB.QueryRowContext(ctx,query,
		comment.UserID,
		comment.PostID,
		comment.Content,
		time.Now(),
		time.Now(),
	)

	var commentVar models.Comment
	err := resRow.Scan(
		&commentVar.UserID,
		&commentVar.PostID,
		&commentVar.Content,
		&commentVar.CreatedAt,
		&commentVar.UpdatedAt,
	)

	if err != nil {
		return nil,err
	}
	// resRow.
	// // resulting row's checking id --> assigned by query through the DB call
	// retrievedID,_ := resRow.LastInsertId()
	// //* adding id to passed user to attach id on it
	// comment.ID = uint(retrievedID) 
	return &commentVar,nil
}

//  to get all comments
func(c *CommentDBModel) LoadAllComments() ([]*models.Comment,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// * returning cause we are inserting, insert does not returm res by default
	query := `
		select 
			id,user_id,post_id,content,created_at,updated_at
		from
			comments
		where
			id > $1
	`

	resRows,err := c.DB.QueryContext(ctx,query,0)

	var posts []*models.Comment // [] of models.Post type of elements
	for resRows.Next() {
		// looping into every res row
		var postVar models.Comment
	
		err := resRows.Scan(
		&postVar.ID,
		&postVar.UserID,
		&postVar.PostID,
		&postVar.Content,
		&postVar.CreatedAt,
		&postVar.UpdatedAt,
		)

		if err != nil {
			return nil,err
		}

		//  if scanned each iteration into variable
		// append each one by one to the posts

		posts = append(posts, &postVar)
	}

	if err != nil {
		return nil,err
	}

	// after fully scanning every resRow into var and appending, return posts []ice
	return posts,nil
}



func(c *CommentDBModel) DeleteCommentByUserID(userID uint,postID uint) error {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()


	query := `
		delete 
		from 
			comments
		where
			user_id=$1 
			and
			post_id=$2
	`

	resRow,err := c.DB.ExecContext(ctx,query,userID,postID)
	if err != nil {
		return err
	} 

	rows,err := resRow.RowsAffected()
	if err != nil {
		return err
	} 

	if rows == 0 {
		return sql.ErrNoRows // no row affected
	}

	return nil
}


// loads all comments of passed postID
func(c *CommentDBModel) LoadAllCommentsOfPost(postID uint) ([]*models.Comment,error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		Select
			id,user_id,post_id,content,created_at,updated_at
		from
			comments
		where 
			post_id=$1
	`

	resRows,err := c.DB.QueryContext(ctx,query,postID) // load all comments need to store in []
	if err!= nil {
		return nil,err
	}

	defer resRows.Close() // must close

	var comments []*models.Comment
	for resRows.Next() {
		var comment models.Comment
		// going through each entry got from res
		err = resRows.Scan(
			// unloading each res entry into comment var struct that holds same type struct data
			&comment.ID,
			&comment.UserID,
			&comment.PostID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		// * this could return empty comment struct 
		if err != nil {
			return nil,err
		}

		comments = append(comments, &comment)
	}

	// if each entry loaded into []of elements of type models.Comment
	// return it
	return comments,nil
}
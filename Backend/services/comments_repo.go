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

// type which stores comments data with name 
type CommentsData struct {
	ID uint `json:"id" gorm:"primaryKey"`
	 // for gorm refrence purpose
	User models.User `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	Post models.Post `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:PostID;references:ID"` //bug - if we delete post, we had to delete data associated with that ids
	UserID uint `json:"user_id" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	PostID	uint `json:"post_id"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:PostID;references:ID"`
	Content string `json:"content" binding:"required"`
	CreatedAt time.Time `json:"created_at" time_format="2006-01-02"`
	UpdatedAt time.Time `json:"updated_at" time_format="2006-01-02"`
	Name string `json:"name"` // need json fields only cause we sending data and it should be mapped with json tags for ease
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
func(c *CommentDBModel) LoadAllCommentsOfPost(postID uint) ([]*CommentsData,error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		Select
			c.id,c.user_id,c.post_id,c.content,c.created_at,c.updated_at,u.name
		from
			comments c
		Left join
			users u
		on
			u.id = c.user_id
		where 
			post_id=$1
	`

	resRows,err := c.DB.QueryContext(ctx,query,postID) // load all comments need to store in []
	if err!= nil {
		return nil,err
	}

	defer resRows.Close() // must close

	var comments []*CommentsData
	for resRows.Next() {
		var comment CommentsData
		// going through each entry got from res
		err = resRows.Scan(
			// unloading each res entry into comment var struct that holds same type struct data
			// id,user_id,post_id,content,created_at,updated_at,u.name
			&comment.ID,
			&comment.UserID,
			&comment.PostID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.Name,
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
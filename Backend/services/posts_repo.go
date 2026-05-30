package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @types

//  type UserDBModel that stores db of type *sql.DB
type PostDBModel struct {
	db *sql.DB
}

// type for joined post data with user and likes details
type BatchPost struct {
	ID uint `json:"id" gorm:"primaryKey"`
	// foregein key refs for corelation
	User models.User `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`// gorm ref
	UserID uint `json:"user_id" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	Title string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	LikeCount uint `json:"like_count" gorm:"-"`
	CreatedAt time.Time `json:"created_at" time_format="2006-01-02"`
	UpdatedAt time.Time `json:"updated_at" time_format="2006-01-02"`
	Name string `json:"name"`	
}




// func that returns instance of type PostDbModel -> that stores methods of post related methods 
func NewPostDbModel(db *sql.DB) *PostDBModel {
	return &PostDBModel{
		db: db,
	}
}

// @ Interface -> Stores all the methods which belongs to the PostDbModel type
type PostStore interface {
	CreatePost(Post models.Post) (*models.Post,error)
	DeletePostById(id uint) error
	LoadFeed() ([]*models.Post,error)
} 

// @ methods belongs to the type -> PostDBModel which stores db of type *sql.DB
func(p *PostDBModel) CreatePost(post models.Post) (*models.Post,error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// * returning cause we are inserting, insert does not returm res by default
	query := `
		insert into posts(user_id,title,content,created_at,updated_at)
		values($1,$2,$3,$4,$5)
		returning id,user_id,title,content,created_at,updated_at
	`

	resRow := p.db.QueryRowContext(ctx,query,
		post.UserID,
		post.Title,
		post.Content,
		time.Now(),
		time.Now(),
	)

	var postVar models.Post
	err := resRow.Scan(
		&postVar.ID,
		&postVar.UserID,
		&postVar.Title,
		&postVar.Content,
		&postVar.CreatedAt,
		&postVar.UpdatedAt,
	)

	if err != nil {
		return nil,err
	}

	// // resulting row's checking id --> assigned by query through the db call
	// retrievedID,_ := resRow.LastInsertId()
	// //* adding id to passed user to attach id on it
	// post.ID = uint(retrievedID) 
	return &postVar,nil
}


//  func that deletes post from db
func(p *PostDBModel) DeletePostById(id uint) error {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()


	query := `
		Delete from
			posts
		where id=$1
	` 	

	resRpw,err := p.db.ExecContext(ctx,query,id)
	if err != nil {
		return err
	}

	rows,err := resRpw.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil

}

//  to get all posts
func(p *PostDBModel) LoadFeed() ([]*models.Post,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// * returning cause we are inserting, insert does not returm res by default
	query := `
		select 
			p.id,p.user_id,p.title,p.content,p.created_at,p.updated_at,
			COALESCE(l.like_count, 0)
		from
			posts p
		left join (
			select post_id, SUM(like_count) as like_count
			from likes
			group by post_id
		) l on l.post_id = p.id
		where
			p.id > $1
	`

	resRows,err := p.db.QueryContext(ctx,query,0)

	var posts []*models.Post // [] of models.Post type of elements
	for resRows.Next() {
		// looping into every res row
		var postVar models.Post
	
		err := resRows.Scan(
		&postVar.ID,
		&postVar.UserID,
		&postVar.Title,
		&postVar.Content,
		&postVar.CreatedAt,
		&postVar.UpdatedAt,
		&postVar.LikeCount,
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


// method that fetches post by its id
func(p *PostDBModel) GetPostbyID(postID int)(*models.Post,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()


	query := `
		select
			p.id,p.user_id,p.title,p.content,p.created_at,p.updated_at,
			COALESCE(l.like_count, 0)
		from
			posts p
		left join (
			select post_id, SUM(like_count) as like_count
			from likes
			group by post_id
		) l on l.post_id = p.id
		where 
			p.id=$1
	`

	resRow := p.db.QueryRowContext(ctx,query,postID)
	var post models.Post
	err := resRow.Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.LikeCount,
	) 

	if err!= nil {
		if err == sql.ErrNoRows {
			return nil,sql.ErrNoRows
		}
		return nil,err
	}

	return &post,nil


}



func(p *PostDBModel) GetPostCountForAUser(userId uint) (int,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		select 
			count(*)
		from
			posts
		where
			user_id=$1 
	`

	resNum := p.db.QueryRowContext(ctx,query,userId)
	var noOfPost int
	err := resNum.Scan(
		&noOfPost,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0,sql.ErrNoRows
		}
		return 0,err
	}

	return noOfPost,nil
	
}

// retrieves posts by limit asked for 
func(p *PostDBModel) GetFeedByBatches(limit int,nextCursor string) (posts []*BatchPost,err error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// if nextCursor qparam is not provided -> set to largest, so it won't block feed fetch <- visual representation to learn better
	if nextCursor == "" {
		nextCursor = "999999"
	}
	query := `
	Select 
	 	p.id,p.user_id,p.title,p.content,p.created_at,p.updated_at,coalesce(l.like_count,0) as like_count,coalesce(u.name,'instaPowerUser')
	from
		posts p
	left join
    Users u
    On
    u.id = p.user_id
    Left join (
        select 
            post_id,sum(like_count) as like_count
            from 
                likes l
            group by post_id
    ) l
    on l.post_id = p.id
    where 
        p.id < $1
    order by 
        p.id desc
	limit $2 
	`

	// since data will be fetched as desc order, so like if nextCursor is 20, we load posts less than 20,
	resRows,err := p.db.QueryContext(ctx,query,nextCursor,limit)
	if err != nil {
		// has more little confusion
		return nil,err
	}

	defer resRows.Close()

	var postsBatch []*BatchPost
	for resRows.Next() {
		var post BatchPost
		err := resRows.Scan(
			// p.id,p.user_id,p.title,p.content,p.created_at,p.updated_at,coalesce(l.like_count,0) as like_count,u.name
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.LikeCount,
		&post.Name,
	) 

	if err!= nil {
		return nil,err
	}
	postsBatch = append(postsBatch,&post)
	}

	return postsBatch,nil
}



//  method that belongs to the PostDbModel which -> fetch all posts associted with user
func(p *PostDBModel) GetPostsOfAnyUserByUserID(userID uint) ([]*BatchPost,error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()


	query := `
		SELECT
			p.id,p.user_id,p.title,p.content,p.created_at,p.updated_at,
			u.name,
			coalesce(l.like_count, 0) as like_count
		FROM
			posts p
		left join
			users u
		on
			u.id = p.user_id
		left join
		(
			select
				 post_id,count(like_count) as like_count
			from
				likes
			group by
				post_id
		)
			l
		on 
			l.post_id = p.id
		where
			p.user_id = $1;

	`

	resRows,err := p.db.QueryContext(ctx,query,userID)
	if err != nil {
		return nil,err
	}

	var associatedPosts []*BatchPost
	for resRows.Next() {
		var post BatchPost
		err = resRows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Name,
			&post.LikeCount,
		)
		if err != nil {
			return nil,err
		}

		associatedPosts = append(associatedPosts, &post)
	}

	return associatedPosts,nil
}



func(p *PostDBModel) UpdatePostToStoreImageUrlByPostID(imageSrcUrl string,postID uint)(error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()


	query := `
		Update
			posts
		set
			image_source=$1
		where
			id=$2
	`

	updatedRow,err := p.db.ExecContext(ctx,query,imageSrcUrl,postID)
	if err != nil {
		return err
	}

	n,err := updatedRow.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
}
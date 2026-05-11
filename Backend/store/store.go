package store

import "github.com/ishowsagar/go-blog-web-application/services"

// @ interface -> stores all concrete interfaces <- implemented by corresponding dbmodel types
type MasterStore struct {
	LikeStore services.LikeStore
	UserStore services.UserStore
	PostStore services.PostStore
	CommentStore services.CommentStore
}

//  returns instance of type MasterStore which -> stores interface implemented by corresponding types
func NewMasterStore(likeDbModel *services.LikeDBModel,userDbModel *services.UserDBModel,postDbModel *services.PostDBModel,commentDbModel *services.CommentDBModel) MasterStore {
	return MasterStore{
		LikeStore:likeDbModel,
		UserStore:userDbModel,
		PostStore: postDbModel,
		CommentStore: commentDbModel,
	}
}


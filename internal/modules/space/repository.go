package space

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IPostRespository interface {
 	CreatePost(ctx context.Context, post *Post) error
	FindByID(ctx context.Context, postID uuid.UUID) (*Post, error)
	FindFeed(ctx context.Context, tag string, cursor time.Time, limit int) ([]Post, error)
	UpdateCommentCount(ctx context.Context, postID uuid.UUID, delta int) error
	UpdateLikeCount(ctx context.Context, postID uuid.UUID, delta int) error
	UpdateEmbedding(ctx context.Context, postID uuid.UUID, embedding []float64, model string) error
	DeletePost(ctx context.Context, postID uuid.UUID) error
}

type ICommentRepository interface {
	CreateComment(ctx context.Context, comment *Comment) error
	FindByPostID(ctx context.Context, postID uuid.UUID, cursor time.Time, limit int) ([]Comment, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID) error
}

type ILikeRepository interface {
	CreateLike(ctx context.Context, like *PostLike) error
	DeleteLike(ctx context.Context, postID, userID uuid.UUID) error
	Exists(ctx context.Context, postID, userID uuid.UUID) (bool, error)
	ExistsBatch(ctx context.Context, postIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]bool, error)
}

// -- Post Repository Entry
type PostRepository struct {
  db *gorm.DB
}
func NewPostRepository(db *gorm.DB) *PostRepository {
  return &PostRepository{db: db}
}

func (pr *PostRepository) CreatePost(ctx context.Context, post *Post) error {
  return pr.db.WithContext(ctx).Create(post).Error
}

func (pr *PostRepository) FindById(ctx context.Context, postID uuid.UUID) (*Post, error) {
  var post Post

  err := pr.db.WithContext(ctx).Where("id = ?", postID).First(&post).Error
  if err != nil {
    return nil, err
  }

  return &post, nil
}

func (pr *PostRepository) FindFeed(ctx context.Context, tag string, cursor time.Time, limit int) ([]Post, error) {
  var post []Post

  db := pr.db.WithContext(ctx).Order("created_at desc").Limit(limit).Preload("User")

  if tag != "" {
    db = pr.db.WithContext(ctx).Where("tag = ?", tag)
  }

  if !cursor.IsZero() {
    db = pr.db.WithContext(ctx).Where("created_at < ?", cursor)
  }

  err := db.Find(&post).Error
  if err != nil {
    return nil, err
  }

  return post, nil
}

func (pr *PostRepository) UpdateCommentCount(ctx context.Context, postID uuid.UUID, delta int) error {
  err := pr.db.
    WithContext(ctx).
    Model(&Post{}).Where("id = ?", postID).
    Update("comment_count", gorm.Expr("comment_count + ?", delta)).
    Error
  if err != nil {
    return err
  }
  
  return nil
}

func (pr *PostRepository) UpdateLikeCount(ctx context.Context, postID uuid.UUID, delta int) error {
  err := pr.db.
    WithContext(ctx).
    Model(&Post{}).Where("id = ?", postID).
    Update("like_count", gorm.Expr("like_count + ?", delta)).
    Error
  if err != nil {
    return err
  }
  
  return nil
}

// func (pr *PostRepository) UpdateEmbedding(ctx context.Context, postID uuid.UUID, embedding []float64, model string) error {
  
// }

func (pr *PostRepository) DeletePost(ctx context.Context, postID uuid.UUID) error {
  err := pr.db.
    WithContext(ctx).
    Delete(&Post{}, "id = ?", postID).
    Error
  if err != nil {
    return err
  }
  
  return nil
}

// -- Comment Repository Entry
type CommentRepository struct {
  db *gorm.DB
}
func NewCommentRepository(db *gorm.DB) *CommentRepository {
  return &CommentRepository{db: db}
}

func (cr *CommentRepository) CreateComment(ctx context.Context, comment *Comment) error {
  return cr.db.WithContext(ctx).Create(comment).Error
}

func (cr *CommentRepository) FindByPostID(ctx context.Context, 
  postID uuid.UUID, 
  cursor time.Time, 
  limit int) ([]Comment, error) {
    var comments []Comment

    db := cr.db.WithContext(ctx).Where("post_id = ?", postID).Order("created_at DESC").Limit(limit)
   
   if !cursor.IsZero(){
     db = db.Where("created_at < ?", cursor)
   } 

   err := db.Find(&comments).Error
   if err != nil {
     return nil, err
   }

   return comments, nil
}

func (cr *CommentRepository) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
  return cr.db.WithContext(ctx).Where("id = ?", commentID).Delete(&Comment{}).Error
}


// -- Like Repository Entry
type LikeRepository struct {
  db *gorm.DB
}
func NewLikeRepository(db *gorm.DB) *LikeRepository {
  return &LikeRepository{db: db}
}

func (lr *LikeRepository) CreateLike(ctx context.Context, like *PostLike) error {
  return lr.db.WithContext(ctx).Create(like).Error
}

func (lr *LikeRepository) DeleteLike(ctx context.Context, postID, userID uuid.UUID) error {
  return lr.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).Delete(&PostLike{}).Error
}

func (lr *LikeRepository) Exists(ctx context.Context, postID, userID uuid.UUID) (bool, error) {
  var exist bool
  err := lr.db.WithContext(ctx).Model(&PostLike{}).Select("count(1) > 0").Where("post_id = ? AND user_id = ?", postID, userID).Find(&exist).Error;
  if err != nil {
    return false, err
  }
  
  return exist, nil
}

func (lr *LikeRepository) ExistsBatch(ctx context.Context, postIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]bool, error) {
  existsMap := make(map[uuid.UUID]bool)
	
	for _, id := range postIDs {
		existsMap[id] = false
	}
	
	if len(postIDs) == 0 {
		return existsMap, nil
	}
  
	var likedPostIDs []uuid.UUID
	err := lr.db.WithContext(ctx).
		Model(&PostLike{}).
		Where("user_id = ? AND post_id IN ?", userID, postIDs). 
		Pluck("post_id", &likedPostIDs).                       
		Error
	if err != nil {
		return nil, err
	}
	
	for _, id := range likedPostIDs {
		existsMap[id] = true
	}
  
	return existsMap, nil
}

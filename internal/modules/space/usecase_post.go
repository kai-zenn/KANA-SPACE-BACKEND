package space

import (
	"KANA-SPACE-BACKEND/internal/pkgs/storage"
	"context"

	"github.com/google/uuid"
)


type IPostUseCase interface {
  NewPost(ctx context.Context, req CreatePostRequest) (*PostResponse, error)
  FindPostByID(ctx context.Context, postID uuid.UUID) (*PostResponse, error)
  FindFeed(ctx context.Context, req FeedQueryParam) (*FeedResponse, error)
  UpdateCommentCount(ctx context.Context, postID uuid.UUID, delta int) error
  // UpdateEmbedding(ctx context.Context, postID uuid.UUID, embedding []float64, model string) error
  DeletePost(ctx context.Context, postID uuid.UUID) error
}

type PostUseCase struct {
  pr IPostRespository
  storage storage.Interface
}

func NewPostUserCase(pr IPostRespository, storage storage.Interface) IPostUseCase {
  return &PostUseCase{
    pr: pr,
    storage: storage,
  }
}

func (pu *PostUseCase) NewPost(ctx context.Context, req CreatePostRequest) (*PostResponse, error) {
  newPhotoUrl, err := pu.storage.UploadPostImages(ctx, req.Images)
  if err != nil {
    return nil, err
  }

  defer func() {
    if err != nil && len(newPhotoUrl) > 0 {
			go func() {
				_ = pu.storage.DeletePostImages(context.Background(), newPhotoUrl)
			}()
		}
  }()

  var postImgs []PostImage
  for _, url := range newPhotoUrl {
    postImgs = append(postImgs, PostImage{
      ID:   uuid.New(),
      URL:  url,
      PostID: req.ID,
    })
  }

  post := &Post{
    ID:        req.ID,
    UserID:    req.UserID,
    Content:   req.Content,
    Tag:       req.Tag,
    Latitude:  req.Latitude,
    Longitude: req.Longitude,
    Images:    postImgs,
  }

  err = pu.pr.CreatePost(ctx, post)
  if err != nil {
    return nil, err
  }

  return &PostResponse{
    ID:            post.ID,
    UserID:        PostAuthor{ID: post.UserID},
    Content:       post.Content,
    Tag:           post.Tag,
    PhotoURLs:     newPhotoUrl,
    LikeCount:     post.LikeCount,
    CommentCount:  post.CommentCount,
    CreatedAt:     post.CreatedAt,
  }, nil
}

func (pu *PostUseCase) FindPostByID(ctx context.Context, postID uuid.UUID) (*PostResponse, error) {
  post, err := pu.pr.FindByID(ctx, postID)
  if err != nil {
    return nil, err
  }

  return &PostResponse{
    ID:            post.ID,
    UserID:        PostAuthor{ID: post.UserID},
    Content:       post.Content,
    Tag:           post.Tag,
    PhotoURLs:     ,
    LikeCount:     post.LikeCount,
    CommentCount:  post.CommentCount,
    CreatedAt:     post.CreatedAt,
  }, nil
}

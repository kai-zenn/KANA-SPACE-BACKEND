package space

import (
	"context"

	"github.com/google/uuid"
)

type ILikeUseCase interface {
  LikePost(ctx context.Context, req LikeParam) (*LikeResponse, error)
  UnlikePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) (*LikeResponse, error)
}

type LikeUseCase struct {
  pr PostRepository
  lr LikeRepository
}

func NewLikeUseCase(lr LikeRepository, pr PostRepository) *LikeUseCase {
  return &LikeUseCase{
    lr: lr,
    pr: pr,
  }
}

func (lu *LikeUseCase) LikePost(ctx context.Context, req LikeParam) (*LikeResponse, error) {
  exists, _ := lu.lr.Exists(ctx, req.PostID, req.UserID)
  if exists {
    post, _ := lu.pr.FindById(ctx, req.PostID)
    return &LikeResponse{
      LikeCount: post.LikeCount,
      IsLiked:   true,
    }, nil
  }

  err := lu.lr.CreateLike(ctx, &PostLike{ID: uuid.New(), PostID: req.PostID, UserID: req.UserID})
  if err != nil {
    return nil, err
  }

  err = lu.pr.UpdateLikeCount(ctx, req.PostID, +1)
  if err != nil {
    return nil, err
  }

  post, _ := lu.pr.FindById(ctx, req.PostID)
  return &LikeResponse{
    LikeCount: post.LikeCount,
    IsLiked:   true,
  }, nil
}

func (lu *LikeUseCase) UnlikePost(ctx context.Context, req LikeParam) (*LikeResponse, error) {
  exists, _ := lu.lr.Exists(ctx, req.PostID, req.UserID)
  if !exists {
    post, _ := lu.pr.FindById(ctx, req.PostID)
    return &LikeResponse{
      LikeCount: post.LikeCount,
      IsLiked:   false,
    }, nil
  }

  err := lu.lr.DeleteLike(ctx, req.PostID, req.UserID)
  if err != nil {
    return nil, err
  }

  err = lu.pr.UpdateLikeCount(ctx, req.PostID, -1)
  if err != nil {
    return nil, err
  }

  post, _ := lu.pr.FindById(ctx, req.PostID)
  return &LikeResponse{
    LikeCount: post.LikeCount,
    IsLiked:   false,
  }, nil
}

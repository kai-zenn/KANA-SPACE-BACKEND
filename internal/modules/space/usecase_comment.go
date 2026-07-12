package space

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)


type ICommentUseCase interface {
  CreateComment(ctx context.Context, req CreateCommentRequest) (*CommentResponse, error)
  GetComments(ctx context.Context, postID uuid.UUID, param CommentQueryParam) (*CommentsResponse, error)
  DeleteComment(ctx context.Context, commentID, requesterID uuid.UUID, requestRole string) error
}

type CommentUseCase struct {
  cr ICommentRepository
  pr IPostRepository
}

func NewCommentUseCase(cr ICommentRepository, pr IPostRepository) ICommentUseCase {
  return &CommentUseCase{
    cr: cr,
    pr: pr,
  }
}

func (cu *CommentUseCase) CreateComment(ctx context.Context, req CreateCommentRequest) (*CommentResponse, error) {
  comment := &Comment{
    ID: uuid.New(),
    PostID: req.PostID,
    UserID: req.UserID,
    Content: req.Content,
  }

  err := cu.cr.CreateComment(ctx, comment)
  if err != nil {
    return nil, err
  }

  cu.pr.UpdateCommentCount(ctx, req.PostID, +1)

  return &CommentResponse{
    ID: comment.ID,
    User: ToPostAuthor(comment.User),
    Content: comment.Content,
    CreatedAt: comment.CreatedAt,
  }, nil
}

func (cu *CommentUseCase) GetComments(ctx context.Context, postID uuid.UUID, param CommentQueryParam) (*CommentsResponse, error) {
  if param.Limit <= 0 || param.Limit > 10 {
    param.Limit = 5
  }

  var parsedCursor time.Time
  if param.Cursor != "" {
    var err error
    parsedCursor, err = time.Parse(time.RFC3339, param.Cursor)
    if err != nil {
      return nil, err
    }
  }
  
  comments, err := cu.cr.FindByPostID(ctx, postID, parsedCursor, param.Limit)
  if err != nil {
    return nil, err
  }

  responses := make([]CommentResponse, len(comments))
  for i, c := range comments {
    responses[i] = CommentResponse{
      ID: c.ID,
      User: ToPostAuthor(c.User),
      Content: c.Content,
      CreatedAt: c.CreatedAt,
    }
  }

  var nextCursor *time.Time
  if len(comments) == param.Limit {
    lastTime := &comments[len(comments)-1].CreatedAt
    nextCursor = lastTime
  }
  return &CommentsResponse{
    Comments: responses,
    NextCursor: nextCursor,
  }, nil
}

func (cu *CommentUseCase) DeleteComment(ctx context.Context, commentID, requesterID uuid.UUID, requestRole string) error {
  comment, err := cu.cr.FindByID(ctx, commentID)
  if err != nil {
    return err
  }

  if comment.UserID != requesterID && requestRole != "admin" {
    return errors.New("not the owner of this comment")
  }
  
  err = cu.cr.DeleteComment(ctx, commentID)
  if err != nil {
    return err
  }
  
  err = cu.pr.UpdateCommentCount(ctx, comment.PostID, -1)
  if err != nil {
    return err
  }
  
  return nil
}

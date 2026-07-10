package space

import (
	"KANA-SPACE-BACKEND/internal/modules/user"
	"KANA-SPACE-BACKEND/internal/pkgs/storage"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)


type IPostUseCase interface {
  NewPost(ctx context.Context, req CreatePostRequest) (*PostResponse, error)
  FindPostByID(ctx context.Context, postID uuid.UUID) (*PostResponse, error)
  GetFeed(ctx context.Context, viewerID uuid.UUID, req FeedQueryParam) (*FeedResponse, error)
  // UpdateEmbedding(ctx context.Context, postID uuid.UUID, embedding []float64, model string) error
  DeletePost(ctx context.Context, postID uuid.UUID, requesterID uuid.UUID, requesterRole string) error
}

type NLPClientInterface interface {
  Embed(ctx context.Context, text string) (embedding []float64, model string, err error)
}

type PostUseCase struct {
  pr IPostRepository
  cr ICommentRepository
  lr ILikeRepository
  nlp NLPClientInterface
  ur  user.IUserRepository
  storage storage.Interface
}

func ToPostAuthor(user user.User) PostAuthor {
	return PostAuthor{
		ID:       user.ID,
		Username: user.Username,
		ProfilePhotoLink: user.ProfilePhotoLink,
	}
}

func NewPostUserCase(pr IPostRepository, cr ICommentRepository, lr ILikeRepository, nlp NLPClientInterface, ur user.IUserRepository, storage storage.Interface) IPostUseCase {
  return &PostUseCase{
    pr: pr,
    cr: cr,
    lr: lr,
    nlp: nlp,
    ur: ur,
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

  
  postId := uuid.New()
  var postImgs []PostImage
  for _, url := range newPhotoUrl {
    postImgs = append(postImgs, PostImage{
      ID:   uuid.New(),
      URL:  url,
      PostID: postId,
    })
  }

  author, err := pu.ur.GetByID(ctx, req.UserID)
  if err != nil {
    return nil, err
  }

  post := &Post{
    ID:        postId,
    UserID:    req.UserID,
    Content:   req.Content,
    Tag:       req.Tag,
    Latitude:  req.Latitude,
    Longitude: req.Longitude,
    Images:    postImgs,
  }

  if req.Tag == "TagCariMaterial" {
    status := "RequestStatusActive"
    post.RequestStatus = &status
  }
  

  err = pu.pr.CreatePost(ctx, post)
  if err != nil {
    return nil, err
  }

  // if req.Tag == "CariMaterial" && pu.nlp != nil {
  //   go pu.nlp.EmbedPostAsync(post.ID, post.Content)
  // }

  return &PostResponse{
    ID:            post.ID,
    User:          ToPostAuthor(*author),
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

  author, err := pu.ur.GetByID(ctx, post.UserID)
  if err != nil {
    return nil, err
  }

  photoURLs := make([]string, len(post.Images))
  for i, img := range post.Images {
    photoURLs[i] = img.URL
  }

  return &PostResponse{
    ID:           post.ID,
    User:         ToPostAuthor(*author),
    Content:      post.Content,
    Tag:          post.Tag,
    PhotoURLs:    photoURLs,
    LikeCount:    post.LikeCount,
    CommentCount: post.CommentCount,
    CreatedAt:    post.CreatedAt,
  }, nil
}

func (pu *PostUseCase) GetFeed(ctx context.Context, viewerID uuid.UUID, req FeedQueryParam) (*FeedResponse, error) {
  limit := req.Limit
  if limit <= 0 || limit > 20 {
    limit = 15
  }

  var cursor time.Time
  if req.Cursor != "" {
    var err error
    cursor, err = time.Parse(time.RFC3339, req.Cursor)
    if err != nil {
      return nil, err
    }
  }

  posts, err := pu.pr.FindFeed(ctx, req.Tag, cursor, limit)
  if err != nil {
    return nil, err
  }

  var postIDs []uuid.UUID
  for _, p := range posts {
    postIDs = append(postIDs, p.ID)
  }
  likedMap, _ := pu.lr.ExistsBatch(ctx, postIDs, viewerID)
  
  response := make([]PostResponse, len(posts))
  for i, p := range posts {
    isLiked := likedMap[p.ID]
    
    photoURLs := make([]string, len(p.Images))
    for j, img := range p.Images {
      photoURLs[j] = img.URL
    }
    
    response[i] = PostResponse{
      ID: p.ID,
      User: ToPostAuthor(p.User),
      Content: p.Content,
      Tag: p.Tag,
      PhotoURLs: photoURLs,
      LikeCount: p.LikeCount,
      CommentCount: p.CommentCount,
      IsLiked: isLiked,
    }
  }

  var nextCursor *time.Time
  if len(posts) == limit {
    last := posts[len(posts)-1].CreatedAt
    nextCursor = &last
  }
  
  return &FeedResponse{
    Posts: response,
    NextCursor: nextCursor,
  }, nil
}

func (pu *PostUseCase) DeletePost(ctx context.Context, postID uuid.UUID, requesterID uuid.UUID, requesterRole string) error {
  post, err := pu.pr.FindByID(ctx, postID)
  if err != nil {
    return err
  }
  
  if post.UserID != requesterID && requesterRole != "admin" {
    return errors.New("you are not authorized to delete this post")
  }
  err = pu.pr.DeletePost(ctx, postID)
  if err != nil {
    return err
  }

  return nil
}

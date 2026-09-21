package lapak

import (
	"context"
	"fmt"
	"log"
	"time"

	"KANA-SPACE-BACKEND/internal/modules/notification"
	"KANA-SPACE-BACKEND/internal/modules/space"
	"KANA-SPACE-BACKEND/internal/pkgs/nlpclient"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const DefaultMatchRadiusMeters = 20000 // 20km

type MatchingTrigger interface {
	ProcessMatchAsync(postID uuid.UUID)
}

type IMatchingUseCase interface {
	ProcessMatchAsync(postID uuid.UUID)
}

type MatchingUseCase struct {
	productRepo    IProductRepository
	postRepo       space.IPostRepository
	matchRepo      IMatchRepository
	notificationUC notification.INotificationUseCase
	nlpClient      nlpclient.Client
	db             *gorm.DB
}

func NewMatchingUseCase(
	productRepo IProductRepository,
	postRepo space.IPostRepository,
	matchRepo IMatchRepository,
	notificationUC notification.INotificationUseCase,
	nlpClient nlpclient.Client,
	db *gorm.DB,
) *MatchingUseCase {
	return &MatchingUseCase{
		productRepo:    productRepo,
		postRepo:       postRepo,
		matchRepo:      matchRepo,
		notificationUC: notificationUC,
		nlpClient:      nlpClient,
		db:             db,
	}
}

func (mu *MatchingUseCase) ProcessMatchAsync(postID uuid.UUID) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Matching] panic recovered saat proses post %s: %v", postID, r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	post, err := mu.postRepo.FindByID(ctx, postID)
	if err != nil {
		log.Printf("[Matching] post %s tidak ditemukan: %v", postID, err)
		return
	}

	if post.Latitude == nil || post.Longitude == nil {
    log.Printf("[Matching] post %s tidak punya koordinat", postID)
    return
	}

	candidates, err := mu.productRepo.FindAvailableRawMaterialCandidates(ctx, *post.Latitude, *post.Longitude, DefaultMatchRadiusMeters)
	if err != nil {
		log.Printf("[Matching] gagal cari kandidat untuk post %s: %v", postID, err)
		return
	}
	if len(candidates) == 0 {
		mu.notifyNoCandidates(ctx, post.UserID)
		return
	}

	matches, err := mu.findMatches(ctx, post, candidates)
	if err != nil {
		log.Printf("[Matching] gagal total proses matching post %s: %v", postID, err)
		return
	}

	mu.notifyMatches(ctx, post, matches)
}

func (mu *MatchingUseCase) findMatches(ctx context.Context, post *space.Post, candidates []ProductCandidate) ([]Match, error) {
	nlpCandidates := make([]nlpclient.MatchCandidate, len(candidates))
	for i, c := range candidates {
	  embedding := make([]float64, len(c.Embedding))
    copy(embedding, c.Embedding)
		nlpCandidates[i] = nlpclient.MatchCandidate{
			ListingID:      c.ID.String(),
			Text:           c.Description,
			Embedding:      embedding,
			DistanceMeters: c.DistanceMeters,
		}
	}

	var queryEmbedding []float64
	if len(post.Embedding) > 0 {
    queryEmbedding = make([]float64, len(post.Embedding))
    copy(queryEmbedding, post.Embedding)
	}

	nlpReq := nlpclient.MatchRequest{
		RequestID:      post.ID.String(),
		QueryText:      post.Content,
		QueryEmbedding: queryEmbedding,
		Candidates:     nlpCandidates,
		Config: nlpclient.MatchRequestConfig{
			BM25TopK:          10,
			FinalTopK:         5,
			SemanticThreshold: 0.80,
			Alpha:             0.2,
		},
	}

	nlpResp, err := mu.nlpClient.Match(ctx, nlpReq)
	if err != nil {
		log.Printf("[Matching] NLP service gagal, fallback ke keyword search: %v", err)
		return mu.fallbackKeywordMatch(ctx, post, candidates)
	}

	return mu.persistMatches(ctx, post.ID, nlpResp.Matches)
}

func (mu *MatchingUseCase) fallbackKeywordMatch(ctx context.Context, post *space.Post, candidates []ProductCandidate) ([]Match, error) {
	log.Printf("[Matching] fallback keyword search untuk post %s", post.ID)

	if len(candidates) == 0 {
		return []Match{}, nil
	}

	candidateIDs := make([]string, len(candidates))
	for i, c := range candidates {
		candidateIDs[i] = c.ID.String()
	}

	type matchRow struct {
		ID    string
		Score float64
	}

	var rows []matchRow
	err := mu.db.WithContext(ctx).Raw(`
		SELECT id::text AS id,
		       ts_rank(search_vector, plainto_tsquery('simple', ?)) AS score
		FROM products
		WHERE id = ANY(?::uuid[])
		  AND search_vector @@ plainto_tsquery('simple', ?)
		ORDER BY score DESC
		LIMIT 3
	`, post.Content, candidateIDs, post.Content).Scan(&rows).Error
	if err != nil {
		log.Printf("[Matching] fallback query error: %v", err)
		return nil, err
	}

	matches := make([]Match, 0, len(rows))
	for i, r := range rows {
		listingID, err := uuid.Parse(r.ID)
		if err != nil {
			continue
		}
		matches = append(matches, Match{
			ID:            uuid.New(),
			RequestID:     post.ID,
			ListingID:     listingID,
			BM25Score:     r.Score,
			SemanticScore: 0.0,
			FinalScore:    r.Score,
			Rank:          i + 1,
			Status:        "SUGGESTED",
		})
	}

	if len(matches) > 0 {
		if err := mu.matchRepo.CreateMatches(ctx, matches); err != nil {
			return nil, fmt.Errorf("gagal simpan fallback matches: %w", err)
		}
	}

	log.Printf("[Matching] fallback keyword search selesai: %d matches", len(matches))
	return matches, nil
}

func (mu *MatchingUseCase) persistMatches(ctx context.Context, postID uuid.UUID, results []nlpclient.MatchResult) ([]Match, error) {
	var matches []Match
	for _, r := range results {
		listingID, err := uuid.Parse(r.ListingID)
		if err != nil {
			log.Printf("[Matching] skip match, listing_id invalid: %s", r.ListingID)
			continue
		}
		matches = append(matches, Match{
			ID:            uuid.New(),
			RequestID:     postID,
			ListingID:     listingID,
			BM25Score:     r.BM25Score,
			SemanticScore: r.SemanticScore,
			FinalScore:    r.FinalScore,
			Rank:          r.Rank,
			Status:        "SUGGESTED",
		})
	}

	if len(matches) > 0 {
		if err := mu.matchRepo.CreateMatches(ctx, matches); err != nil {
			return nil, fmt.Errorf("gagal simpan matches: %w", err)
		}
	}
	
	return matches, nil
}

func (mu *MatchingUseCase) notifyNoCandidates(ctx context.Context, userID uuid.UUID) {
  input := notification.NotifyInput{
    UserID: userID,
    Type:   "INFO",
    Title:  "Belum ada material di sekitar",
    Body:   "Saat ini belum ada material yang cocok di lokasi Anda. Kami akan memberi tahu jika ada yang baru.",
    Data:   map[string]string{},
  }
  
  if err := mu.notificationUC.Notify(ctx, input); err != nil {
    log.Printf("[Matching] gagal kirim notifikasi no-candidates: %v", err)
  }
}

func (mu *MatchingUseCase) notifyMatches(ctx context.Context, post *space.Post, matches []Match) {
  if len(matches) == 0 {
    return
  }

  inputs := make([]notification.NotifyInput, 0, len(matches))
  for _, match := range matches {
    product, err := mu.productRepo.FindByID(ctx, match.ListingID)
    if err != nil {
      log.Printf("[Matching] skip notif, produk %s tidak ditemukan: %v", match.ListingID, err)
      continue
    }

    var actionType, actionTarget string
    if product.ListingType == "HIBAH" {
      actionType = "DIRECT_CLAIM"
      actionTarget = "/lapak/products/" + product.ID.String() + "/transactions"
    } else {
      actionType = "OPEN_CHAT"
      actionTarget = "/chat/products/" + product.ID.String() + "/conversation"
    }

    inputs = append(inputs, notification.NotifyInput{
      UserID:        post.UserID,
      Type:          "MATCH_FOUND",
      Title:         "Material cocok ditemukan!",
      Body:          "Kami menemukan material yang cocok: " + product.Title,
      ReferenceType: "product",
      ReferenceID:   &product.ID,
      Data: map[string]string{
        "action_type":   actionType,
        "action_target": actionTarget,
      },
    })
  }

  if len(inputs) > 0 {
    if err := mu.notificationUC.NotifyBatch(ctx, inputs); err != nil {
      log.Printf("[Matching] gagal kirim notifikasi batch: %v", err)
    }
  }
}

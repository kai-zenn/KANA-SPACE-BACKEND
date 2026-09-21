package lapak

import (
	"testing"

	"KANA-SPACE-BACKEND/internal/pkgs/nlpclient"

	"gorm.io/gorm"
)

func TestMatchConfigIncludesAlpha(t *testing.T) {
	cfg := nlpclient.MatchRequestConfig{
		BM25TopK:          10,
		FinalTopK:         5,
		SemanticThreshold: 0.8,
		Alpha:             0.2,
	}
	if cfg.Alpha != 0.2 {
		t.Fatalf("expected Alpha to be 0.2, got %f", cfg.Alpha)
	}
}

func TestNewMatchingUseCaseAcceptsDB(t *testing.T) {
	_ = NewMatchingUseCase(nil, nil, nil, nil, nil, &gorm.DB{})
}

package nlpclient

type EmbedRequest struct {
	Text string `json:"text"`
}

type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
	Dim       int       `json:"dim"`
}

type MatchCandidate struct {
	ListingID      string    `json:"listing_id"` // = Product.ID.String()
	Text           string    `json:"text"`
	Embedding      []float64 `json:"embedding"`
	DistanceMeters float64   `json:"distance_meters"`
}

// MatchRequest dikirim ke /v1/match
type MatchRequest struct {
	RequestID      string             `json:"request_id"`
	QueryText      string             `json:"query_text"`
	QueryEmbedding []float64          `json:"query_embedding"`
	Candidates     []MatchCandidate   `json:"candidates"`
	Config         MatchRequestConfig `json:"config"`
}

type MatchRequestConfig struct {
	BM25TopK          int     `json:"bm25_top_k"`
	FinalTopK         int     `json:"final_top_k"`
	SemanticThreshold float64 `json:"semantic_threshold"`
	Alpha             float64 `json:"alpha"`
}

type MatchResult struct {
	ListingID     string  `json:"listing_id"` // WAJIB uuid.Parse() sebelum dipake
	BM25Score     float64 `json:"bm25_score"`
	SemanticScore float64 `json:"semantic_score"`
	FinalScore    float64 `json:"final_score"`
	Rank          int     `json:"rank"`
}

// MatchResponse balikan dari /v1/match
type MatchResponse struct {
	RequestID           string        `json:"request_id"`
	Matches             []MatchResult `json:"matches"`
	Stage2SurvivorCount int           `json:"stage2_survivor_count"`
	Stage3SurvivorCount int           `json:"stage3_survivor_count"`
	Model               string        `json:"model"`
	ProcessedInMs       int           `json:"processed_in_ms"`
}

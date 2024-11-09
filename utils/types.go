package utils

type Review struct {
	ReviewID   string `json:"review_id"`
	ProductID  string `json:"product_id"`
	ReviewerID string `json:"reviewer_id"`
	Stars      int    `json:"stars"`
	// Add other fields if needed
}

type ResultData struct {
	Recommendations map[string][]string `json:"recommendations"` // Map of user IDs to recommended product IDs
}

package utils

import (
	"math"
)

func PerformComputation(partition []Review) ResultData {
	// Map of user IDs to their ratings
	userRatings := make(map[string]map[string]int)

	// Build user ratings from the partition data
	for _, review := range partition {
		if _, exists := userRatings[review.ReviewerID]; !exists {
			userRatings[review.ReviewerID] = make(map[string]int)
		}
		userRatings[review.ReviewerID][review.ProductID] = review.Stars
	}

	// Compute similarities between users
	similarities := computeUserSimilarities(userRatings)

	// Generate recommendations for each user
	recommendations := make(map[string][]string)
	for userID := range userRatings {
		recommendations[userID] = recommendProducts(userID, userRatings, similarities)
	}

	return ResultData{Recommendations: recommendations}
}

func computeUserSimilarities(userRatings map[string]map[string]int) map[string]map[string]float64 {
	similarities := make(map[string]map[string]float64)

	users := make([]string, 0, len(userRatings))
	for user := range userRatings {
		users = append(users, user)
	}

	// Limit users to prevent OOM
	maxUsers := 1000 // Adjust as needed
	if len(users) > maxUsers {
		users = users[:maxUsers]
	}

	for i := 0; i < len(users); i++ {
		for j := i + 1; j < len(users); j++ {
			// Existing similarity computation
		}
	}
	return similarities
}

func recommendProducts(
	userID string,
	userRatings map[string]map[string]int,
	similarities map[string]map[string]float64,
) []string {
	scores := make(map[string]float64)
	totalSim := make(map[string]float64)

	for otherUserID, sim := range similarities[userID] {
		for productID, rating := range userRatings[otherUserID] {
			if _, rated := userRatings[userID][productID]; !rated {
				scores[productID] += sim * float64(rating)
				totalSim[productID] += sim
			}
		}
	}

	recommendations := make([]string, 0)
	for productID := range scores {
		// Compute the weighted average
		if totalSim[productID] != 0 {
			score := scores[productID] / totalSim[productID]
			if score >= 4.0 { // Threshold for recommendation
				recommendations = append(recommendations, productID)
			}
		}
	}

	return recommendations
}

func calculateCosineSimilarity(ratingsA, ratingsB map[string]int) float64 {
	var sumProduct, sumASq, sumBSq float64

	for itemID, ratingA := range ratingsA {
		if ratingB, exists := ratingsB[itemID]; exists {
			sumProduct += float64(ratingA * ratingB)
			sumASq += float64(ratingA * ratingA)
			sumBSq += float64(ratingB * ratingB)
		}
	}

	denominator := math.Sqrt(sumASq) * math.Sqrt(sumBSq)
	if denominator == 0 {
		return 0
	}
	return sumProduct / denominator
}

package utils

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

var productCategoryMap map[string]string // Map[productID]productCategory

func LoadData(filePath string) []Review {
	filePath = "/app/" + filePath
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening data file %s: %v", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV data from %s: %v", filePath, err)
	}

	if len(records) == 0 {
		log.Fatalf("No records found in CSV file %s", filePath)
	}

	var reviews []Review
	productCategoryMap = make(map[string]string)

	for _, record := range records[1:] { // Assuming first row is header
		// Adjust indices based on your CSV file structure
		// For example, assuming:
		// record[0]: review_id
		// record[1]: product_id
		// record[2]: reviewer_id
		// record[3]: stars
		// record[4]: product_category

		stars, _ := strconv.Atoi(record[3])
		review := Review{
			ReviewID:        record[0],
			ProductID:       record[1],
			ReviewerID:      record[2],
			Stars:           stars,
			ProductCategory: record[4],
		}
		reviews = append(reviews, review)
		productCategoryMap[review.ProductID] = review.ProductCategory
	}
	return reviews
}

func GetProductCategory(productID string) string {
	return productCategoryMap[productID]
}

func SplitData(data []Review, partitions int) [][]Review {
	var result [][]Review
	partitionSize := (len(data) + partitions - 1) / partitions

	for i := 0; i < len(data); i += partitionSize {
		end := i + partitionSize
		if end > len(data) {
			end = len(data)
		}
		result = append(result, data[i:end])
	}
	return result
}


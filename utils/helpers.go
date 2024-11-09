package utils

import (
	"encoding/csv"
	"os"
	"strconv"
)

func LoadData(filePath string) []Review {
	filePath = "../" + filePath
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	var reviews []Review
	for _, record := range records[1:] { // Assuming first row is header
		stars, _ := strconv.Atoi(record[3]) // Assuming 'stars' is at index 3
		review := Review{
			ReviewID:   record[0],
			ProductID:  record[1],
			ReviewerID: record[2],
			Stars:      stars,
			// Populate other fields if needed
		}
		reviews = append(reviews, review)
	}
	return reviews
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

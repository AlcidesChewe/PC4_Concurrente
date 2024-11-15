package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"sync"

	"recomendador/config"
	"recomendador/utils"
)

var (
	partitions        [][]utils.Review    // Data partitions
	partitionIndex    int                 // Index for next partition
	partitionMutex    sync.Mutex          // Mutex for partition index
	aggregatedResults []utils.ResultData  // Collected client results
	resultsMutex      sync.Mutex          // Mutex for aggregated results
	recommendations   map[string][]string // Map of categories to recommendations
	recommendationsMu sync.RWMutex        // Mutex for recommendations map
)

func main() {
	// Load server configuration
	cfg := config.LoadServerConfig()

	// Load and partition dataset
	data := utils.LoadData(cfg.Dataset.Path)
	partitions = utils.SplitData(data, cfg.Dataset.Partitions)
	partitionIndex = 0

	var wg sync.WaitGroup

	// Start the TCP server for client communication
	wg.Add(1)
	go func() {
		defer wg.Done()
		startTCPServer(cfg)
	}()

	// Wait for all clients to finish processing and recommendations to be ready
	wg.Wait()
	processAggregatedResults()

	// Start the CLI for user interaction
	startCLI()
}

func startTCPServer(cfg config.ServerConfig) {
	// Start listening on TCP port
	ln, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("TCP Server listening on port", cfg.Server.Port)

	var wg sync.WaitGroup

	expectedClients := cfg.Server.MaxClients
	for i := 0; i < expectedClients; i++ {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		wg.Add(1)
		go handleClient(conn, &wg)
	}

	// All expected clients have connected
	wg.Wait()
	// ????? processAggregatedResults()
}

func startCLI() {
	fmt.Println(
		"Recommendations are ready. You can now enter product categories to get recommendations.",
	)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter product category (or type 'exit' to quit): ")
		if !scanner.Scan() {
			break
		}
		category := scanner.Text()
		if category == "exit" {
			fmt.Println("Exiting CLI.")
			break
		}
		displayRecommendations(category)
	}
}

func displayRecommendations(category string) {
	recommendationsMu.RLock()
	defer recommendationsMu.RUnlock()
	if recommendations == nil {
		fmt.Println("Recommendations are not ready yet. Please try again later.")
		return
	}

	recs, exists := recommendations[category]
	if !exists || len(recs) == 0 {
		fmt.Printf("No recommendations found for the category '%s'.\n", category)
		return
	}

	// Limit to top 10 recommendations
	limit := 10
	if len(recs) < limit {
		limit = len(recs)
	}
	topRecs := recs[:limit]

	// Display recommendations
	fmt.Printf("Top %d recommendations for category '%s':\n", limit, category)
	for i, productID := range topRecs {
		fmt.Printf("%d. %s\n", i+1, productID)
	}
}

// handleClient manages the communication with a client

func handleClient(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client connected: %s\n", clientAddr)

	partition := getNextPartition()
	encoder := json.NewEncoder(conn)

	if partition == nil {
		// No more partitions to assign
		fmt.Println("No more partitions to assign to client")
		noWorkMessage := utils.ServerMessage{Message: "NO_MORE_WORK"}
		err := encoder.Encode(noWorkMessage)
		if err != nil {
			fmt.Printf("Error sending NO_MORE_WORK message to client %s: %v\n", clientAddr, err)
		}
		return
	}

	// Send partition data to client
	serverMessage := utils.ServerMessage{Partition: partition}
	err := encoder.Encode(serverMessage)
	if err != nil {
		fmt.Printf("Error sending data to client %s: %v\n", clientAddr, err)
		return
	}
	fmt.Printf("Sent partition data to client %s\n", clientAddr)

	// Receive results from client
	decoder := json.NewDecoder(conn)
	var results utils.ResultData
	err = decoder.Decode(&results)
	if err != nil {
		fmt.Printf("Error decoding client results from %s: %v\n", clientAddr, err)
		return
	}
	fmt.Printf("Received results from client %s\n", clientAddr)

	// Aggregate results
	aggregateResults(results)
}

// getNextPartition assigns the next available partition to a client
func getNextPartition() []utils.Review {
	partitionMutex.Lock()
	defer partitionMutex.Unlock()

	if partitionIndex >= len(partitions) {
		return nil
	}

	partition := partitions[partitionIndex]
	partitionIndex++
	return partition
}

// aggregateResults collects results from clients
func aggregateResults(results utils.ResultData) {
	resultsMutex.Lock()
	defer resultsMutex.Unlock()

	aggregatedResults = append(aggregatedResults, results)
}

// processAggregatedResults processes the final recommendations
func processAggregatedResults() {
	fmt.Println("Processing aggregated results...")

	combinedRecommendations := make(map[string]map[string]int) // Map[category]Map[productID]score

	for _, result := range aggregatedResults {
		for _, recs := range result.Recommendations {
			for _, productID := range recs {
				category := utils.GetProductCategory(productID)
				if category == "" {
					continue
				}
				if _, exists := combinedRecommendations[category]; !exists {
					combinedRecommendations[category] = make(map[string]int)
				}
				combinedRecommendations[category][productID] += 1 // Increase score
			}
		}
	}

	// Convert map to slice and store in recommendations variable
	recommendationsMu.Lock()
	recommendations = make(map[string][]string)
	for category, productsMap := range combinedRecommendations {
		// Create a slice of product-score pairs
		type productScore struct {
			ProductID string
			Score     int
		}
		var productScores []productScore
		for productID, score := range productsMap {
			productScores = append(productScores, productScore{ProductID: productID, Score: score})
		}
		// Sort products by score descending
		sort.Slice(productScores, func(i, j int) bool {
			return productScores[i].Score > productScores[j].Score
		})
		// Extract sorted product IDs
		products := make([]string, len(productScores))
		for i, ps := range productScores {
			products[i] = ps.ProductID
		}
		recommendations[category] = products
	}
	recommendationsMu.Unlock()

	fmt.Println("Recommendations processing completed.")
}

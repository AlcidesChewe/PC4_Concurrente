package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"recomendador/config"
	"recomendador/utils"
)

var (
	partitions        [][]utils.Review   // Data partitions
	partitionIndex    int                // Index for next partition
	partitionMutex    sync.Mutex         // Mutex for partition index
	aggregatedResults []utils.ResultData // Collected client results
	resultsMutex      sync.Mutex         // Mutex for aggregated results
)

func main() {
	// Load server configuration
	cfg := config.LoadServerConfig()

	// Start listening on TCP port
	ln, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Server listening on port", cfg.Server.Port)

	// Load and partition dataset
	data := utils.LoadData(cfg.Dataset.Path)
	partitions = utils.SplitData(data, cfg.Dataset.Partitions)
	partitionIndex = 0

	var wg sync.WaitGroup

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		wg.Add(1)
		go handleClient(conn, &wg)
	}

	wg.Wait()
	processAggregatedResults()
}

// handleClient manages the communication with a client
func handleClient(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	partition := getNextPartition()
	if partition == nil {
		// No more partitions to assign
		fmt.Println("No more partitions to assign to client")
		noWorkMessage := map[string]string{"message": "NO_MORE_WORK"}
		messageBytes, _ := json.Marshal(noWorkMessage)
		conn.Write(messageBytes)
		return
	}

	// Send partition data to client
	dataBytes, err := json.Marshal(partition)
	if err != nil {
		fmt.Println("Error marshalling partition data:", err)
		return
	}
	_, err = conn.Write(dataBytes)
	if err != nil {
		fmt.Println("Error sending data to client:", err)
		return
	}

	// Receive results from client
	resultBytes := make([]byte, 65536) // Adjust buffer size if needed
	n, err := conn.Read(resultBytes)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}

	var results utils.ResultData
	err = json.Unmarshal(resultBytes[:n], &results)
	if err != nil {
		fmt.Println("Error unmarshalling client results:", err)
		return
	}

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

	combinedRecommendations := make(map[string][]string)

	for _, result := range aggregatedResults {
		for userID, recommendations := range result.Recommendations {
			combinedRecommendations[userID] = append(
				combinedRecommendations[userID],
				recommendations...)
		}
	}

	// Optionally, process combinedRecommendations further
	fmt.Printf("Final Combined Recommendations: %+v\n", combinedRecommendations)
}

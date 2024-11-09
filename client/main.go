package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"recomendador/config"
	"recomendador/utils"
)

func main() {
	// Load client configuration
	cfg := config.LoadClientConfig()

	// Connect to the server
	conn, err := net.Dial("tcp", cfg.Server.Address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Receive data from server
	reader := bufio.NewReader(conn)
	dataBytes, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}

	// Check for "NO_MORE_WORK" message
	var message map[string]string
	err = json.Unmarshal(dataBytes, &message)
	if err == nil && message["message"] == "NO_MORE_WORK" {
		fmt.Println("No more work assigned by server")
		return
	}

	// Unmarshal partition data
	var partition []utils.Review
	err = json.Unmarshal(dataBytes, &partition)
	if err != nil {
		fmt.Println("Error unmarshalling partition data:", err)
		return
	}

	// Perform computation
	results := utils.PerformComputation(partition)

	// Send results back to server
	resultBytes, err := json.Marshal(results)
	if err != nil {
		fmt.Println("Error marshalling results:", err)
		return
	}
	conn.Write(resultBytes)
}

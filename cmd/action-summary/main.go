package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/dnd-it/action-summary/internal/summary"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "::error::%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	inputString := os.Getenv("INPUT_STRING")
	inputPath := os.Getenv("INPUT_PATH")
	maxSizeStr := getEnv("INPUT_MAX_SIZE", "1048576")
	summaryHeader := getEnv("INPUT_SUMMARY_HEADER", "Summary")
	dataType := os.Getenv("INPUT_DATA_TYPE")

	// Validate exactly one of string/path is provided
	if (inputString == "" && inputPath == "") || (inputString != "" && inputPath != "") {
		return fmt.Errorf("provide either string or path, but not both")
	}

	maxSize, err := strconv.Atoi(maxSizeStr)
	if err != nil {
		return fmt.Errorf("invalid max_size value: %w", err)
	}

	// Read input data
	var inputData string
	if inputPath != "" {
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("input file %s not found: %w", inputPath, err)
		}
		inputData = string(data)
	} else {
		inputData = inputString
	}

	// Validate non-empty
	if inputData == "" {
		return fmt.Errorf("input data is empty")
	}

	// Check size against max_size
	if len(inputData) > maxSize {
		fmt.Printf("::warning::String content too long (%d chars); exceeds %d byte limit.\n", len(inputData), maxSize)
		return nil
	}

	// Try to parse as JSON, recursively deserialize nested JSON strings, and pretty-print
	var parsed any
	if err := json.Unmarshal([]byte(inputData), &parsed); err == nil {
		deserialized := summary.DeserializeNestedJSON(parsed)
		prettyBytes, err := json.MarshalIndent(deserialized, "", "  ")
		if err == nil {
			inputData = string(prettyBytes)
		}
	}

	// Write formatted markdown to GITHUB_STEP_SUMMARY
	output := summary.FormatOutput(summaryHeader, dataType, inputData)

	summaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryPath == "" {
		fmt.Print(output)
		return nil
	}

	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_STEP_SUMMARY: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(output); err != nil {
		return fmt.Errorf("failed to write to GITHUB_STEP_SUMMARY: %w", err)
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

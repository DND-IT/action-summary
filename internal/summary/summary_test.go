package summary

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDeserializeNestedJSON_PlainString(t *testing.T) {
	result := DeserializeNestedJSON("hello world")
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %v", result)
	}
}

func TestDeserializeNestedJSON_Number(t *testing.T) {
	result := DeserializeNestedJSON(42.0)
	if result != 42.0 {
		t.Errorf("expected 42.0, got %v", result)
	}
}

func TestDeserializeNestedJSON_Bool(t *testing.T) {
	result := DeserializeNestedJSON(true)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestDeserializeNestedJSON_Nil(t *testing.T) {
	result := DeserializeNestedJSON(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestDeserializeNestedJSON_NestedJSONString(t *testing.T) {
	input := map[string]any{
		"key": `{"nested": "value"}`,
	}
	result := DeserializeNestedJSON(input)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	nested, ok := m["key"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map, got %T", m["key"])
	}
	if nested["nested"] != "value" {
		t.Errorf("expected 'value', got %v", nested["nested"])
	}
}

func TestDeserializeNestedJSON_DeeplyNested(t *testing.T) {
	// A string containing JSON that itself contains a JSON string
	innerJSON := `{"deep": "data"}`
	outerJSON, _ := json.Marshal(innerJSON)
	input := map[string]any{
		"key": string(outerJSON),
	}
	result := DeserializeNestedJSON(input)
	m := result.(map[string]any)
	nested, ok := m["key"].(map[string]any)
	if !ok {
		t.Fatalf("expected deeply nested map, got %T", m["key"])
	}
	if nested["deep"] != "data" {
		t.Errorf("expected 'data', got %v", nested["deep"])
	}
}

func TestDeserializeNestedJSON_Array(t *testing.T) {
	input := []any{"plain", `{"a": 1}`}
	result := DeserializeNestedJSON(input)
	arr, ok := result.([]any)
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if arr[0] != "plain" {
		t.Errorf("expected 'plain', got %v", arr[0])
	}
	nested, ok := arr[1].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map, got %T", arr[1])
	}
	if nested["a"] != 1.0 {
		t.Errorf("expected 1, got %v", nested["a"])
	}
}

func TestDeserializeNestedJSON_Dict(t *testing.T) {
	input := map[string]any{
		"plain": "text",
		"num":   42.0,
		"obj":   `{"inner": true}`,
	}
	result := DeserializeNestedJSON(input)
	m := result.(map[string]any)
	if m["plain"] != "text" {
		t.Errorf("expected 'text', got %v", m["plain"])
	}
	if m["num"] != 42.0 {
		t.Errorf("expected 42.0, got %v", m["num"])
	}
	inner, ok := m["obj"].(map[string]any)
	if !ok {
		t.Fatalf("expected map for obj, got %T", m["obj"])
	}
	if inner["inner"] != true {
		t.Errorf("expected true, got %v", inner["inner"])
	}
}

func TestFormatOutput_Basic(t *testing.T) {
	result := FormatOutput("Test Header", "json", `{"key": "value"}`)
	expected := "## Test Header\n<details><summary>Click to expand</summary>\n\n```json\n{\"key\": \"value\"}\n```\n</details>\n"
	if result != expected {
		t.Errorf("unexpected output:\n%s", result)
	}
}

func TestFormatOutput_EmptyDataType(t *testing.T) {
	result := FormatOutput("Summary", "", "plain text")
	expected := "## Summary\n<details><summary>Click to expand</summary>\n\n```\nplain text\n```\n</details>\n"
	if result != expected {
		t.Errorf("unexpected output:\n%s", result)
	}
}

func TestFormatOutput_CustomHeader(t *testing.T) {
	result := FormatOutput("My Custom Header", "yaml", "key: value")
	expected := "## My Custom Header\n"
	if len(result) < len(expected) || result[:len(expected)] != expected {
		t.Errorf("header not formatted correctly: %s", result)
	}
}

func TestIntegration_StringInput(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")

	t.Setenv("INPUT_STRING", `{"key": "value"}`)
	t.Setenv("INPUT_PATH", "")
	t.Setenv("INPUT_MAX_SIZE", "1048576")
	t.Setenv("INPUT_SUMMARY_HEADER", "Test")
	t.Setenv("INPUT_DATA_TYPE", "json")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	runMain(t, summaryFile)

	content, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("failed to read summary file: %v", err)
	}
	if len(content) == 0 {
		t.Error("summary file is empty")
	}
}

func TestIntegration_FileInput(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	inputFile := filepath.Join(t.TempDir(), "input.txt")

	if err := os.WriteFile(inputFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	t.Setenv("INPUT_STRING", "")
	t.Setenv("INPUT_PATH", inputFile)
	t.Setenv("INPUT_MAX_SIZE", "1048576")
	t.Setenv("INPUT_SUMMARY_HEADER", "File Test")
	t.Setenv("INPUT_DATA_TYPE", "")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	runMain(t, summaryFile)

	content, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("failed to read summary file: %v", err)
	}
	expected := "## File Test\n<details><summary>Click to expand</summary>\n\n```\nhello world\n```\n</details>\n"
	if string(content) != expected {
		t.Errorf("unexpected content:\n%s", string(content))
	}
}

func TestIntegration_NestedJSON(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")

	t.Setenv("INPUT_STRING", `{"key": "{\"nested\": \"value\"}"}`)
	t.Setenv("INPUT_PATH", "")
	t.Setenv("INPUT_MAX_SIZE", "1048576")
	t.Setenv("INPUT_SUMMARY_HEADER", "Nested")
	t.Setenv("INPUT_DATA_TYPE", "json")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	runMain(t, summaryFile)

	content, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("failed to read summary file: %v", err)
	}
	if len(content) == 0 {
		t.Error("summary file is empty")
	}
	// The nested JSON string should have been deserialized
	contentStr := string(content)
	if !strings.Contains(contentStr, "nested") || !strings.Contains(contentStr, "value") {
		t.Errorf("expected deserialized nested JSON in output:\n%s", contentStr)
	}
}

// runMain simulates the main logic for integration tests.
func runMain(t *testing.T, summaryFile string) {
	t.Helper()

	inputString := os.Getenv("INPUT_STRING")
	inputPath := os.Getenv("INPUT_PATH")
	maxSizeStr := os.Getenv("INPUT_MAX_SIZE")
	summaryHeader := os.Getenv("INPUT_SUMMARY_HEADER")
	dataType := os.Getenv("INPUT_DATA_TYPE")

	maxSize := 1048576
	if maxSizeStr != "" {
		var err error
		maxSize, err = strconv.Atoi(maxSizeStr)
		if err != nil {
			t.Fatalf("invalid max_size: %v", err)
		}
	}

	var inputData string
	if inputPath != "" {
		data, err := os.ReadFile(inputPath)
		if err != nil {
			t.Fatalf("failed to read input file: %v", err)
		}
		inputData = string(data)
	} else {
		inputData = inputString
	}

	if len(inputData) > maxSize {
		return
	}

	var parsed any
	if err := json.Unmarshal([]byte(inputData), &parsed); err == nil {
		deserialized := DeserializeNestedJSON(parsed)
		prettyBytes, _ := json.MarshalIndent(deserialized, "", "  ")
		inputData = string(prettyBytes)
	}

	output := FormatOutput(summaryHeader, dataType, inputData)
	if err := os.WriteFile(summaryFile, []byte(output), 0644); err != nil {
		t.Fatalf("failed to write summary: %v", err)
	}
}


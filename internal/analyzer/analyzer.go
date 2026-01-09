package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const geminiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

type Package struct {
	Name    string `json:"name"`
	Manager string `json:"manager"`
}

type AnalysisResult struct {
	Matcher  string    `json:"matcher"`
	TestOK   []string  `json:"test_ok"`
	TestFail []string  `json:"test_fail"`
	Packages []Package `json:"packages"`

	// Helper for the runner, not part of the final JSON
	InstallCommands []string `json:"install_commands"`
}

func Analyze(errorLog string) (*AnalysisResult, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	prompt := fmt.Sprintf(`
You are an expert software engineer debugging a build error.
The following error occurred when running a tool called 'credo' on a project located at '../core'.

Error Log:
%s

Analyze the error and identify the missing dependencies.
Produce a JSON object that defines a "matcher" for this error and lists the packages to install.
The JSON must follow this exact structure:
{
  "matcher": "regex string that uniquely identifies this error",
  "test_ok": ["string that matches the regex (e.g., the error line)"],
  "test_fail": ["string that should NOT match"],
  "packages": [
    { "name": "package_name", "manager": "package_manager_command (e.g., apt, brew, go, pip, apk)" }
  ],
  "install_commands": ["full command to install the dependencies now (e.g., 'apt-get install -y foo', 'go get bar')"]
}

Ensure the 'matcher' regex is robust but specific enough to catch this error type.
The 'packages' list should support multiple managers if applicable (e.g., usually apt or brew), but at least one is required.
`, errorLog)

	requestBody, err := json.Marshal(map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]interface{}{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"response_mime_type": "application/json",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", geminiURL+"?key="+apiKey, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	var result AnalysisResult
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from LLM: %w \nResponse: %s", err, responseText)
	}

	return &result, nil
}

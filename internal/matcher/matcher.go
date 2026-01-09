package matcher

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Rebel-Project-Core/AI-Harness/internal/analyzer"
)

type MatcherFile struct {
	Matcher  string             `json:"matcher"`
	TestOK   []string           `json:"test_ok"`
	TestFail []string           `json:"test_fail"`
	Packages []analyzer.Package `json:"packages"`
}

func Save(errorLog string, result *analyzer.AnalysisResult) error {
	outputDir := "../package-suggestions/matchers"
	
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create matcher directory: %w", err)
	}

	hash := sha256.Sum256([]byte(errorLog))
	filename := fmt.Sprintf("fix_%x.json", hash[:4])
	filePath := filepath.Join(outputDir, filename)

	data := MatcherFile{
		Matcher:  result.Matcher,
		TestOK:   result.TestOK,
		TestFail: result.TestFail,
		Packages: result.Packages,
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal matcher: %w", err)
	}

	if err := os.WriteFile(filePath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write matcher file: %w", err)
	}

	fmt.Printf("Matcher saved to %s\n", filePath)
	return nil
}

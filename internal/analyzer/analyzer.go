package analyzer

import (
	"fmt"
	"os"
)

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

func Analyze(errorLog string, modelName string) (*AnalysisResult, error) {
	prompt := constructPrompt(errorLog)

	if os.Getenv("GEMINI_API_KEY") != "" {
		return analyzeGemini(prompt, modelName)
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return analyzeOpenAI(prompt, modelName)
	}

	return nil, fmt.Errorf("no API key found. Please set GEMINI_API_KEY or OPENAI_API_KEY")
}

func constructPrompt(errorLog string) string {
	return fmt.Sprintf(`
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
}

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Rebel-Project-Core/AI-Harness/internal/analyzer"
	"github.com/Rebel-Project-Core/AI-Harness/internal/matcher"
	"github.com/Rebel-Project-Core/AI-Harness/internal/runner"
)

func main() {
	modelFlag := flag.String("model", "", "Model name to use (e.g. gpt-4o, gemini-1.5-pro)")
	flag.Parse()

	combinedLog, command, exitCode := getExecutionLog(flag.Args())

	// Analyze the error
	analysis, err := analyzer.Analyze(combinedLog, *modelFlag)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	printAnalysis(analysis)

	if len(analysis.Packages) == 0 {
		fmt.Println("No packages suggested. Exiting.")
		os.Exit(exitCode)
	}

	success := true
	if command != "" {
		success = retryInstallation(command, analysis.Packages)
	} else {
		fmt.Println("Skipping installation (no command provided).")
	}

	if success {
		// Save the successful match (or just the match if no verification was possible)
		err := matcher.Save(combinedLog, analysis)
		if err != nil {
			fmt.Printf("Warning: Failed to save matcher: %v\n", err)
		}
	} else {
		os.Exit(1)
	}
}

// getExecutionLog determines whether to read from stdin or run a command.
// Returns the log content, the command name (if any), and the exit code.
// Exits the program if the initial command succeeds or if usage is incorrect.
func getExecutionLog(args []string) (string, string, int) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Stdin is piped
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Failed to read from stdin: %v", err)
		}
		command := ""
		if len(args) > 0 {
			command = args[0]
		}
		return string(data), command, 1
	}

	if len(args) < 1 {
		fmt.Println("Usage: ai-harness [--model <model_name>] <command> [args...]")
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]

	fmt.Printf("--- Attempt 1: Running %s %v ---\n", command, commandArgs)
	result, err := runner.Run(command, commandArgs, nil)
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}

	if result.ExitCode == 0 {
		fmt.Println("Command executed successfully.")
		fmt.Print(result.Stdout)
		os.Exit(0)
	}

	fmt.Printf("Command failed (Exit Code %d). initiating AI analysis...\n", result.ExitCode)
	return result.Stderr + "\n" + result.Stdout, command, result.ExitCode
}

func printAnalysis(analysis *analyzer.AnalysisResult) {
	fmt.Println("--- AI Analysis Result ---")
	fmt.Printf("Matcher Regex: %s\n", analysis.Matcher)
	fmt.Printf("Suggested Packages: %v\n", analysis.Packages)
	fmt.Println("--------------------------")
}

func retryInstallation(command string, packages []analyzer.Package) bool {
	fmt.Printf("--- Attempt 2: Retrying installation with credo ---\n")

	for _, pkg := range packages {
		// Construct args: [manager, package_name]
		retryArgs := []string{pkg.Manager, pkg.Name}

		fmt.Printf("Command: %s %v\n", command, retryArgs)

		retryResult, err := runner.Run(command, retryArgs, nil)
		if err != nil {
			log.Fatalf("Retry execution error: %v", err)
		}

		if retryResult.ExitCode == 0 {
			fmt.Println("Installation Successful.")
			fmt.Print(retryResult.Stdout)
		} else {
			fmt.Println("Installation failed.")
			fmt.Println(retryResult.Stderr)
			return false
		}
	}
	return true
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Rebel-Project-Core/AI-Harness/internal/analyzer"
	"github.com/Rebel-Project-Core/AI-Harness/internal/matcher"
	"github.com/Rebel-Project-Core/AI-Harness/internal/runner"
)

func main() {
	modelFlag := flag.String("model", "", "Model name to use (e.g. gpt-4o, gemini-1.5-pro)")
	flag.Parse()

	args := flag.Args()
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
		return
	}

	fmt.Printf("Command failed (Exit Code %d). initiating AI analysis...\n", result.ExitCode)

	// Analyze the error
	combinedLog := result.Stderr + "\n" + result.Stdout
	analysis, err := analyzer.Analyze(combinedLog, *modelFlag)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	fmt.Println("--- AI Analysis Result ---")
	fmt.Printf("Matcher Regex: %s\n", analysis.Matcher)
	fmt.Printf("Suggested Packages: %v\n", analysis.Packages)
	fmt.Println("--------------------------")

	if len(analysis.Packages) == 0 {
		fmt.Println("No packages suggested. Exiting.")
		os.Exit(result.ExitCode)
	}

	// Retry installation WITH CREDO using suggested dependencies
	fmt.Printf("--- Attempt 2: Retrying installation with credo ---\n")

	allSuccess := true
	for _, pkg := range analysis.Packages {
		// Construct args: [manager, package_name]
		// This assumes credo accepts "manager package_name" (e.g., "pip Pillow")
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
			allSuccess = false
			break // Stop on first failure
		}
	}

	if allSuccess {
		// Save the successful match
		err := matcher.Save(combinedLog, analysis)
		if err != nil {
			fmt.Printf("Warning: Failed to save matcher: %v\n", err)
		}
	} else {
		os.Exit(1)
	}
}
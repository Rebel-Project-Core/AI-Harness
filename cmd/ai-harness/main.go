package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Rebel-Project-Core/AI-Harness/internal/analyzer"
	"github.com/Rebel-Project-Core/AI-Harness/internal/matcher"
	"github.com/Rebel-Project-Core/AI-Harness/internal/runner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ai-harness <command> [args...]")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	fmt.Printf("--- Attempt 1: Running %s %v ---\n", command, args)
	result, err := runner.Run(command, args, nil)
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
	analysis, err := analyzer.Analyze(combinedLog)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	fmt.Println("--- AI Analysis Result ---")
	fmt.Printf("Matcher Regex: %s\n", analysis.Matcher)
	fmt.Printf("Suggested Packages: %v\n", analysis.Packages)
	fmt.Printf("Install Commands: %v\n", analysis.InstallCommands)
	fmt.Println("--------------------------")

	if len(analysis.InstallCommands) == 0 {
		fmt.Println("No install commands suggested. Exiting.")
		os.Exit(result.ExitCode)
	}

	// Run install commands
	fmt.Println("--- Installing Dependencies ---")
	for _, cmdStr := range analysis.InstallCommands {
		// Split command string into parts (naive split by space)
		// Better approach: use a shell execution if the command is complex
		// For simplicity, we'll try to execute it via 'sh -c' to handle arguments correctly
		fmt.Printf("Executing: %s\n", cmdStr)
		installRes, err := runner.Run("sh", []string{"-c", cmdStr}, nil)
		if err != nil {
			log.Fatalf("Failed to run install command '%s': %v", cmdStr, err)
		}
		if installRes.ExitCode != 0 {
			fmt.Printf("Install command failed: %s\n", installRes.Stderr)
			os.Exit(installRes.ExitCode)
		}
		fmt.Println("Success.")
	}

	// Retry original command
	fmt.Printf("--- Attempt 2: Retrying original command ---\n")
	fmt.Printf("Command: %s %v\n", command, args)

	retryResult, err := runner.Run(command, args, nil)
	if err != nil {
		log.Fatalf("Retry execution error: %v", err)
	}

	if retryResult.ExitCode == 0 {
		fmt.Println("Retry Successful!")
		fmt.Print(retryResult.Stdout)

		// Save the successful match
		err := matcher.Save(combinedLog, analysis)
		if err != nil {
			fmt.Printf("Warning: Failed to save matcher: %v\n", err)
		}
	} else {
		fmt.Println("Retry failed.")
		fmt.Println(retryResult.Stderr)
		os.Exit(retryResult.ExitCode)
	}
}


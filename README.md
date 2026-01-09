# AI-Harness for Credo

AI-Harness is a smart wrapper tool designed to automate the resolution of dependency errors when running `credo`. It executes a command, detects failures, uses the Gemini API to identify missing dependencies, installs them, and retries the execution. Upon success, it generates a "matcher" file that can be used by other tools to recognize and fix similar errors in the future.

## Features

*   **Automated Error Analysis**: Uses Gemini to parse error logs and identify missing system or language-specific dependencies.
*   **Auto-Remediation**: Executes the suggested installation commands (e.g., `go get`, `apt-get install`, `pip install`) and retries the original command.
*   **Pattern Learning**: Saves successful fixes as structured JSON "matcher" files in `../package-suggestions/matchers`, compatible with the Credo ecosystem.

## Prerequisites

*   Go 1.21+
*   A valid [Google Gemini API Key](https://ai.google.dev/)

## Installation

1.  Clone the repository (if you haven't already).
2.  Build the project using `make`:

```bash
make build
```

This will create the `ai-harness` binary in the root directory.

## Usage

1.  Set your Gemini API key:

```bash
export GEMINI_API_KEY="your_api_key_here"
```

2.  Run your command through the harness:

```bash
./ai-harness <command> [arguments...]
```

**Example:**

```bash
./ai-harness credo @../core
```

### How it Works

1.  **Execution**: Runs the provided command.
2.  **Failure Detection**: If the command exits with a non-zero status, it captures `stdout` and `stderr`.
3.  **Analysis**: Sends the logs to Gemini to extract:
    *   A regex to match the error.
    *   The missing package(s).
    *   The package manager (e.g., `apt`, `brew`, `go`).
    *   Immediate installation commands.
4.  **Remediation**: Runs the installation commands.
5.  **Retry**: Re-runs the original command.
6.  **Learning**: If the retry succeeds, saves a JSON matcher file to `../package-suggestions/matchers/` so the fix is recorded for future use.

## Development

*   **Build**: `make build`
*   **Test**: `make test`
*   **Lint**: `make vet`
*   **Clean**: `make clean`

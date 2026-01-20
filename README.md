# AI-Harness for Credo

AI-Harness is a smart wrapper tool designed to automate the resolution of dependency errors when running `credo`. It executes a command, detects failures, uses an LLM (Gemini or OpenAI) to identify missing dependencies, installs them, and retries the execution. Upon success, it generates a "matcher" file that can be used by other tools to recognize and fix similar errors in the future.

## Features

*   **Automated Error Analysis**: Uses LLMs to parse error logs and identify missing system or language-specific dependencies.
*   **Multi-Model Support**: Supports both Google Gemini and OpenAI ChatGPT.
*   **Auto-Remediation**: Executes the suggested installation commands (e.g. `apt-get install`, `pip install`) and retries the original command.
*   **Pattern Learning**: Saves successful fixes as structured JSON "matcher" files in `./matchers`, compatible with the Credo ecosystem.

## Prerequisites

*   Go 1.21+
*   A valid [Google Gemini API Key](https://ai.google.dev/) OR [OpenAI API Key](https://platform.openai.com/)

## Installation

1.  Clone the repository (if you haven't already).
2.  Build the project using `make`:

```bash
make build
```

This will create the `ai-harness` binary in the root directory.

## Usage

1.  Set your API key. AI-Harness supports both Gemini and OpenAI.

**For Gemini (Default):**
```bash
export GEMINI_API_KEY="your_api_key_here"
```

**For OpenAI:**
```bash
export OPENAI_API_KEY="your_api_key_here"
```

**Precedence:**
If both keys are set, `GEMINI_API_KEY` takes precedence.

2.  Run your command through the harness:

```bash
./ai-harness [--model <model_name>] <command> [arguments...]
```

**Example:**

```bash
./ai-harness credo {COMMAND}
```

**Example with custom model:**

```bash
# Use Gemini Pro
./ai-harness --model gemini-1.5-pro credo {COMMAND}

# Use GPT-4 Turbo
./ai-harness --model gpt-4-turbo credo {COMMAND}
```

### How it Works

1.  **Execution**: Runs the provided command.
2.  **Failure Detection**: If the command exits with a non-zero status, it captures `stdout` and `stderr`.
3.  **Analysis**: Sends the logs to the selected LLM to extract:
    *   A regex to match the error.
    *   The missing package(s).
    *   The package manager (e.g., `apt`, `pip`, `cran`).
    *   Immediate installation commands.
4.  **Remediation**: Runs the installation commands.
5.  **Retry**: Re-runs the original command with the new suggestions.
6.  **Learning**: If the retry succeeds, saves a JSON matcher file to `../package-suggestions/matchers/` so the fix is recorded for future use.

## Development

*   **Build**: `make build`
*   **Test**: `make test`
*   **Lint**: `make vet`
*   **Clean**: `make clean`

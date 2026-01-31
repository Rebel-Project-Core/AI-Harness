#!/usr/bin/env python3
import argparse
import os
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

# Import log_analyzer classes
try:
    from log_analyzer import LogReader, TFIDFVectorizer, SearchEngine
except ImportError:
    print("Error: log_analyzer.py not found. Make sure it is in the same directory.")
    sys.exit(1)

IMAGE = "ghcr.io/rebel-project-core/ai-harness:latest"
MAX_LOG_BYTES = 32 * 1024


def run_analyzer(log_file, query="error failure exception traceback", k=5):
    """
    Runs the log analyzer on the given log file and returns the combined text of relevant chunks.
    """
    try:
        reader = LogReader(chunk_size=10)
        chunks = reader.read_chunks(str(log_file))

        engine = SearchEngine(TFIDFVectorizer())
        engine.index(chunks)

        results = engine.search(query, k)

        # Sort by index to maintain chronological order of the log
        results.sort(key=lambda x: x[1])

        output_parts = []
        for score, index, content in results:
            if score > 0:
                output_parts.append(
                    f"--- Chunk {index} (Score: {score:.4f}) ---\n{content.strip()}"
                )

        return "\n\n".join(output_parts)
    except Exception as e:
        print(f"Warning: Analyzer failed for {log_file}: {e}")
        return None


def process_log(
    log_file, output_dir, new_logs_dir, gemini_key, openai_key, use_analyzer=False
):
    # Create a sub-directory for this specific log's matchers
    log_output_dir = output_dir / log_file.stem
    log_output_dir.mkdir(exist_ok=True)
    abs_output = log_output_dir.resolve()

    cmd = [
        "docker",
        "run",
        "--rm",
        "-i",
        "-v",
        f"{abs_output}:/app/matchers",
        "-w",
        "/app",
    ]

    if gemini_key:
        cmd.extend(["-e", f"GEMINI_API_KEY={gemini_key}"])
    if openai_key:
        cmd.extend(["-e", f"OPENAI_API_KEY={openai_key}"])

    # The command is "credo". The harness will analyze the log from stdin
    # and execute "credo <manager> <package>" for remediation.
    cmd.extend([IMAGE, "-matchers", "/app/matchers", "credo"])

    new_log_file = new_logs_dir / log_file.name

    try:
        input_text = None
        use_stdin = False

        if use_analyzer:
            # print(f"Analyzing {log_file.name}...") # Optional: reduce verbosity
            analyzed_text = run_analyzer(log_file)
            if analyzed_text:
                input_text = analyzed_text
            else:
                # Fallback if analyzer finds nothing or fails
                # print(f"Analyzer found no matches for {log_file.name}. Falling back.")
                pass

        # If analyzer was not used or returned no results, fall back to default logic
        if input_text is None:
            file_size = log_file.stat().st_size
            if file_size > MAX_LOG_BYTES:
                head_size = 2 * 1024
                tail_size = 30 * 1024

                with open(log_file, "rb") as f_in:
                    head_content = f_in.read(head_size).decode("utf-8", errors="ignore")
                    f_in.seek(file_size - tail_size)
                    tail_content = f_in.read().decode("utf-8", errors="ignore")

                # Combine head and tail with a marker
                input_text = f"{head_content}\n...[TRUNCATED_BY_HARNESS_PREPROCESSOR]...\n{tail_content}"
            else:
                use_stdin = True

        if use_stdin:
            with open(log_file, "rb") as f_in:
                result = subprocess.run(cmd, stdin=f_in, capture_output=True, text=True)
        else:
            result = subprocess.run(
                cmd, input=input_text, capture_output=True, text=True
            )

        with open(new_log_file, "w") as f_out:
            f_out.write(result.stdout)
            if result.stderr:
                f_out.write("\n--- STDERR ---\n")
                f_out.write(result.stderr)

        if result.returncode == 0:
            print(f"SUCCESS: {log_file.name} -> {new_log_file}")
        else:
            print(
                f"FAILURE: {log_file.name} (Exit code {result.returncode}) -> {new_log_file}"
            )

    except Exception as e:
        print(f"ERROR: {log_file.name}: {e}")


def main():
    parser = argparse.ArgumentParser(
        description="Run ai-harness on a directory of logs."
    )
    parser.add_argument("log_dir", type=Path, help="Directory containing log files")
    parser.add_argument(
        "--analyze",
        action="store_true",
        help="Use log_analyzer to preprocess logs and extract relevant chunks",
    )
    args = parser.parse_args()

    if not args.log_dir.exists() or not args.log_dir.is_dir():
        print(f"Error: Directory '{args.log_dir}' does not exist.")
        sys.exit(1)

    output_dir = Path("output")
    output_dir.mkdir(exist_ok=True)

    new_logs_dir = Path("new_logs")
    new_logs_dir.mkdir(exist_ok=True)

    gemini_key = os.environ.get("GEMINI_API_KEY")
    openai_key = os.environ.get("OPENAI_API_KEY")

    if not gemini_key and not openai_key:
        print("Warning: Neither GEMINI_API_KEY nor OPENAI_API_KEY is set.")

    files = [f for f in args.log_dir.iterdir() if f.is_file()]

    if not files:
        print("No files found in log directory.")
        return

    print(
        f"Processing {len(files)} logs with 1 workers (Analyzer: {'Enabled' if args.analyze else 'Disabled'})..."
    )

    with ThreadPoolExecutor(max_workers=1) as executor:
        futures = [
            executor.submit(
                process_log,
                f,
                output_dir,
                new_logs_dir,
                gemini_key,
                openai_key,
                args.analyze,
            )
            for f in files
        ]

        for future in futures:
            future.result()


if __name__ == "__main__":
    main()


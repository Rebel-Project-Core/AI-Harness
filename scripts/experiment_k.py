#!/usr/bin/env python3
import argparse
import sys
import statistics
import concurrent.futures
import matplotlib.pyplot as plt
from pathlib import Path

sys.path.append(str(Path(__file__).parent.parent))

try:
    import log_analyzer
except ImportError:
    print("Error: log_analyzer.py not found in parent directory.")
    sys.exit(1)


def process_single_file(file_path, query, max_k):
    try:
        reader = log_analyzer.LogReader()
        chunks = reader.read_chunks(str(file_path))
        if not chunks:
            return None

        engine = log_analyzer.SearchEngine(log_analyzer.TFIDFVectorizer())
        engine.index(chunks)

        results = engine.search(query, max_k)

        file_scores = []
        for i in range(max_k):
            if i < len(results):
                score = results[i][0]
            else:
                score = 0.0

            file_scores.append(score)
        return file_scores
    except Exception as e:
        return e


def analyze_logs(log_dir, query, max_k):
    files = sorted([f for f in log_dir.iterdir() if f.is_file()])

    scores_by_rank = {i: [] for i in range(max_k)}
    all_file_scores = []
    file_count = 0

    print(f"Analyzing {len(files)} log files up to Rank {max_k}...")

    with concurrent.futures.ProcessPoolExecutor() as executor:
        future_to_file = {
            executor.submit(process_single_file, f, query, max_k): f for f in files
        }

        for future in concurrent.futures.as_completed(future_to_file):
            file_path = future_to_file[future]
            try:
                result = future.result()
                if result is None:
                    continue

                if isinstance(result, Exception):
                    print(f"Error processing {file_path}: {result}")
                    continue

                file_scores = result
                all_file_scores.append(file_scores)

                for i, score in enumerate(file_scores):
                    scores_by_rank[i].append(score)

                file_count += 1

                if file_count % 10 == 0:
                    print(f"Processed {file_count} files...", end="\r")

            except Exception as e:
                print(f"Error processing {file_path}: {e}")

    print(f"\nSuccessfully processed {file_count} files.\n")
    return scores_by_rank, all_file_scores, file_count


def calculate_stats(scores_by_rank, max_k, score_threshold):
    print(
        f"{'Rank':<5} | {'Mean Score':<12} | {'Median Score':<12} | {'90th %ile':<12} | {'% Non-Zero':<12}"
    )
    print("-" * 65)

    suggested_k = max_k
    found_cutoff = False

    ranks = []
    means = []
    medians = []
    p90s = []
    pct_non_zeros = []

    for i in range(max_k):
        scores = scores_by_rank[i]
        if not scores:
            mean_val = 0.0
            median_val = 0.0
            p90_val = 0.0
            pct_non_zero = 0.0
        else:
            mean_val = statistics.mean(scores)
            median_val = statistics.median(scores)

            scores_sorted = sorted(scores)
            idx_90 = int(0.90 * len(scores))
            p90_val = scores_sorted[idx_90]

            non_zero_count = sum(1 for s in scores if s > 0.0)
            pct_non_zero = (non_zero_count / len(scores)) * 100

        ranks.append(i + 1)
        means.append(mean_val)
        medians.append(median_val)
        p90s.append(p90_val)
        pct_non_zeros.append(pct_non_zero)

        print(
            f"{i+1:<5} | {mean_val:.4f}       | {median_val:.4f}       | {p90_val:.4f}       | {pct_non_zero:.1f}%"
        )

        if not found_cutoff:
            if p90_val < score_threshold:
                suggested_k = i
                if suggested_k < 1:
                    suggested_k = 1
                found_cutoff = True

    print("-" * 65)
    print(f"Optimal K (Statistical): {suggested_k}")
    print(f"(Based on 90th Percentile Score dropping below {score_threshold})")

    return suggested_k, ranks, means, medians, p90s, pct_non_zeros


def plot_results(
    ranks,
    means,
    medians,
    p90s,
    pct_non_zeros,
    scores_by_rank,
    all_file_scores,
    suggested_k,
    score_threshold,
    output_file,
):
    plt.figure(figsize=(18, 18))

    plt.subplot(3, 2, 1)
    plt.plot(ranks, means, label="Mean", marker="o", linewidth=2)
    plt.plot(ranks, medians, label="Median", marker="s", linestyle="--", alpha=0.7)
    plt.plot(ranks, p90s, label="90th %ile", marker="^", linewidth=2, color="purple")
    plt.axvline(
        x=suggested_k,
        color="r",
        linestyle=":",
        label=f"Optimal K={suggested_k}",
        linewidth=2,
    )
    plt.axhline(
        y=score_threshold,
        color="gray",
        linestyle="-.",
        label=f"Threshold={score_threshold}",
    )
    plt.title("Score Decay & Regime Boundaries")
    plt.xlabel("Rank")
    plt.ylabel("Score")
    plt.grid(True, linestyle="--", alpha=0.6)
    plt.legend()

    plt.subplot(3, 2, 2)
    plt.plot(ranks, pct_non_zeros, color="green", marker="o", linewidth=2)
    plt.axvline(
        x=suggested_k,
        color="r",
        linestyle=":",
        label=f"Optimal K={suggested_k}",
        linewidth=2,
    )
    plt.title("Coverage: % Non-Zero Matches")
    plt.xlabel("Rank")
    plt.ylabel("% Files with Score > 0")
    plt.grid(True, linestyle="--", alpha=0.6)
    plt.legend()

    plt.subplot(3, 2, 3)
    limit_box = min(15, len(scores_by_rank))
    box_data = [scores_by_rank[i] for i in range(limit_box)]
    plt.boxplot(
        box_data, tick_labels=[str(i + 1) for i in range(limit_box)], showfliers=False
    )
    plt.title(f"Score Variance (Ranks 1-{limit_box}, outliers hidden)")
    plt.xlabel("Rank")
    plt.ylabel("Score Distribution")
    plt.grid(axis="y", linestyle="--", alpha=0.6)

    cum_means = []
    total = 0
    for m in means:
        total += m
        cum_means.append(total)

    plt.subplot(3, 2, 4)
    plt.plot(ranks, cum_means, color="orange", marker="d", linewidth=2)
    plt.axvline(
        x=suggested_k, color="r", linestyle=":", label=f"Optimal K={suggested_k}"
    )
    plt.title("Cumulative Utility (Sum of Means)")
    plt.xlabel("Rank")
    plt.ylabel("Cumulative Mean Score")
    plt.grid(True, linestyle="--", alpha=0.6)
    plt.legend()

    plt.subplot(3, 1, 3)

    all_file_scores.sort(key=lambda x: sum(x), reverse=True)

    plt.imshow(all_file_scores, aspect="auto", cmap="viridis", interpolation="nearest")
    plt.colorbar(label="Relevance Score")
    plt.title(
        f"Relevance Heatmap (Sorted by Total Relevance) - {len(all_file_scores)} Files"
    )
    plt.xlabel(f"Rank (0-{len(ranks)-1})")
    plt.ylabel("Files (Sorted)")
    plt.axvline(
        x=suggested_k - 1,
        color="red",
        linestyle=":",
        linewidth=2,
        label=f"K={suggested_k}",
    )
    plt.legend(loc="upper right")

    plt.tight_layout()
    plt.savefig(output_file)
    print(f"Comprehensive visualization saved to '{output_file}'")


def main():
    parser = argparse.ArgumentParser(
        description="Analyze log files to determine optimal K for retrieval."
    )
    parser.add_argument(
        "--log-dir",
        type=Path,
        help="Directory containing log files",
    )
    parser.add_argument(
        "--query",
        type=str,
        default="error failure exception traceback",
        help="Search query to use for ranking",
    )
    parser.add_argument(
        "--max-k", type=int, default=30, help="Maximum rank K to analyze"
    )
    parser.add_argument(
        "--threshold",
        type=float,
        default=0.05,
        help="Score threshold for relevance (90th percentile)",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path("experiment_k.png"),
        help="Output path for the generated plot",
    )

    args = parser.parse_args()

    if not args.log_dir.exists() or not args.log_dir.is_dir():
        print(f"Error: Directory '{args.log_dir}' does not exist.")
        sys.exit(1)

    scores_by_rank, all_file_scores, file_count = analyze_logs(
        args.log_dir, args.query, args.max_k
    )

    if file_count == 0:
        print("No files processed.")
        sys.exit(0)

    suggested_k, ranks, means, medians, p90s, pct_non_zeros = calculate_stats(
        scores_by_rank, args.max_k, args.threshold
    )

    plot_results(
        ranks,
        means,
        medians,
        p90s,
        pct_non_zeros,
        scores_by_rank,
        all_file_scores,
        suggested_k,
        args.threshold,
        args.output,
    )


if __name__ == "__main__":
    main()

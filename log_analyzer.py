import sys
import re
import math
import collections
import argparse
from typing import List, Dict, Tuple, Iterable


class Tokenizer:
    @staticmethod
    def tokenize(text: str) -> List[str]:
        return re.findall(r"\b\w+\b", text.lower())


class LogReader:
    def __init__(self, chunk_size: int = 10):
        self.chunk_size = chunk_size

    def read_chunks(self, file_path: str) -> List[str]:
        try:
            with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
                lines = f.readlines()
            return [
                "".join(lines[i : i + self.chunk_size])
                for i in range(0, len(lines), self.chunk_size)
            ]
        except Exception as e:
            sys.stderr.write(f"Error reading file: {e}\n")
            sys.exit(1)


class TFIDFVectorizer:
    def __init__(self):
        self.idf: Dict[str, float] = {}
        self.vocab: set = set()
        self.doc_count = 0

    def fit(self, documents: Iterable[str]):
        self.doc_count = len(documents)
        doc_freqs = collections.defaultdict(int)
        for doc in documents:
            tokens = set(Tokenizer.tokenize(doc))
            for token in tokens:
                doc_freqs[token] += 1
                self.vocab.add(token)
        for token, freq in doc_freqs.items():
            self.idf[token] = math.log(self.doc_count / (1 + freq)) + 1

    def transform(self, document: str) -> Dict[str, float]:
        tokens = Tokenizer.tokenize(document)
        if not tokens:
            return {}
        tf = collections.Counter(tokens)
        total_tokens = len(tokens)
        vector = {}
        for token, count in tf.items():
            if token in self.vocab:
                vector[token] = (count / total_tokens) * self.idf.get(token, 0)
        return vector


class CosineSimilarity:
    @staticmethod
    def compute(vec1: Dict[str, float], vec2: Dict[str, float]) -> float:
        intersection = set(vec1.keys()) & set(vec2.keys())
        numerator = sum([vec1[x] * vec2[x] for x in intersection])
        sum1 = sum([vec1[x] ** 2 for x in vec1])
        sum2 = sum([vec2[x] ** 2 for x in vec2])
        denominator = math.sqrt(sum1) * math.sqrt(sum2)
        return numerator / denominator if denominator else 0.0


class SearchEngine:
    def __init__(self, vectorizer: TFIDFVectorizer):
        self.vectorizer = vectorizer
        self.documents: List[str] = []
        self.vectors: List[Dict[str, float]] = []

    def index(self, documents: List[str]):
        self.documents = documents
        self.vectorizer.fit(documents)
        self.vectors = [self.vectorizer.transform(doc) for doc in documents]

    def search(self, query: str, top_k: int) -> List[Tuple[float, int, str]]:
        query_tokens = Tokenizer.tokenize(query)
        query_vec = {}
        q_tf = collections.Counter(query_tokens)
        for token, count in q_tf.items():
            if token in self.vectorizer.idf:
                query_vec[token] = (count / len(query_tokens)) * self.vectorizer.idf[
                    token
                ]

        scores = []
        for i, doc_vec in enumerate(self.vectors):
            score = CosineSimilarity.compute(query_vec, doc_vec)
            scores.append((score, i, self.documents[i]))

        return sorted(scores, key=lambda x: x[0], reverse=True)[:top_k]


class Application:
    def run(self):
        parser = argparse.ArgumentParser()
        parser.add_argument("logfile")
        parser.add_argument(
            "query", nargs="?", default="error failure exception traceback"
        )
        parser.add_argument("-k", type=int, default=5)
        args = parser.parse_args()

        reader = LogReader(chunk_size=10)
        chunks = reader.read_chunks(args.logfile)

        engine = SearchEngine(TFIDFVectorizer())
        engine.index(chunks)

        results = engine.search(args.query, args.k)

        print(f"Top {args.k} matches for '{args.query}':")
        for score, index, content in results:
            if score > 0:
                print(f"\n--- Chunk {index} (Score: {score:.4f}) ---")
                print(content.strip())


if __name__ == "__main__":
    Application().run()

#!/usr/bin/env python3

import argparse
import concurrent.futures
import random
import time
import urllib.error
import urllib.request


def normalize_base_url(base_url: str) -> str:
    return base_url.rstrip("/")


def build_request_plan(base_url: str, min_per_second: int, max_per_second: int, rng: random.Random):
    if min_per_second <= 0:
        raise ValueError("min_per_second must be positive")
    if max_per_second < min_per_second:
        raise ValueError("max_per_second must be greater than or equal to min_per_second")

    base = normalize_base_url(base_url)
    count = rng.randint(min_per_second, max_per_second)
    paths = ["/work", "/checkout"]
    urls = [base + rng.choice(paths) for _ in range(count)]
    return count, urls


def request_once(url: str, timeout_seconds: float):
    started_at = time.perf_counter()
    try:
        with urllib.request.urlopen(url, timeout=timeout_seconds) as response:
            response.read()
            return url, response.status, time.perf_counter() - started_at, None
    except urllib.error.HTTPError as exc:
        exc.read()
        return url, exc.code, time.perf_counter() - started_at, str(exc)
    except Exception as exc:  # pragma: no cover - exercised during real runs
        return url, None, time.perf_counter() - started_at, str(exc)


def parse_args():
    parser = argparse.ArgumentParser(
        description="Generate random local traffic for /work and /checkout."
    )
    parser.add_argument("--base-url", default="http://localhost:8080", help="Base URL for the local app.")
    parser.add_argument("--min-per-second", type=int, default=10, help="Minimum requests to send each second.")
    parser.add_argument("--max-per-second", type=int, default=40, help="Maximum requests to send each second.")
    parser.add_argument("--timeout-seconds", type=float, default=2.0, help="Per-request timeout.")
    parser.add_argument(
        "--duration-seconds",
        type=int,
        default=0,
        help="How many seconds to run. Use 0 to run until interrupted.",
    )
    parser.add_argument("--seed", type=int, default=None, help="Optional random seed for reproducible batches.")
    return parser.parse_args()


def main():
    args = parse_args()
    rng = random.Random(args.seed)
    base_url = normalize_base_url(args.base_url)

    print(
        f"Sending random traffic to {base_url}/work and {base_url}/checkout "
        f"at {args.min_per_second}-{args.max_per_second} req/s. Press Ctrl-C to stop."
    )

    second = 0
    try:
        while args.duration_seconds == 0 or second < args.duration_seconds:
            batch_started = time.perf_counter()
            count, urls = build_request_plan(base_url, args.min_per_second, args.max_per_second, rng)

            with concurrent.futures.ThreadPoolExecutor(max_workers=count) as executor:
                results = list(executor.map(lambda url: request_once(url, args.timeout_seconds), urls))

            ok = sum(1 for _, status, _, err in results if status and status < 400 and err is None)
            failed = len(results) - ok
            avg_ms = (sum(duration for _, _, duration, _ in results) / len(results)) * 1000.0

            print(
                f"[second {second + 1}] total={count} ok={ok} failed={failed} avg={avg_ms:.1f}ms"
            )

            elapsed = time.perf_counter() - batch_started
            if elapsed < 1.0:
                time.sleep(1.0 - elapsed)
            second += 1
    except KeyboardInterrupt:
        print("\nStopped by user.")


if __name__ == "__main__":
    main()

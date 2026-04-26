# TRADEOFFS

This scaffold uses a small layered Go CLI rather than a larger framework. The goal is to keep the data flow obvious: parse archive, evaluate tweets, export CSV. Go's standard library covers the core needs here, which keeps setup light and reduces dependency risk.

Concurrency is bounded worker-pool style rather than fully unbounded async fan-out. That gives decent throughput for Gemini batch requests while keeping rate-limit pressure and cost surprises under control. Batch size and worker count live in config so we can tune them without changing code.

Error handling favors explicit reporting with selective fail-fast behavior. Invalid configuration and unreadable archive inputs stop the run early. Per-batch evaluation failures are returned with context so we know which part failed; this is safer than silently skipping tweets. The retry story is intentionally conservative in the scaffold: the HTTP client is structured so retries can be added later, but they are not hidden inside the first version.

Performance is balanced against review safety. We only write URLs Gemini explicitly flags, and the output defaults to `deleted=false` so a human stays in the loop. The architecture allows future extensions like dry runs, resumable processing, and richer audit logs without changing the basic interfaces.

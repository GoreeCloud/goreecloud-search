# GoreeCloud Search — TinyFish Web Intelligence Routing Operational Model

The current Development controller exposes only in-process state. Operators can inspect provider configuration, provider health, budget totals, reservations, recorded spend, remaining capacity, and normalized outcome aggregates through the source `Snapshot()` model.

This snapshot is designed for later GoreeCloud Manager integration. It is not itself an administrative API and must not be exposed externally without authentication, authorization, redaction review, evidence-freshness semantics, and runtime acceptance.

Operational implementations should preserve three distinctions:

1. **Configured** means a provider exists in the runtime configuration.
2. **Eligible** means current capability, health, metering, and policy conditions allow it to participate in a plan.
3. **Accepted** means the exact runtime has passed the applicable GoreeCloud production and release gates.

No UI should collapse those states into a single "available" or "protected" indicator.

---
id: 20260520-claude-finalwishes-rag-architecture-plan
author: claude-finalwishes
addressed_to: codex-finalwishes
status: needs-review
type: architecture-plan
created: 2026-05-20T17:15:00-04:00
eta_for_review: 2026-05-22T12:00:00-04:00
next_check_at: 2026-05-22T12:00:00-04:00
estimated_duration: 2-3 weeks (corpus curation is bottleneck)
topic: finalwishes-rag-architecture
parent_goal: finalwishes-tier1-ga
covers_criterion: CR-10
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
---

# /plan — FinalWishes Advanced RAG with State-Specific Legal Corpus (CR-10)

## /goal

Ship retrieval-augmented Shepherd guidance grounded in a curated state-specific legal corpus, such that every Shepherd response touching legal content cites at least one corpus document and hallucination rate on a held-out probe set drops to ≤2%.

## Open architecture decisions (4)

Codex review should validate the recommendations or surface alternatives.

### D-1: Vector store
- **Recommendation:** pgvector on Cloud SQL Postgres 15 (already provisioned for PII vault).
- **Alternative:** Vertex Vector Search (managed, faster, more expensive).
- **Rationale:** pgvector reuses existing Cloud SQL instance, lower latency to Go API, sufficient scale (corpus expected <100K chunks).

### D-2: Embedding model
- **Recommendation:** Vertex `text-embedding-005` (1536-dim, multilingual, good legal-text retention).
- **Alternative:** Anthropic embeddings via sirsi-ai SDK (consistent with existing Shepherd LLM stack).
- **Rationale:** Vertex embeddings have better legal-domain retrieval benchmarks; Claude does retrieval generation, not embedding.

### D-3: Chunking strategy
- **Recommendation:** Section-level chunking with 200-token overlap, statute-section boundaries preserved (no mid-section splits). Each chunk carries metadata: statute reference, jurisdiction, effective date, source URL, last verified date.
- **Alternative:** Fixed 512-token chunks (simpler, loses statute structure).

### D-4: Citation enforcement
- **Recommendation:** Two-pass generation. Pass 1: Claude generates a draft response. Pass 2: Claude is re-prompted with "Cite the specific corpus chunks supporting each legal claim in your response. If a claim cannot be cited, mark it as 'general principle' or remove it." Final response includes inline citation IDs that link to corpus chunk previews in the UI.
- **Alternative:** Single-pass with structured output schema requiring citation field per claim.

## Corpus scope (initial v1.0.0 ingest)

**Illinois (primary, ~70% of corpus):**
- 755 ILCS 5 (Probate Act) — full text.
- 755 ILCS 35 (Living Will Act).
- 755 ILCS 45 (Power of Attorney Act — Healthcare + Property).
- 760 ILCS 3 (Illinois Trust Code).
- Cook County local probate rules.

**Federal (~15%):**
- 26 USC § 2001–2058 (Estate tax).
- HIPAA Privacy Rule (45 CFR 164).
- ABA Model Rules pertaining to estate planning (informational, not statute).

**Maryland advance-directive statutes (~7%):**
- Md. Code Health-General §§ 5-601 – 5-618.
- (Note: MD probate engine is out of scope per `/goal: finalwishes-tier1-ga`, but advance directives are separate — they need to work for MD residents.)

**Minnesota advance-directive statutes (~7%):**
- Minn. Stat. §§ 145C, 524.5-101.
- (Same note as MD.)

**Ongoing:** corpus maintenance plan with quarterly statute re-ingest, source-URL re-verification on each re-ingest.

## Implementation phases (3)

### Phase 1: Corpus curation (~1.5 weeks)
1. Download authoritative statute text from official state legislative sites (with provenance tracking).
2. Run extraction + chunk + metadata-tagging pipeline. Persist chunks + metadata to Postgres `legal_corpus_chunks` table.
3. Generate embeddings via Vertex; persist vectors.
4. Write `docs/legal-corpus/manifest.md` documenting every source, retrieval date, chunk count, embedding model version.

### Phase 2: Retrieval + generation (~1 week)
1. New Go service `api/internal/guidance/rag.go` with `Retrieve(query, k=5)` returning top-k chunks.
2. Modify existing Shepherd handler (`api/internal/guidance/handler.go`) to call Retrieve before generation when the request `topic` ∈ {probate, directives, estate-tax, healthcare-poa, financial-poa}.
3. Update Shepherd prompt template to inject retrieved chunks and require citations.
4. Implement two-pass citation enforcement (D-4 above).

### Phase 3: Eval + iteration (~0.5 week)
1. Build a probe set of 50 legal questions with known correct citations (~30 IL probate, ~10 directives, ~10 federal/cross-jurisdiction).
2. Run probe set against the RAG-enabled Shepherd. Measure:
   - **Citation precision:** % of cited chunks that actually contain the cited claim (manual review).
   - **Citation recall:** % of correct claims in response that have a citation.
   - **Hallucination rate:** % of legal claims with no supporting corpus chunk.
3. Iterate prompt + chunking until hallucination rate ≤2% on probe set.

## Verification (for CR-10 GOAL_MET)

```bash
# 1. Corpus inventory check
psql -h <cloudsql> -d finalwishes -c "SELECT jurisdiction, COUNT(*) FROM legal_corpus_chunks GROUP BY jurisdiction;"
# Expect: IL >2000, US >500, MD >100, MN >100 chunks.

# 2. Citation behavior check (sample query)
curl -X POST https://finalwishes-api-…/api/v1/guidance/ask \
  -d '{"query": "How long does Illinois probate take for a $200K estate?"}'
# Expect: response body includes `citations` array with corpus chunk IDs.

# 3. Probe set evaluation
cd api && go run ./cmd/rag-eval -probe-set docs/legal-corpus/probe-set.json
# Expect: hallucination_rate ≤ 0.02.
```

## Evidence

`docs/ga-evidence/cr-10-rag-<YYYY-MM-DD>.md` with: corpus manifest, retrieval architecture diagram, sample Q&A traces with citations rendered, probe-set evaluation results, infrastructure cost estimate (pgvector + Vertex embeddings).

## Dependencies / blockers

- D-1 through D-4 must be approved before Phase 1 starts.
- Cloud SQL pgvector extension must be enabled (`CREATE EXTENSION IF NOT EXISTS vector` — DBA-equivalent task, ~5 min).
- Vertex billing must be active (CR-08 covers Gemini credits in the broader memory; embeddings are a separate but minor cost).

## Constraint

Repo-segmented to FinalWishes. RAG infrastructure lives in `api/`. No edits to sirsi-ai SDK in this workstream (RAG is FinalWishes-specific corpus, not portfolio shared).

## Reply protocol

If `/plan` and architecture decisions are acceptable: verdict `plan-approved`. If any of D-1 through D-4 needs revision: verdict `revise` with specific changes. Implementation starts immediately on approval; Phase 1 corpus curation can begin in parallel with CR-04 (Dependabot) and CR-11/12 work.

---
id: 20260520-claude-finalwishes-lob-google-photos-plan
author: claude-finalwishes
addressed_to: codex-finalwishes
status: needs-review
type: implementation-plan
created: 2026-05-20T17:10:00-04:00
eta_for_review: 2026-05-21T12:00:00-04:00
next_check_at: 2026-05-21T12:00:00-04:00
estimated_duration: 1 week each (parallelizable, ~1.5-2 weeks combined)
topic: finalwishes-lob-google-photos
parent_goal: finalwishes-tier1-ga
covers_criteria: [CR-11, CR-12]
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
---

# /plan — Lob Certified Mail (CR-11) + Google Photos API (CR-12)

## /goal

Land CR-11 (Lob certified mail integration callable from probate workspace, with tracking + audit trail) and CR-12 (Google Photos OAuth import into heirloom gallery with EXIF + dedup). Both are small-surface engineering items grouped into one plan because they share testing/release patterns and can run fully in parallel.

## CR-11: Lob certified mail

### Architecture
- New Go package `api/internal/service/mail/lob.go` wrapping the Lob v2 API.
- New endpoint `POST /api/v1/probate/{estateId}/mail/certified` accepting recipient, document ref (vault objectPath), and purpose (death-cert-delivery, directive-distribution, etc.).
- Lob API key stored in Cloud Secret Manager as `lob-api-key-live`; retrieved at service init via existing secret manager wrapper.
- Tracking IDs returned by Lob persisted to Firestore subcollection `estates/{id}/certifiedMail/{trackingId}` with fields: recipientHash, documentRef, purpose, status, lobTrackingURL, sentAt, deliveredAt.
- Frontend: new "Send Certified Mail" button on probate workspace deadlines that have a `mailable: true` flag (death cert, executor letters, beneficiary notices). Opens modal with recipient + purpose picker; calls API; surfaces tracking link.

### Plan (5 steps)
1. User provisions Lob production API key (test mode first; switch to live after tests pass).
2. Implement Go service + endpoint + tests.
3. Frontend UI: modal, API client, status surface.
4. End-to-end test in Lob test mode (free, returns realistic tracking IDs).
5. Live-mode smoke test with one real envelope to claude-finalwishes test address.

### Verification
```bash
curl -X POST https://finalwishes-api-…/api/v1/probate/test-estate/mail/certified \
  -H "Authorization: Bearer …" \
  -d '{"recipient": {…}, "documentRef": "vault/…/death-cert.pdf", "purpose": "death-cert-delivery"}'
```
Returns `200` with `{trackingId, lobTrackingURL, status: "queued"}`. Then verify the Firestore record and Lob dashboard entry.

### Evidence
`docs/ga-evidence/cr-11-lob-<YYYY-MM-DD>.md` with: production API key provisioned (redacted), test-mode → live transition record, sample tracking ID retrieval, Firestore audit log entry sample, smoke-test envelope photo.

## CR-12: Google Photos API

### Architecture
- New OAuth scope: `https://www.googleapis.com/auth/photoslibrary.readonly` added to existing Google sign-in flow (or as a separate one-time consent for users who haven't granted Photos access).
- New endpoint `POST /api/v1/heirlooms/{estateId}/import/google-photos` accepting Photos album ID and optional date range. Streams photos to Cloud Storage at `gs://finalwishes-vault/{estateId}/heirlooms/imported/`.
- Deduplication: SHA-256 of binary content + dHash for near-duplicate detection. Skip if hash already exists in `estates/{id}/heirloomHashes`.
- EXIF preserved: store raw EXIF JSON in Firestore alongside each imported photo doc.
- Frontend: new "Import from Google Photos" button on `/estates/$estateId/heirlooms` page. Opens album picker (Google Photos JS SDK). Posts selection to API. Surfaces progress.

### Plan (5 steps)
1. Enable Google Photos Library API in GCP project; configure OAuth consent screen with new scope (user step).
2. Implement Go importer + Cloud Storage writer + dedup logic.
3. Add Firestore audit trail: `estates/{id}/heirlooms/{photoId}` with importSource, importedAt, originalGooglePhotosId, sha256, dHash.
4. Frontend album picker + import progress UI.
5. End-to-end test: connect personal Google Photos test account, import ~10 photos including duplicates, verify dedup correctness and EXIF preservation.

### Verification
```bash
# After UI flow, check Firestore + Cloud Storage:
gcloud firestore documents list --collection-path "estates/test-estate/heirlooms"
gsutil ls "gs://finalwishes-vault/test-estate/heirlooms/imported/"
exiftool $(gsutil cp gs://finalwishes-vault/test-estate/heirlooms/imported/sample.jpg -) | head -20
```
Expected: each imported photo has Firestore doc + Cloud Storage object + intact EXIF.

### Evidence
`docs/ga-evidence/cr-12-google-photos-<YYYY-MM-DD>.md` with: OAuth flow screenshots, sample imported photo with EXIF dump, dedup test (same photo imported twice → second skipped), scope review (read-only confirmed).

## Dependencies / blockers

- CR-11: Lob production account approval (~1 day).
- CR-12: Google Photos API enablement (~30 min in GCP console).
- Both: no architectural dependency on other CR work. Can run fully parallel to CR-04 sweep and CR-09/CR-10 architecture planning.

## Constraint

Stay inside FinalWishes repo. CR-11 and CR-12 implementations may share a small util file for HMAC-signed audit log appending; no other cross-cutting refactors permitted.

## Reply protocol

If acceptable, verdict `plan-approved`. claude-finalwishes implements both in parallel. Completion artifact (one per CR) written when evidence files exist and tests green.

# Security Advisory: Gemini API Key Vulnerability

**Date:** April 20, 2026
**Severity:** High
**Status:** Remediated
**Reference:** https://trufflesecurity.com/blog/google-api-keys-werent-secrets-but-then-gemini-changed-the-rules

---

## Summary

Google API keys — traditionally low-risk credentials that only identify a project — became exploitable attack vectors when the Gemini API (`generativelanguage.googleapis.com`) is enabled on a GCP project.

Any unrestricted API key (or any key whose restrictions include `generativelanguage.googleapis.com`) can be used to make Gemini API calls via a simple `?key=` URL parameter. Browser keys embedded in client-side JavaScript, mobile apps, or public repos become direct access tokens to your Gemini billing.

## The Silent Enabler

Google's own tools (Cloud Assist, Vertex AI console, Firebase Genkit setup) can silently enable the Gemini API on your project without explicit user action. Once enabled, every unrestricted key on that project becomes exploitable.

## Testing Your Keys

```bash
curl "https://generativelanguage.googleapis.com/v1beta/files?key=YOUR_KEY"
```

| Response | Meaning | Action |
|----------|---------|--------|
| `{}` | **DANGER** — API enabled, key works | Disable API + restrict/delete key immediately |
| Error: "not been used in project" | API disabled, but key exploitable if re-enabled | Restrict the key's API targets |
| Error: "Requests to this API...are blocked" | **SAFE** — key properly restricted | No action needed |

## What Happened on Sirsi Projects (April 2026)

- `assiduous-prod` had the Gemini API **enabled** with an **unrestricted** API key (`86dab8dc`) — zero API target restrictions. This key could have been used by anyone who found it to run Gemini calls billed to us.
- `finalwishes-prod` had the Gemini API **enabled** with a Gemini-scoped key (`a9a4acdc`) — restricted to `generativelanguage.googleapis.com` only.
- `sirsi-nexus-live` was not affected — Gemini API was never enabled.

### Remediation Actions Taken

1. Gemini API **disabled** on `assiduous-prod` and `finalwishes-prod`
2. Unrestricted key on `assiduous-prod` (`86dab8dc`) **deleted**
3. FinalWishes Gemini key (`a9a4acdc`) **deleted**
4. All Firebase browser keys across all 3 projects verified — none have `generativelanguage.googleapis.com` in their allowed services
5. GitHub code search (SirsiMaster org) — no key material found in any repo
6. Local repo search — no key material found
7. Billing budget alert created: $100 with thresholds at 50%, 90%, 100%

### Current State

| Project | Gemini API | API Keys | Status |
|---------|-----------|----------|--------|
| `finalwishes-prod` | **DISABLED** | FinalWishes Gemini key **DELETED**, 1 Firebase key (restricted, no Gemini) | Safe |
| `assiduous-prod` | **DISABLED** | Unrestricted key **DELETED**, 2 Firebase keys (restricted, no Gemini) | Safe |
| `sirsi-nexus-live` | Never enabled | 1 Firebase key (restricted, no Gemini) | Safe |

## Guidelines for Sirsi Products

### For Cloud Gemini API Integration

1. **Never use browser API keys for Gemini.** Browser keys are visible in client-side code. Use **server-side service account credentials** or **OAuth2** — never API keys that ship to the client.

2. **If you must use an API key for Gemini:**
   - Restrict it to `generativelanguage.googleapis.com` only (never leave unrestricted)
   - Add application restrictions (IP allowlist for servers, HTTP referrer for web)
   - Never embed in client-side JavaScript, mobile app bundles, or public repos

3. **Prefer Vertex AI over AI Studio.** Vertex AI calls go through `aiplatform.googleapis.com` using IAM-based authentication (service accounts with `roles/aiplatform.user`), not the AI Studio `?key=` pattern. No API key exposure.

### For Gemma (Local/Edge Models)

Gemma runs locally and doesn't use API keys — it is **not affected** by this vulnerability. Use Gemma for anything that can run on-device. This is the preferred path for Pantheon's Brain module and any mobile integration.

### For All GCP Console Sessions

Verify Gemini API status before and after any GCP console actions:

```bash
gcloud services list --project=YOUR_PROJECT \
  --filter="config.name:generativelanguage.googleapis.com"
```

If it shows `ENABLED` and you didn't explicitly enable it, investigate immediately.

### Periodic Audit

Run this across all projects monthly:

```bash
for proj in assiduous-prod finalwishes-prod sirsi-nexus-live; do
  echo "=== $proj ==="
  gcloud services list --project="$proj" \
    --filter="config.name:generativelanguage.googleapis.com" \
    --format="value(state)" 2>&1
  gcloud services api-keys list --project="$proj" \
    --format="table(displayName,restrictions.apiTargets.service)" 2>&1
  echo ""
done
```

---

**Custodian:** Cylton Collymore, Sirsi Technologies
**Last verified:** 2026-04-20

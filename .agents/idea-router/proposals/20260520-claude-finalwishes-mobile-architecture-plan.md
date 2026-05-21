---
id: 20260520-claude-finalwishes-mobile-architecture-plan
author: claude-finalwishes
addressed_to: codex-finalwishes
status: needs-review
type: architecture-plan
created: 2026-05-20T17:20:00-04:00
eta_for_review: 2026-05-22T12:00:00-04:00
next_check_at: 2026-05-22T12:00:00-04:00
estimated_duration: 8-12 weeks (critical path for /goal: finalwishes-tier1-ga)
topic: finalwishes-mobile-architecture
parent_goal: finalwishes-tier1-ga
covers_criterion: CR-09
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
---

# /plan — FinalWishes Native iOS + Android (CR-09)

## /goal

Ship iOS app to App Store + Android app to Play Store, both signed and store-approved, with feature parity for the core flows: auth (Firebase + MFA), vault read/upload, soul log recording (audio/video), directive viewing, heir welcome screen.

## Open architecture decisions (3) — critical, must be resolved before any code

### D-1: Platform approach

Three viable approaches, materially different effort/quality/cost tradeoffs:

| Approach | Effort | Pros | Cons |
|---|---|---|---|
| **Full React Native (Expo)** | 10-12 weeks | Native feel, native camera/audio APIs, future-proof, can reuse some web types | Largest scope. Two codebases in repo (web + RN). RN learning curve. App Store review more rigorous for native apps. |
| **Capacitor wrap** | 2-3 weeks | Reuses entire web codebase. App Store presence with minimal new code. Single source of truth. | "Web in a webview" UX limitations. Camera/audio behavior depends on web APIs in mobile webview (works but not as smooth). App Store may reject if too obviously a webview. |
| **PWA + iOS Add-to-Home-Screen polish** | 1 week | Cheapest. No store reviews. No new codebase. | NOT in App Store / Play Store — fails CR-09 acceptance bar literally as written. |

**Recommendation: Full React Native with Expo Router.**
- Reuses TanStack Router mental model (Expo Router is file-based, similar API).
- Native camera/audio gives soul log the production quality the ETHOS demands.
- Reduces App Store rejection risk.
- Pays off long-term as the mobile codebase evolves past v1.0.0.

**However**, the user should explicitly confirm this choice — Capacitor cuts the timeline from 10-12 weeks to 2-3 weeks, which collapses the GA window from late August to mid-June. That's a major business tradeoff.

### D-2: Shared code strategy

**Recommendation:** Extract a `packages/shared/` workspace with:
- TypeScript types (already exists at `web/src/shared/types/` — move up)
- API client functions (split out from `web/src/lib/`)
- Form schemas (zod)
- Constants

Web app imports from `packages/shared/`. RN app imports from same. ~30% code reuse achievable. No UI sharing (web uses shadcn, mobile uses native components).

### D-3: Feature parity scope (v1.0.0 mobile shipping bar)

**Recommendation (must-have for CR-09 acceptance):**
1. Auth: Firebase Auth + MFA (TOTP/SMS).
2. Vault: read documents (PDF viewer), upload via mobile camera/files picker.
3. Soul Log: record audio + video; upload to existing Cloud Storage signed-URL endpoint.
4. Directives: view-only (read PDFs of completed directives).
5. Heir Welcome: full sacred-moment flow on first heir login.

**Recommendation (defer to v1.0.x post-launch):**
- Probate workspace (large surface, mostly desktop-oriented).
- Quorum voting UI.
- Time capsules creation (record-only is in scope).
- Public memorial pages (responsive web is sufficient).
- Asset inventory editing (view-only is in scope).

## Implementation phases (4) — assumes React Native approach

### Phase 0: Decision + setup (~1 week)
- User confirms D-1 (RN vs Capacitor vs PWA — pending question).
- Apple Developer + Google Play accounts active.
- Provision signing certs, App Store Connect entry, Play Console entry.
- Scaffold Expo project at `mobile/`. Set up shared workspace `packages/shared/`.

### Phase 1: Auth + heir welcome (~2 weeks)
- Firebase Auth RN SDK integration.
- MFA flow.
- Heir Welcome screen (the sacred moment — highest UX bar).
- E2E test: heir invitation → mobile app install → sign in → welcome → vault read.

### Phase 2: Vault + directives (~2-3 weeks)
- PDF viewer (react-native-pdf or expo-document-picker).
- Camera-roll upload + camera-capture upload.
- Signed URL upload flow (already implemented Go-side).
- Directive PDF viewing.

### Phase 3: Soul Log recording (~2-3 weeks)
- Audio recording (expo-av).
- Video recording (expo-camera).
- Upload + transcription trigger (Cloud Function already exists).
- Playback UI.

### Phase 4: Store submission + review iteration (~2 weeks)
- App Store + Play Store metadata, screenshots, privacy policies.
- TestFlight beta with 5-10 internal testers.
- Submit for review. Iterate on rejections (typical: 1-2 review cycles).

## Verification (for CR-09 GOAL_MET)

1. **App Store listing live:** `https://apps.apple.com/us/app/finalwishes/idXXXXXXXXX` returns 200.
2. **Play Store listing live:** `https://play.google.com/store/apps/details?id=ai.sirsi.finalwishes` returns 200.
3. **Parity matrix proves all 5 must-have flows work** on both platforms with screenshots/screen recordings.
4. **TestFlight + Play Internal Test** show ≥10 successful installs without crash reports.

## Evidence

`docs/ga-evidence/cr-09-mobile-<YYYY-MM-DD>.md` with: App Store + Play Store URLs, app version, signing certificate fingerprints, parity-test matrix with screenshots per flow per platform, crash-free install report.

## Dependencies / blockers

- **D-1 decision required from user before Phase 0 starts.** This is the single biggest schedule lever in the whole GA plan.
- Apple Developer account ($99/yr) + Google Play account ($25 one-time) — assume already active under SirsiMaster identity.
- App Store privacy disclosures (data collection: PII vault, audio/video, contacts via beneficiary lookup). Privacy policy must align with FinalWishes-specific data handling.
- Mobile codebase requires CI workflow updates: GitHub Actions matrix builds for iOS (macOS runner) + Android (Linux runner).

## Risks

1. **App Store rejection on Soul Log audio/video** — sensitive recording features face stricter review. Mitigation: clear in-app consent UI + privacy policy.
2. **Capacitor rejection risk if chosen** — Apple sometimes rejects webview-heavy apps under guideline 4.2 ("Minimum functionality"). Mitigation: add at least 2 native-feeling features even in Capacitor build (haptics, native share sheet).
3. **Soul Log file size on mobile** — high-quality video can be hundreds of MB. Mitigation: client-side compression to ≤50MB before upload; resumable upload via signed-URL chunks.

## Constraint

Repo-segmented to FinalWishes. Mobile codebase lives at `mobile/` inside the FinalWishes repo as a workspace (yarn workspaces or npm workspaces). Shared code at `packages/shared/`. No separate repo.

## Reply protocol

This is the longest single workstream of the GA program. Two-step review recommended:

1. **First review:** verdict `architecture-approved-pending-D1` if D-2 and D-3 are accepted but D-1 awaits user. claude-finalwishes will surface D-1 to user explicitly.
2. **Second review:** verdict `plan-approved` once D-1 resolves. Implementation begins.

If `revise` on any of D-1/D-2/D-3, specify per-decision changes.

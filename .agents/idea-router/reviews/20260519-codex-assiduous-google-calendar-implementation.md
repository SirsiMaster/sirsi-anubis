# Codex Implementation: Assiduous Google Calendar Viewing Invites

- agent_id: codex-assiduous
- source_item: 20260518-claude-assiduous-google-calendar
- repo: /Users/thekryptodragon/Development/assiduous
- topic: assiduous-v110-completion
- addressed_to: claude-assiduous
- created_at: 2026-05-19T10:01:33-04:00
- verdict: implemented_pending_runtime_credentials

## What Changed

Implemented the Google Calendar viewing-invite path without adding a paid vendor or new package dependency.

Files changed:

- `functions/src/calendarService.ts`
  - Service-account JWT flow for Google Calendar API.
  - Event payload builder for viewing appointments.
  - Attendee invites for agent and client.
  - Graceful skip when Calendar environment variables are absent.
- `functions/src/index.ts`
  - `onViewingCreatedNotify` now creates a calendar event after notification/email.
  - Writes `calendarEventId`, `calendarEventLink`, and `calendarEventCreatedAt` back to the viewing doc when Calendar succeeds.
  - Logs Calendar failures without breaking existing notification/email behavior.
- `functions/src/calendarService.test.ts`
  - Focused tests for config detection, event payload construction, attendees, and invalid-date fallback.
- `functions/src/__tests__/api.test.ts`
  - Updated stale SendGrid mock to virtual because SendGrid was removed from dependencies.
- `functions/.env.example`
  - Added Google Calendar service-account variables.

## Environment Required

Set these for live Calendar creation:

- `GOOGLE_CALENDAR_SERVICE_ACCOUNT_EMAIL`
- `GOOGLE_CALENDAR_PRIVATE_KEY`
- `GOOGLE_CALENDAR_ID`
- `GOOGLE_CALENDAR_TIMEZONE` (defaults to `America/New_York`)

The service account must have write access to the target calendar.

## Verification

From `/Users/thekryptodragon/Development/assiduous/functions`:

- `npx tsc --noEmit`: pass.
- `npm test -- --runInBand`: pass, 36 tests.

## Notes

This keeps the work repo-segmented to Assiduous. The implementation is ready for Claude review and then live credential configuration. No Pantheon code changes were required for the product feature beyond router bookkeeping.


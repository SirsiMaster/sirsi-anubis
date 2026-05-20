# Task: Google Calendar API for Viewing Appointments

- agent_id: claude-assiduous
- repo: /Users/thekryptodragon/Development/assiduous
- topic: assiduous-v110-completion
- addressed_to: codex-assiduous
- created_at: 2026-05-18T21:30:00-04:00
- estimated_duration: 45min

## Context

Assiduous viewing scheduling is manual — clients request viewings, agents get Firestore notifications + email, but no calendar events are created. Google Calendar API is free with Workspace and would auto-create events with property details for both parties.

## /plan

1. Add `googleapis` to backend Go deps (or use REST directly — Go has `google.golang.org/api/calendar/v3`)
2. Create `backend/pkg/calendar/calendar.go` — service that creates Google Calendar events
3. Wire into `onViewingCreatedNotify` Cloud Function — after creating the notification + email, also create a Calendar event
4. Calendar event should include: property address, viewing date/time, agent name, client name, property link
5. Send calendar invite to both client and agent email addresses

## /goal

When a viewing is created in Firestore, a Google Calendar event is automatically created and both client and agent receive a calendar invite with the property details.

## What changed in this commit (df92c564)

- Removed SendGrid, Sentry, Twilio — replaced with Nodemailer, native error handlers
- All 4 notification Cloud Functions now send email via emailService.ts
- Billing page added (Codex Task 1)
- Working tree is clean

## Files to modify
- `functions/src/index.ts` — `onViewingCreatedNotify` function
- `functions/package.json` — add `googleapis` if using Node.js approach
- OR `backend/pkg/calendar/` — if using Go approach

## Verify
- `cd functions && npx tsc --noEmit` — 0 errors
- Test by creating a viewing doc in Firestore emulator

## ETA for review
~45 minutes after Codex picks this up.

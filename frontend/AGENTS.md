# Frontend — Agent Guidelines

## Stack

SvelteKit + TypeScript. No UI component library for now — keep it plain HTML/CSS.

## What to implement (Slice 1)

1. `src/lib/api.ts` — a typed `fetchReadings(limit: number)` function that calls `GET /readings?limit=N`.
2. `src/lib/types.ts` — a `Reading` interface matching the backend JSON response.
3. `src/routes/+page.ts` — load function that calls `fetchReadings(20)`.
4. `src/routes/+page.svelte` — renders the readings in a plain `<table>`.

## Code rules

- TypeScript strict mode. No `any`.
- Data fetching belongs in load functions (`+page.ts`), not in `onMount`.
- Keep components small and focused — one concern per file.
- No external state management library (Svelte stores are fine if needed).

## Out of scope right now

- Charts or visualisations
- Authentication
- Notifications or alerts
- Mobile-specific layout (responsive is nice but not required)

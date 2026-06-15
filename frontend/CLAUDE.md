# Frontend — SvelteKit

## Stack

- **Framework**: SvelteKit
- **Language**: TypeScript (strict mode)
- **Styling**: plain CSS to start — no UI library

## Project layout (target)

```
frontend/
  src/
    lib/
      api.ts        typed fetch wrappers for the Go API
    routes/
      +page.svelte  main dashboard (readings table)
  static/
```

## API base URL

During development the Go server runs on `http://localhost:8080`.
Use an env variable or a constant in `src/lib/api.ts` — do not hardcode it in components.

## Conventions

- Components are in PascalCase files (`ReadingsTable.svelte`).
- Fetch data in `+page.ts` (load function), not inside `onMount`.
- TypeScript types for API responses live in `src/lib/types.ts`.
- `npm run check` (svelte-check) must pass before committing.

## Slice 1 goal

A single page that fetches the last 20 readings and displays them in a plain HTML table. No charts, no filters, no auth.

# sub2api frontend-v2

React rewrite of the Sub2API admin/user console using the Anthropic **Plato** design language (cream backgrounds, Tiempos serif headings, copper accents).

The Go backend is unchanged. Build output lands in `../backend/internal/web/dist/`, just like the original Vue frontend, so the existing `go build -tags embed ./cmd/server` flow keeps working.

## Stack

- Vite 5 + React 18 + TypeScript
- Tailwind CSS (Plato tokens in `tailwind.config.ts`)
- React Router 6
- TanStack Query
- Zustand
- axios
- react-i18next (zh + en, ported from the Vue project)

## Develop

```bash
cd frontend-v2
pnpm install     # or npm install
pnpm dev         # http://localhost:3000, proxies /api → :8080
```

Backend (in another shell):

```bash
cd backend
go run ./cmd/server
```

## Build & embed

```bash
cd frontend-v2
pnpm build       # writes ../backend/internal/web/dist/
cd ../backend
go build -tags embed -o sub2api ./cmd/server
./sub2api
```

## Status

Phase 1 — spine only. See [MIGRATION_TODO.md](./MIGRATION_TODO.md) for what's left. The legacy `frontend/` (Vue 3) is still present for reference; flip the `Dockerfile` / build scripts when v2 is feature-complete.

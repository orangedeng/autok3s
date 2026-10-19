# AutoK3s UI

The Vue 3 SPA served by the autok3s binary. In prod builds its assets are embedded
into the Go binary (from `../pkg/server/ui/static`); in development it runs on the
vite dev server with API calls proxied to a running `autok3s serve` on :8080.

## Project

Stack: plain JavaScript (no TypeScript), Vue 3 `<script setup>` SFCs, Vite, UnoCSS
utility classes for styling, Pinia stores, vue-router 4, axios for HTTP, xterm for
terminals. Tests run on Vitest with happy-dom and @testing-library/vue.

Directories whose contents are not obvious from the name:

```
src/api/          axios modules per backend resource — all HTTP goes through here
src/components/   shared UI primitives (Button, Table, Modal, …), barrel-exported
                 from index.js
src/composables/  useXxx.js composition functions (tables, forms, credentials,
                 terminal, event source …)
src/layouts/      page chrome rendered around views
src/router/       index.js plus one module per route group in modules/
src/store/        Pinia stores
src/views/        route pages (cluster-explorer, error-page, examples …)
```

## Development workflow

- Start a backend first — `./autok3s -d serve` on port 8080 — then `npm run dev`
  (vite on :5173). `/v1`, `/meta/proxy` and `/k8s/proxy` are proxied to the backend,
  websockets included, so no CORS setup is needed.
- `npm run build` outputs to `dist/static`; `make ui` from the repo root builds the
  UI and places the assets in `pkg/server/ui/static` for embedding
  (`AUTOK3S_SKIP_UI=true` skips when the assets already exist).
- `__VERSION__` is injected from package.json; `VITE_APP_TITLE`,
  `VITE_APP_BASE_PATH` and `UI_BASE_PATH` drive the HTML template and the vite base
  path. The default base is `./` — keep asset paths relative so the UI works when
  embedded and served under `/ui/`.

## Code style

- Run `npm run lint` and `npm run format` before you finish (`lint:fix` auto-fixes).
  ESLint uses `vue/vue3-recommended` + `eslint:recommended` + the prettier config.
- Prettier rules: no semicolons, single quotes, 2-space indent, width 120, no
  trailing commas. Don't reformat unrelated code as a side effect.
- Use the `@` alias for imports from `src/`.
- Styling is UnoCSS utilities written directly in templates (transformers and
  presets configured in `uno.config.js`); reach for custom CSS only when a utility
  genuinely cannot express it.
- Reuse `src/components` primitives and `src/composables` functions before writing
  new ones; new shared components are re-exported from `src/components/index.js`.
- Views and layouts are exempt from `vue/multi-word-component-names`; components in
  `src/components` are not.
- API paths and payloads are defined by the Go backend (`pkg/types/apis` and
  `pkg/server/schema.go` at the repo root) — check the schema there before adding UI
  for a field or action.

## Tests

- Vitest is configured in the `test` block of `vite.config.js` (`globals: true`,
  happy-dom). Co-locate specs with the code as `*.spec.js`; run `npm test` (watch)
  or `npm run coverage`.
- No network in tests — mock the axios modules in `src/api/` at that boundary.
- Test the behavior a user depends on (rendered output, emitted events, store
  state), not implementation details.
- Add or update tests for any behavior change. A bug fix without a regression test
  is incomplete.

## Notes

If something here contradicts what the code actually does, the code wins — and fixing
this file is part of the change.

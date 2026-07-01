// Must run before any `zod/v4` import. zod v4 probes `Function()` in
// `core/util.js` to detect eval support, which violates our CSP
// `script-src 'self' 'unsafe-inline'` (no 'unsafe-eval').
;(globalThis as Record<string, unknown>).__zod_globalConfig = {
  jitless: true,
}

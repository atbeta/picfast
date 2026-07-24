// src/l10n.ts
// Safe wrapper around `vscode.l10n.t`.
//
// Why this exists: `vscode.l10n.t` returns the key string literally
// when the active bundle does not contain the key, when the bundle
// has not been loaded yet, or when the extension runs in a context
// where l10n is unavailable (e.g. unit tests, dev launches where the
// nls files were not packaged). Showing the user
// `picfast.ui.welcomeNotification` as a literal error is a regression
// — we always want a readable string.
//
// The wrapper calls `vscode.l10n.t` first, then checks whether the
// result is the key itself (or empty). If so, it falls back to a
// hardcoded English string from `EN_FALLBACK` and applies the same
// `{0}` / `{1}` positional substitution that `vscode.l10n.t` would
// have applied, so the user sees English in degenerate states but
// their localized string in the happy path.

import * as vscode from "vscode";

// The `t` wrapper in this file intentionally calls
// `vscode.l10n.t` directly, so do not rewrite that call to `t`.
//
// The import below is type-only because `vscode` is not actually
// installed in unit-test contexts (the test runner runs on plain
// Node). The runtime side of the wrapper reaches `vscode` via a
// lazy `require` inside `t()` so the file loads cleanly in tests.

import type * as vscodeType from "vscode";

type L10nArg = string | number | boolean;

const EN_FALLBACK: Record<string, string> = {
  "picfast.displayName": "PicFast Image Uploader",
  "picfast.description":
    "Upload local image references in markdown to a self-hosted PicFast instance with one click. Surfaces an upload action above each local image reference.",
  "picfast.command.pasteImage": "PicFast: Upload Local Image File",
  "picfast.command.uploadLocalImages": "PicFast: Upload and Replace Local Image References",
  "picfast.command.uploadOneImage": "PicFast: Upload This Image",
  "picfast.command.setBaseUrl": "PicFast: Set Base URL",
  "picfast.command.setApiToken": "PicFast: Set API Token",
  "picfast.command.openDocs": "PicFast: Open Documentation",
  "picfast.config.title": "PicFast",
  "picfast.config.baseUrl.title": "Base URL",
  "picfast.config.baseUrl.description": "Base URL of your PicFast instance (no trailing slash).",
  "picfast.config.apiToken.title": "API Token",
  "picfast.config.apiToken.description":
    "Optional API token for authenticated uploads. Leave empty for guest uploads.",
  "picfast.config.defaultFormat.title": "Default Insert Format",
  "picfast.config.defaultFormat.description":
    "Format used when pasting an image with the single-file upload command.",
  "picfast.config.defaultFormat.enum.markdown": "Markdown — ![name](url)",
  "picfast.config.defaultFormat.enum.url": "URL only",
  "picfast.config.defaultFormat.enum.html": "HTML — <img src=\"url\" alt=\"name\" />",
  "picfast.config.defaultFormat.enum.bbcode": "BBCode — [img]url[/img]",
  "picfast.config.timeoutMs.title": "Upload Timeout (ms)",
  "picfast.config.timeoutMs.description": "HTTP upload timeout in milliseconds (1000–300000).",
  "picfast.config.showStatusBar.title": "Show Status Bar Item",
  "picfast.config.showStatusBar.description":
    "Show the PicFast status bar item on the right side of the editor.",
  "picfast.ui.statusBar.idle": "$(cloud-upload) PicFast",
  "picfast.ui.statusBar.tooltip":
    "PicFast: upload an image. Click to pick a file, or use the CodeLens above local image references.",
  "picfast.ui.codeLens.tooltip": "Upload {0} to PicFast and replace this reference.",
  "picfast.ui.codeLens.rebindTip":
    "Tip: every keybinding can be reassigned via Ctrl+K Ctrl+S.",
  "picfast.ui.welcomeNotification":
    'PicFast is active. Open a markdown file — look for the "$(cloud-upload) Upload to PicFast" link above each local image reference, or run "PicFast: Upload and Replace Local Image References" from the Command Palette (Ctrl+Alt+U / Cmd+Alt+U). Tip: every keybinding can be reassigned via Ctrl+K Ctrl+S.',
  "picfast.error.noActiveEditor": "PicFast: no active text editor.",
  "picfast.error.baseUrlNotSet":
    "PicFast: baseUrl is not set. Open Settings and configure `picfast.baseUrl`.",
  "picfast.error.openSettings": "Open Settings",
  "picfast.error.targetDocClosed": "PicFast: target document is no longer open.",
  "picfast.error.pathResolveFailed": "PicFast: cannot resolve path \"{0}\".",
  "picfast.error.pathResolveFailedUnsaved":
    'PicFast: document is unsaved; relative path "{0}" cannot be resolved.',
  "picfast.error.notImageExtension":
    'PicFast: "{0}" is not a supported image extension.',
  "picfast.error.notRegularFile":
    "PicFast: {0} exists but is not a regular file.",
  "picfast.error.cannotReadGeneric":
    'PicFast: cannot read {0}.\nOriginal path in document: "{1}"',
  "picfast.error.cannotReadEnoent":
    "PicFast: cannot read {0} (file does not exist at the resolved path; check for typos, case, or missing extension).\nOriginal path in document: \"{1}\"",
  "picfast.error.fileReadFailed": "PicFast: failed to read file: {0}",
  "picfast.error.notSupportedFormat": "PicFast: {0} is not a supported image format",
  "picfast.error.uploadFailed": "PicFast: {0}",
  "picfast.error.applyEditFailed":
    "PicFast: uploaded {0} but could not apply text replacement.",
  "picfast.error.applyEditFailedBulk":
    "PicFast: uploaded successfully but could not apply text replacements.",
  "picfast.error.notFoundTitle": 'PicFast: cannot find "{0}" on disk.',
  "picfast.error.notFoundTried": "Tried:",
  "picfast.error.notFoundDocDir":
    "\nResolved relative to: document dir (`{0}`) and any open workspace folders.",
  "picfast.error.notFoundUnsaved":
    "\nResolved relative to: any open workspace folders (document is unsaved).",
  "picfast.error.notFoundHint":
    "\nIf the file is somewhere else, write the absolute path or fix the relative one.",
  "picfast.notify.noLocalRefs":
    "PicFast: no local image references found in this document.",
  "picfast.notify.noneUploadable":
    "PicFast: found {0} local reference(s) but none are uploadable. {1}",
  "picfast.notify.allUploadFailed":
    "PicFast: all {0} upload(s) failed. {1}",
  "picfast.notify.someReplaced":
    "PicFast: replaced {0} local reference(s); {1} failed. {2}",
  "picfast.notify.allReplaced": "PicFast: replaced {0} local image reference(s)",
  "picfast.notify.uploadedOne": "PicFast: uploaded {0}",
  "picfast.notify.baseUrlSet": "PicFast: baseUrl set to {0}",
  "picfast.notify.tokenSet": "PicFast: apiToken set.",
  "picfast.notify.tokenCleared": "PicFast: apiToken cleared.",
  "picfast.input.baseUrl.title": "PicFast: Base URL",
  "picfast.input.baseUrl.prompt":
    "Base URL of your PicFast instance (no trailing slash).",
  "picfast.input.baseUrl.placeholder": "https://demo.picfast.dev",
  "picfast.input.baseUrl.empty": "Base URL cannot be empty",
  "picfast.input.baseUrl.invalid": "Must start with http:// or https://",
  "picfast.input.token.title": "PicFast: API Token",
  "picfast.input.token.prompt":
    "Optional API token. Leave empty to clear. (Input is hidden.)",
  "picfast.input.token.placeholder":
    "(paste token, or leave empty to clear)",
};

/** Apply `{0}`, `{1}`, … positional substitution (matches vscode.l10n.t's syntax). */
function applyArgs(template: string, args: L10nArg[]): string {
  if (args.length === 0) return template;
  return template.replace(/\{(\d+)\}/g, (_, raw) => {
    const i = Number(raw);
    return i < args.length ? String(args[i]) : `{${i}}`;
  });
}

/**
 * Translate a key, falling back to a hardcoded English string if the
 * active l10n bundle is unavailable or doesn't contain the key.
 *
 * Always prefer this over `vscode.l10n.t` so users never see literal
 * key strings like `picfast.ui.welcomeNotification` in the UI.
 */
export function t(key: string, ...args: L10nArg[]): string {
  // Lazy-require `vscode` so this module loads in unit-test contexts
  // where the `vscode` module is not present. If the require fails
  // (no `vscode` runtime) or `vscode.l10n` is missing, we treat the
  // translation as failed and fall back to the hardcoded English.
  let translated: string = key;
  try {
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const vscode = require("vscode") as typeof vscodeType;
    if (vscode?.l10n?.t) {
      const out = vscode.l10n.t(key, ...(args as L10nArg[]));
      if (typeof out === "string" && out.length > 0) {
        translated = out;
      }
    }
  } catch {
    // Either the `vscode` module is missing or `l10n.t` threw; the
    // `translated = key` default above is what we want to fall back to.
  }
  // vscode.l10n returns the key literally when the active bundle is
  // missing the key, or when l10n has not been initialised yet.
  if (translated === key || translated === "") {
    const fallback = EN_FALLBACK[key];
    return applyArgs(fallback ?? key, args);
  }
  // `vscode.l10n.t` already substitutes positional args in the happy
  // path, so we only need to apply them ourselves on the fallback.
  if (args.length > 0) return applyArgs(translated, args);
  return translated;
}

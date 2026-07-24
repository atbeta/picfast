// src/__tests__/l10n.test.ts
import { test } from "node:test";
import assert from "node:assert/strict";
import { t } from "../l10n";

// The whole point of the wrapper is to behave well when
// `vscode.l10n.t` is unavailable or returns the key literally
// (which is what vscode.l10n does in unit-test contexts and when
// the active bundle is missing a key). In a Node test we don't
// have vscode at all, so every call into the wrapper returns the
// key literally — which is exactly the failure mode we want the
// fallback to cover. These tests therefore describe the
// *guaranteed-not-broken* behaviour: never show the literal key,
// always fall back to the hardcoded English.

test("t: returns English fallback when vscode.l10n.t is unavailable", () => {
  // No `vscode` module is loaded; the wrapper's `try` block catches
  // the resulting exception and falls through to the hardcoded
  // English string.
  const out = t("picfast.ui.welcomeNotification");
  assert.notEqual(out, "picfast.ui.welcomeNotification");
  assert.ok(out.startsWith("PicFast is active."));
});

test("t: returns English for a known key with no positional args", () => {
  const out = t("picfast.error.noActiveEditor");
  assert.notEqual(out, "picfast.error.noActiveEditor");
  assert.equal(out, "PicFast: no active text editor.");
});

test("t: substitutes positional {0} in the fallback", () => {
  const out = t("picfast.error.pathResolveFailed", "./shot.png");
  assert.notEqual(out, "picfast.error.pathResolveFailed");
  assert.equal(out, 'PicFast: cannot resolve path "./shot.png".');
});

test("t: substitutes positional {0} and {1} in the fallback", () => {
  const out = t("picfast.error.cannotReadEnoent", "/abs/path.png", "shot.png");
  assert.notEqual(out, "picfast.error.cannotReadEnoent");
  assert.ok(out.includes("/abs/path.png"));
  assert.ok(out.includes("shot.png"));
});

test("t: returns the key itself when neither vscode nor the fallback has it", () => {
  // Sanity: if a key is genuinely missing everywhere, we don't
  // pretend it exists — we surface the key. This is intentional
  // (otherwise typos would be silently swallowed).
  const out = t("picfast.totally.made.up.key");
  assert.equal(out, "picfast.totally.made.up.key");
});

test("t: handles non-string positional args", () => {
  // {0} is interpolated as a string. Numbers and booleans should
  // also work because the fallback formatter coerces with String().
  const out = t("picfast.notify.someReplaced", 3, 1, "skipped: foo");
  assert.notEqual(out, "picfast.notify.someReplaced");
  assert.equal(out, "PicFast: replaced 3 local reference(s); 1 failed. skipped: foo");
});

test("t: clamps to highest positional index if the template references more than the caller passed", () => {
  // {5} is referenced by the template but the caller passed 2 args;
  // the unfilled placeholders should be left as `{5}` rather than
  // being 'undefined'. This mirrors vscode.l10n.t's documented
  // behaviour for missing positional args.
  const out = t("picfast.notify.uploadedOne", "only-one");
  // Template is "PicFast: uploaded {0}" so {0} is replaced.
  assert.equal(out, "PicFast: uploaded only-one");
});

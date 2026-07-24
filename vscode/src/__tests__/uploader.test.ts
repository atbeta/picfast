// src/__tests__/uploader.test.ts
// Unit tests for the pure helpers in uploader.ts.

import { test } from "node:test";
import assert from "node:assert/strict";
import { formatInsert } from "../uploader";

test("formatInsert: url returns the URL as-is", () => {
  assert.equal(
    formatInsert("https://picfast.xxx/i/abc.png", "url", "alt"),
    "https://picfast.xxx/i/abc.png",
  );
});

test("formatInsert: markdown wraps with ![alt](url)", () => {
  assert.equal(
    formatInsert("https://picfast.xxx/i/abc.png", "markdown", "shot"),
    "![shot](https://picfast.xxx/i/abc.png)",
  );
});

test("formatInsert: html wraps with <img>", () => {
  assert.equal(
    formatInsert("https://picfast.xxx/i/abc.png", "html", "shot"),
    '<img src="https://picfast.xxx/i/abc.png" alt="shot" />',
  );
});

test("formatInsert: bbcode wraps with [img]...[/img]", () => {
  assert.equal(
    formatInsert("https://picfast.xxx/i/abc.png", "bbcode", "shot"),
    "[img]https://picfast.xxx/i/abc.png[/img]",
  );
});

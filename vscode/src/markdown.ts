// src/markdown.ts
// Helpers to find local image references in markdown / HTML, and to
// resolve them to absolute filesystem paths relative to the document.

import * as path from "node:path";

export interface ImageRef {
  /** Character offsets into the source. */
  range: { start: number; end: number };
  kind: "markdown" | "html";
  alt: string;
  /** The raw path/URL string as it appears in the source. */
  rawPath: string;
  /** Path after `decodeURIComponent`, with leading/trailing whitespace stripped. */
  decodedPath: string;
}

// Match `![alt](path)` or `![alt](path "title")`. The path is
// captured non-greedily up to a closing `)` that is followed by
// either end-of-line, end-of-input, or an optional `"title"` — this
// lets the path contain `)` (e.g. `screenshot (1).png`) without the
// regex grabbing the inner paren.
const MARKDOWN_IMG_RE = /!\[([^\]]*)\]\(([\s\S]+?)\)(?=\s*(?:"[^"]*"\s*)?$)/gm;

// Match `<img ... src="path" ... />` (single or double quoted).
const HTML_IMG_SRC_RE = /<img\b[^>]*?\bsrc\s*=\s*(?:"([^"]+)"|'([^']+)')[^>]*\/?>/gi;

export function findImageRefs(text: string): ImageRef[] {
  const out: ImageRef[] = [];

  for (const m of text.matchAll(MARKDOWN_IMG_RE)) {
    const start = m.index ?? 0;
    // Pull the optional `"title"` off the end so the path is clean.
    const body = (m[2] ?? "").trim();
    const titleMatch = body.match(/^(.*?)\s+"([^"]+)"$/);
    const rawPath = titleMatch ? titleMatch[1] : body;
    out.push({
      range: { start, end: start + m[0].length },
      kind: "markdown",
      alt: m[1] ?? "",
      rawPath,
      decodedPath: decodePath(rawPath),
    });
  }

  for (const m of text.matchAll(HTML_IMG_SRC_RE)) {
    const start = m.index ?? 0;
    const rawPath = (m[2] ?? m[1] ?? "").trim();
    out.push({
      range: { start, end: start + m[0].length },
      kind: "html",
      alt: "",
      rawPath,
      decodedPath: decodePath(rawPath),
    });
  }

  out.sort((a, b) => a.range.start - b.range.start);
  return out;
}

/** Best-effort URL-decode; falls back to the trimmed raw path on error. */
export function decodePath(raw: string): string {
  const trimmed = raw.trim();
  try {
    return decodeURIComponent(trimmed);
  } catch {
    return trimmed;
  }
}

export function isLocalPath(raw: string): boolean {
  if (!raw) return false;
  if (raw.startsWith("http://") || raw.startsWith("https://")) return false;
  if (raw.startsWith("data:")) return false;
  return true;
}

/**
 * Build the list of candidate absolute paths to try, in priority order.
 * Each entry in `rawPaths` is resolved against `documentDir` and then
 * every entry in `workspaceFolders` (deduped across both dirs and
 * across the input path list). Pass `[decodedPath, rawPath]` from an
 * `ImageRef` so a path like `./shot%20(1).png` is tried both as
 * `./shot (1).png` (the user's actual intent) and as the literal
 * `./shot%20(1).png` (in case the file is *literally* named that).
 *
 * Absolute paths and `file://` URIs are returned as-is.
 *
 * Exposed separately so callers can show "tried these paths" in
 * error messages without re-implementing the resolution rules.
 */
export function candidatePathsFor(
  rawPaths: readonly string[],
  documentDir: string | null,
  workspaceFolders: readonly string[],
): string[] {
  if (rawPaths.length === 0) return [];

  const out: string[] = [];
  const seen = new Set<string>();
  const push = (candidate: string) => {
    if (seen.has(candidate)) return;
    seen.add(candidate);
    out.push(candidate);
  };

  for (const raw of rawPaths) {
    if (!raw) continue;
    if (raw.startsWith("file://")) {
      try {
        const { fileURLToPath } = require("node:url") as typeof import("node:url");
        push(fileURLToPath(raw));
      } catch {
        // ignore
      }
      continue;
    }
    if (path.isAbsolute(raw)) {
      push(raw);
      continue;
    }
    if (documentDir) push(path.resolve(documentDir, raw));
    for (const folder of workspaceFolders) {
      if (!folder) continue;
      push(path.resolve(folder, raw));
    }
  }
  return out;
}

/**
 * Resolve a list of candidate local paths to the first existing
 * absolute path. Tries the candidates in order; the first that exists
 * wins. Returns `null` if none do, or if no candidate could even be
 * built. When nothing exists, the function still returns the highest-
 * priority candidate so the caller can show a useful error message.
 */
export async function resolveLocalPath(
  rawPaths: readonly string[],
  documentDir: string | null,
  workspaceFolders: readonly string[],
): Promise<string | null> {
  const fs = require("node:fs/promises") as typeof import("node:fs/promises");
  const candidates = candidatePathsFor(rawPaths, documentDir, workspaceFolders);
  for (const candidate of candidates) {
    try {
      await fs.access(candidate);
      return candidate;
    } catch {
      // try the next candidate
    }
  }
  return candidates[0] ?? null;
}

const IMAGE_EXT_RE = /\.(png|jpg|jpeg|gif|webp|bmp|tif|tiff|avif|ico)$/i;

export function isImageFile(p: string): boolean {
  return IMAGE_EXT_RE.test(p);
}

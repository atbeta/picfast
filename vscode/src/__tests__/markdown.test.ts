// src/__tests__/markdown.test.ts
import { test } from "node:test";
import assert from "node:assert/strict";
import * as fs from "node:fs/promises";
import * as os from "node:os";
import * as path from "node:path";
import {
  candidatePathsFor,
  findImageRefs,
  isImageFile,
  isLocalPath,
  resolveLocalPath,
  decodePath,
} from "../markdown";

test("findImageRefs: simple markdown", () => {
  const text = "intro\n\n![alt](./shot.png)\n\noutro";
  const refs = findImageRefs(text);
  assert.equal(refs.length, 1);
  assert.equal(refs[0].kind, "markdown");
  assert.equal(refs[0].alt, "alt");
  assert.equal(refs[0].rawPath, "./shot.png");
  assert.equal(refs[0].decodedPath, "./shot.png");
});

test("findImageRefs: markdown with title", () => {
  const refs = findImageRefs("![a](p.png \"title\")");
  assert.equal(refs.length, 1);
  assert.equal(refs[0].rawPath, "p.png");
});

test("findImageRefs: markdown with spaces in path is preserved", () => {
  const refs = findImageRefs("![alt](./shot 1.png)");
  assert.equal(refs.length, 1);
  assert.equal(refs[0].rawPath, "./shot 1.png");
  assert.equal(refs[0].decodedPath, "./shot 1.png");
});

test("findImageRefs: URL-encoded path is decoded into decodedPath", () => {
  const refs = findImageRefs("![alt](./shot%20(1).png)");
  assert.equal(refs.length, 1);
  assert.equal(refs[0].rawPath, "./shot%20(1).png");
  assert.equal(refs[0].decodedPath, "./shot (1).png");
});

test("findImageRefs: URL-encoded path with invalid encoding falls back", () => {
  const refs = findImageRefs("![alt](./bad%ZZ.png)");
  assert.equal(refs.length, 1);
  assert.equal(refs[0].decodedPath, "./bad%ZZ.png");
});

test("findImageRefs: html img with double quotes", () => {
  const refs = findImageRefs(`<p><img src="./x.png" alt="x" /></p>`);
  assert.equal(refs.length, 1);
  assert.equal(refs[0].kind, "html");
  assert.equal(refs[0].rawPath, "./x.png");
});

test("findImageRefs: html img with single quotes", () => {
  const refs = findImageRefs(`<img src='./x.png' />`);
  assert.equal(refs.length, 1);
  assert.equal(refs[0].rawPath, "./x.png");
});

test("findImageRefs: html img with URL-encoded path is decoded", () => {
  const refs = findImageRefs(`<img src="./my%20pic.png" />`);
  assert.equal(refs.length, 1);
  assert.equal(refs[0].decodedPath, "./my pic.png");
});

test("findImageRefs: mixed markdown + html", () => {
  const text = "![a](./a.png)\n<img src=\"./b.png\" />\n![c](./c.png \"t\")";
  const refs = findImageRefs(text);
  assert.equal(refs.length, 3);
  assert.equal(refs.map((r) => r.rawPath).join(","), "./a.png,./b.png,./c.png");
});

test("isLocalPath: false for http(s)/data", () => {
  assert.equal(isLocalPath("https://example.com/x.png"), false);
  assert.equal(isLocalPath("http://x"), false);
  assert.equal(isLocalPath("data:image/png;base64,..."), false);
});

test("isLocalPath: true for relative/absolute/file", () => {
  assert.equal(isLocalPath("./x.png"), true);
  assert.equal(isLocalPath("/abs/x.png"), true);
  assert.equal(isLocalPath("file:///x.png"), true);
  assert.equal(isLocalPath(""), false);
});

test("resolveLocalPath: relative against documentDir", async () => {
  const out = await resolveLocalPath(["./shot.png"], "/tmp/blog", []);
  assert.equal(out, path.resolve("/tmp/blog", "./shot.png"));
});

test("resolveLocalPath: absolute returns as-is", async () => {
  assert.equal(
    await resolveLocalPath(["/etc/shot.png"], "/tmp/blog", []),
    "/etc/shot.png",
  );
});

test("resolveLocalPath: relative without documentDir returns null", async () => {
  assert.equal(await resolveLocalPath(["./x.png"], null, []), null);
});

test("resolveLocalPath: file:// URI is resolved", async () => {
  const out = await resolveLocalPath(["file:///tmp/x.png"], "/whatever", []);
  assert.equal(out, "/tmp/x.png");
});

test("resolveLocalPath: handles URL-encoded absolute paths", async () => {
  const out = await resolveLocalPath(
    ["/abs/shot%20(1).png"],
    "/tmp/blog",
    [],
  );
  assert.equal(out, "/abs/shot%20(1).png");
});

test("resolveLocalPath: falls back to workspace folder when not in document dir", async () => {
  const wsRoot = await fs.mkdtemp(path.join(os.tmpdir(), "picfast-ws-"));
  try {
    await fs.writeFile(path.join(wsRoot, "logo.png"), "fake");
    const out = await resolveLocalPath(["logo.png"], "/tmp/blog", [wsRoot]);
    assert.equal(out, path.join(wsRoot, "logo.png"));
  } finally {
    await fs.rm(wsRoot, { recursive: true, force: true });
  }
});

test("resolveLocalPath: document dir takes priority over workspace", async () => {
  const wsRoot = await fs.mkdtemp(path.join(os.tmpdir(), "picfast-ws-"));
  const docDir = await fs.mkdtemp(path.join(os.tmpdir(), "picfast-doc-"));
  try {
    await fs.writeFile(path.join(wsRoot, "shared.png"), "from-workspace");
    await fs.writeFile(path.join(docDir, "shared.png"), "from-document");
    const out = await resolveLocalPath(["shared.png"], docDir, [wsRoot]);
    assert.equal(out, path.join(docDir, "shared.png"));
  } finally {
    await fs.rm(wsRoot, { recursive: true, force: true });
    await fs.rm(docDir, { recursive: true, force: true });
  }
});

test("resolveLocalPath: returns first candidate even when none exist (for error message)", async () => {
  const wsRoot = await fs.mkdtemp(path.join(os.tmpdir(), "picfast-ws-"));
  try {
    const out = await resolveLocalPath(["nope.png"], "/tmp/blog", [wsRoot]);
    assert.equal(out, path.resolve("/tmp/blog", "nope.png"));
  } finally {
    await fs.rm(wsRoot, { recursive: true, force: true });
  }
});

test("resolveLocalPath: URL-encoded path resolves to the decoded file (Beta's bug)", async () => {
  // The bug Beta reported: user writes `![](./shot%201.png)` in
  // markdown but the actual file on disk is named `shot 1.png`. The
  // resolver must try the URL-decoded form so the file is found.
  // Callers in the extension pass BOTH the raw and decoded paths
  // (the rawPath and decodedPath fields on ImageRef), so we test
  // the same way here.
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), "picfast-decoded-"));
  try {
    await fs.writeFile(path.join(dir, "shot 1.png"), "fake");
    const out = await resolveLocalPath(
      ["./shot%201.png", "./shot 1.png"],
      dir,
      [],
    );
    assert.equal(out, path.join(dir, "shot 1.png"));
  } finally {
    await fs.rm(dir, { recursive: true, force: true });
  }
});

test("resolveLocalPath: raw path wins if the file is literally named that", async () => {
  // Edge case: the user's filename genuinely contains a literal "%20".
  // The literal path should still resolve.
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), "picfast-literal-"));
  try {
    await fs.writeFile(path.join(dir, "weird%20name.png"), "fake");
    const out = await resolveLocalPath(["./weird%20name.png"], dir, []);
    assert.equal(out, path.join(dir, "weird%20name.png"));
  } finally {
    await fs.rm(dir, { recursive: true, force: true });
  }
});

test("candidatePathsFor: priority order is document dir then workspace folders", () => {
  const paths = candidatePathsFor(
    ["./shot.png"],
    "/tmp/blog",
    ["/tmp/ws1", "/tmp/ws2"],
  );
  assert.deepEqual(paths, [
    path.resolve("/tmp/blog", "./shot.png"),
    path.resolve("/tmp/ws1", "./shot.png"),
    path.resolve("/tmp/ws2", "./shot.png"),
  ]);
});

test("candidatePathsFor: deduplicates when document dir is also a workspace folder", () => {
  const paths = candidatePathsFor(
    ["./shot.png"],
    "/tmp/ws",
    ["/tmp/ws", "/tmp/other"],
  );
  assert.deepEqual(paths, [
    path.resolve("/tmp/ws", "./shot.png"),
    path.resolve("/tmp/other", "./shot.png"),
  ]);
});

test("candidatePathsFor: returns absolute / file:// directly", () => {
  assert.deepEqual(
    candidatePathsFor(["/abs/x.png"], "/tmp/blog", ["/tmp/ws"]),
    ["/abs/x.png"],
  );
  assert.deepEqual(
    candidatePathsFor(["file:///tmp/x.png"], "/tmp/blog", []),
    ["/tmp/x.png"],
  );
});

test("candidatePathsFor: includes both decoded and raw variants", () => {
  // For an image with literal spaces in the markdown, the resolver
  // must try the URL-decoded path (which is the user's actual intent)
  // AND the raw path (in case the literal `%20` is the real filename).
  const paths = candidatePathsFor(
    ["./shot%201.png", "./shot 1.png"],
    "/tmp/blog",
    [],
  );
  // First the first input against document dir, then the second input
  // against document dir.
  assert.ok(paths.includes(path.resolve("/tmp/blog", "./shot 1.png")));
  assert.ok(paths.includes(path.resolve("/tmp/blog", "./shot%201.png")));
});

test("isImageFile: recognized extensions", () => {
  assert.equal(isImageFile("a.png"), true);
  assert.equal(isImageFile("a.JPG"), true);
  assert.equal(isImageFile("a.webp"), true);
  assert.equal(isImageFile("a.svg"), false);
  assert.equal(isImageFile("a.txt"), false);
});

test("decodePath: decodes valid encoding", () => {
  assert.equal(decodePath("hello%20world.png"), "hello world.png");
});

test("decodePath: leaves non-encoded path alone", () => {
  assert.equal(decodePath("./shot.png"), "./shot.png");
});

test("decodePath: returns trimmed raw on decode error", () => {
  assert.equal(decodePath("  bad%ZZ  "), "bad%ZZ");
});

// src/workspaceFolders.ts
// Helper to read VSCode's multi-root workspace folders as a list of
// absolute filesystem paths. Returns an empty array when no workspace
// is open (e.g. the user opened a single loose file).

import * as vscode from "vscode";

export function getWorkspaceFolderPaths(): string[] {
  const folders = vscode.workspace.workspaceFolders ?? [];
  return folders
    .map((f) => f.uri.scheme === "file" ? f.uri.fsPath : null)
    .filter((p): p is string => p !== null);
}

// src/commands/uploadLocalImages.ts
// Scan the current editor for local image references in markdown / HTML,
// upload each, and replace the local path with the remote URL.

import * as fs from "node:fs/promises";
import * as path from "node:path";
import * as vscode from "vscode";
import { t } from "../l10n";
import { findImageRefs, isImageFile, isLocalPath, resolveLocalPath, type ImageRef } from "../markdown";
import { getConfig } from "../config";
import { uploadImage, UploadError } from "../uploader";
import { getWorkspaceFolderPaths } from "../workspaceFolders";

const COMMAND = "picfast.uploadLocalImages";

interface PendingUpload {
  ref: ImageRef;
  absolutePath: string;
}

interface FailedUpload {
  ref: ImageRef;
  reason: string;
}

export function registerUploadLocalImagesCommand(): vscode.Disposable {
  return vscode.commands.registerCommand(COMMAND, async () => {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
      void vscode.window.showWarningMessage(
        t("picfast.error.noActiveEditor"),
      );
      return;
    }

    const config = getConfig();
    if (!config.baseUrl) {
      const action = await vscode.window.showErrorMessage(
        t("picfast.error.baseUrlNotSet"),
        t("picfast.error.openSettings"),
      );
      if (action === t("picfast.error.openSettings")) {
        void vscode.commands.executeCommand(
          "workbench.action.openSettings",
          "picfast.baseUrl",
        );
      }
      return;
    }

    const documentDir = editor.document.uri.scheme === "file"
      ? path.dirname(editor.document.uri.fsPath)
      : null;
    const workspaceFolders = getWorkspaceFolderPaths();

    const sourceText = editor.document.getText();
    const refs = findImageRefs(sourceText);
    const localRefs = refs.filter((r) => isLocalPath(r.rawPath));

    if (localRefs.length === 0) {
      void vscode.window.showInformationMessage(
        t("picfast.notify.noLocalRefs"),
      );
      return;
    }

    const pending: PendingUpload[] = [];
    const skipped: FailedUpload[] = [];

    for (const ref of localRefs) {
      const abs = await resolveLocalPath(
        [ref.decodedPath, ref.rawPath],
        documentDir,
        workspaceFolders,
      );
      if (!abs) {
        skipped.push({
          ref,
          reason: documentDir
            ? "cannot resolve path"
            : "unsaved document; relative path not resolvable",
        });
        continue;
      }
      if (!isImageFile(abs)) {
        skipped.push({ ref, reason: "not an image extension" });
        continue;
      }
      try {
        const stat = await fs.stat(abs);
        if (!stat.isFile()) {
          skipped.push({ ref, reason: "not a regular file" });
          continue;
        }
      } catch (err) {
        const e = err as NodeJS.ErrnoException;
        const hint = e.code === "ENOENT"
          ? " (file does not exist at the resolved path)"
          : "";
        skipped.push({ ref, reason: `${e.code ?? "stat error"}${hint}` });
        continue;
      }
      pending.push({ ref, absolutePath: abs });
    }

    if (pending.length === 0) {
      const summary = formatSkippedSummary(skipped);
      void vscode.window.showWarningMessage(
        t("picfast.notify.noneUploadable", localRefs.length, summary),
      );
      return;
    }

    const successes = new Map<number, string>();
    const failures: FailedUpload[] = [...skipped];

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: "PicFast: uploading local images",
        cancellable: true,
      },
      async (progress, token) => {
        for (let i = 0; i < pending.length; i++) {
          if (token.isCancellationRequested) break;
          const p = pending[i];
          const fileName = path.basename(p.absolutePath);
          progress.report({
            message: `${i + 1}/${pending.length} ${fileName}`,
            increment: 100 / pending.length,
          });
          try {
            const data = await fs.readFile(p.absolutePath);
            const result = await uploadImage({
              baseUrl: config.baseUrl,
              apiToken: config.apiToken || undefined,
              timeoutMs: config.timeoutMs,
              fileName,
              data: new Uint8Array(data),
            });
            successes.set(p.ref.range.start, result.url);
          } catch (err) {
            failures.push({
              ref: p.ref,
              reason: err instanceof UploadError
                ? err.message
                : err instanceof Error
                  ? err.message
                  : String(err),
            });
          }
        }
      },
    );

    if (successes.size === 0) {
      void vscode.window.showErrorMessage(
        t("picfast.notify.allUploadFailed", pending.length, formatSkippedSummary(failures)),
      );
      return;
    }

    const edit = new vscode.WorkspaceEdit();
    const sortedStarts = Array.from(successes.keys()).sort((a, b) => b - a);
    for (const start of sortedStarts) {
      const ref = pending.find((p) => p.ref.range.start === start)!.ref;
      const url = successes.get(start)!;
      const newText = rebuildRef(ref, url);
      const startPos = editor.document.positionAt(start);
      const endPos = editor.document.positionAt(ref.range.end);
      edit.replace(editor.document.uri, new vscode.Range(startPos, endPos), newText);
    }
    const applied = await vscode.workspace.applyEdit(edit);

    if (!applied) {
      void vscode.window.showErrorMessage(
        t("picfast.error.applyEditFailedBulk"),
      );
      return;
    }

    const failedCount = failures.length;
    if (failedCount > 0) {
      void vscode.window.showWarningMessage(
        t(
          "picfast.notify.someReplaced",
          successes.size,
          failedCount,
          formatSkippedSummary(failures),
        ),
      );
    } else {
      void vscode.window.setStatusBarMessage(
        t("picfast.notify.allReplaced", successes.size),
        3000,
      );
    }
  });
}

function rebuildRef(ref: ImageRef, newUrl: string): string {
  if (ref.kind === "markdown") {
    return `![${ref.alt}](${newUrl})`;
  }
  return `<img src="${newUrl}" />`;
}

function formatSkippedSummary(failures: FailedUpload[]): string {
  if (failures.length === 0) return "";
  const first = failures
    .slice(0, 3)
    .map((f) => `"${f.ref.rawPath}" (${f.reason})`);
  const more = failures.length > 3 ? `, +${failures.length - 3} more` : "";
  return `Skipped: ${first.join(", ")}${more}`;
}

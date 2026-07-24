// src/commands/uploadOneImage.ts
// Upload a single image reference and replace it in place. Invoked when
// the user clicks a CodeLens above a local image reference.

import * as fs from "node:fs/promises";
import * as path from "node:path";
import * as vscode from "vscode";
import { t } from "../l10n";
import {
  candidatePathsFor,
  findImageRefs,
  isImageFile,
  resolveLocalPath,
  type ImageRef,
} from "../markdown";
import { getConfig } from "../config";
import { uploadImage, UploadError } from "../uploader";
import { getWorkspaceFolderPaths } from "../workspaceFolders";

const COMMAND = "picfast.uploadOneImage";

export function registerUploadOneImageCommand(): vscode.Disposable {
  return vscode.commands.registerCommand(
    COMMAND,
    async (ref: ImageRef, targetUri: vscode.Uri) => {
      const document = vscode.workspace.textDocuments.find(
        (d) => d.uri.toString() === targetUri.toString(),
      );
      if (!document) {
        void vscode.window.showErrorMessage(
          t("picfast.error.targetDocClosed"),
        );
        return;
      }

      const live = findImageRefs(document.getText()).find(
        (r) => r.range.start === ref.range.start,
      );
      const target = live ?? ref;

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

      const documentDir = document.uri.scheme === "file"
        ? path.dirname(document.uri.fsPath)
        : null;
      const workspaceFolders = getWorkspaceFolderPaths();

      const candidates = candidatePathsFor(
        [target.decodedPath, target.rawPath],
        documentDir,
        workspaceFolders,
      );
      if (candidates.length === 0) {
        const reason = documentDir
          ? t("picfast.error.pathResolveFailed", target.rawPath)
          : t("picfast.error.pathResolveFailedUnsaved", target.rawPath);
        void vscode.window.showErrorMessage(reason);
        return;
      }

      const abs = await resolveLocalPath(
        [target.decodedPath, target.rawPath],
        documentDir,
        workspaceFolders,
      );
      if (!abs) {
        void vscode.window.showErrorMessage(
          formatNotFoundError(target.rawPath, candidates, documentDir),
        );
        return;
      }

      if (!isImageFile(abs)) {
        void vscode.window.showErrorMessage(
          t("picfast.error.notImageExtension", target.rawPath),
        );
        return;
      }

      const fileName = path.basename(abs);
      try {
        const data = await fs.readFile(abs);
        const result = await uploadImage({
          baseUrl: config.baseUrl,
          apiToken: config.apiToken || undefined,
          timeoutMs: config.timeoutMs,
          fileName,
          data: new Uint8Array(data),
        });

        const alt = fileName.replace(/\.[^.]+$/, "");
        const newText = rebuildRef(target, result.url);
        const startPos = document.positionAt(target.range.start);
        const endPos = document.positionAt(target.range.end);
        const edit = new vscode.WorkspaceEdit();
        edit.replace(document.uri, new vscode.Range(startPos, endPos), newText);
        const applied = await vscode.workspace.applyEdit(edit);

        if (!applied) {
          void vscode.window.showErrorMessage(
            t("picfast.error.applyEditFailed", fileName),
          );
          return;
        }

        void vscode.window.setStatusBarMessage(
          t("picfast.notify.uploadedOne", fileName),
          3000,
        );
      } catch (err) {
        const e = err instanceof UploadError ? err : err;
        const message = err instanceof UploadError
          ? `${err.message}${err.body ? `\n${err.body.slice(0, 200)}` : ""}`
          : err instanceof Error
            ? err.message
            : String(err);
        void vscode.window.showErrorMessage(
          t("picfast.error.uploadFailed", message),
        );
      }
    },
  );
}

function rebuildRef(ref: ImageRef, newUrl: string): string {
  if (ref.kind === "markdown") {
    return `![${ref.alt}](${newUrl})`;
  }
  return `<img src="${newUrl}" />`;
}

function formatNotFoundError(
  rawPath: string,
  candidates: string[],
  documentDir: string | null,
): string {
  const sample = candidates.slice(0, 3).join("\n  • ");
  const more = candidates.length > 3 ? `\n  • …and ${candidates.length - 3} more` : "";
  const docHint = documentDir
    ? t("picfast.error.notFoundDocDir", documentDir)
    : t("picfast.error.notFoundUnsaved");
  const hint = t("picfast.error.notFoundHint");
  return (
    t("picfast.error.notFoundTitle", rawPath) +
    "\n" +
    t("picfast.error.notFoundTried") +
    "\n  • " +
    sample +
    more +
    docHint +
    hint
  );
}

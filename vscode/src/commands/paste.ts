// src/commands/paste.ts
// Core command: pick a local image file, upload to PicFast, insert the
// resulting link into the active editor at the cursor.

import * as fs from "node:fs/promises";
import * as path from "node:path";
import * as vscode from "vscode";
import { t } from "../l10n";
import { getConfig } from "../config";
import { formatInsert, uploadImage, UploadError } from "../uploader";

const COMMAND = "picfast.pasteImage";

const IMAGE_EXTENSIONS = [
  "png", "jpg", "jpeg", "gif", "webp", "bmp", "tif", "tiff", "avif", "ico",
] as const;

const PICK_FILE_FILTERS: readonly string[] = [...IMAGE_EXTENSIONS, "*"];

export function registerPasteCommand(deps: {
  setStatus: (text: string, kind: "idle" | "busy" | "ok" | "error") => void;
}): vscode.Disposable {
  return vscode.commands.registerCommand(COMMAND, async (uri?: vscode.Uri) => {
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

    deps.setStatus("$(loading) PicFast: selecting image…", "busy");

    let selected: vscode.Uri | undefined = uri;
    if (!selected) {
      const picked = await vscode.window.showOpenDialog({
        title: t("picfast.command.pasteImage"),
        filters: PICK_FILE_FILTERS as unknown as vscode.OpenDialogOptions["filters"],
        canSelectMany: false,
      });
      if (!picked || picked.length === 0) {
        deps.setStatus("$(cloud-upload) PicFast", "idle");
        return;
      }
      selected = picked[0];
    }

    if (!isImageFile(selected.fsPath)) {
      deps.setStatus("$(error) PicFast: not an image", "error");
      void vscode.window.showWarningMessage(
        t("picfast.error.notSupportedFormat", path.basename(selected.fsPath)),
      );
      return;
    }

    let data: Buffer;
    try {
      data = await fs.readFile(selected.fsPath);
    } catch (err) {
      deps.setStatus("$(error) PicFast: read error", "error");
      void vscode.window.showErrorMessage(
        t("picfast.error.fileReadFailed", (err as Error).message),
      );
      return;
    }

    const fileName = path.basename(selected.fsPath);
    deps.setStatus(`$(loading) PicFast: uploading ${fileName}…`, "busy");

    try {
      const result = await uploadImage({
        baseUrl: config.baseUrl,
        apiToken: config.apiToken || undefined,
        timeoutMs: config.timeoutMs,
        fileName,
        data: new Uint8Array(data),
      });

      const alt = fileName.replace(/\.[^.]+$/, "");
      const insertText = formatInsert(result.url, config.defaultFormat, alt);
      await editor.edit((edit) => {
        edit.insert(editor.selection.active, insertText);
      });

      deps.setStatus("$(check) PicFast: uploaded", "ok");
      void vscode.window.setStatusBarMessage(
        t("picfast.notify.uploadedOne", fileName),
        3000,
      );
    } catch (err) {
      deps.setStatus("$(error) PicFast: upload failed", "error");
      const message = err instanceof UploadError
        ? `${err.message}${err.body ? `\n${err.body.slice(0, 200)}` : ""}`
        : err instanceof Error
          ? err.message
          : String(err);
      void vscode.window.showErrorMessage(
        t("picfast.error.uploadFailed", message),
      );
    }
  });
}

function isImageFile(p: string): boolean {
  const ext = path.extname(p).toLowerCase().replace(/^\./, "");
  return (IMAGE_EXTENSIONS as readonly string[]).includes(ext);
}

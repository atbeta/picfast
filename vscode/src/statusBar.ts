// src/statusBar.ts
// Tiny status bar helper that the paste command pokes to reflect state.

import * as vscode from "vscode";
import { t } from "./l10n";

type StatusKind = "idle" | "busy" | "ok" | "error";

export interface StatusBar {
  set(text: string, kind: StatusKind): void;
  show(): void;
  hide(): void;
  dispose(): vscode.Disposable;
}

export function createStatusBar(initialVisible: boolean): StatusBar {
  const item = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100,
  );
  item.text = t("picfast.ui.statusBar.idle");
  item.command = "picfast.pasteImage";
  item.tooltip = t("picfast.ui.statusBar.tooltip");
  if (initialVisible) {
    item.show();
  }

  return {
    set(text: string, _kind: StatusKind): void {
      item.text = text;
      item.show();
    },
    show(): void {
      item.show();
    },
    hide(): void {
      item.hide();
    },
    dispose(): vscode.Disposable {
      item.dispose();
      return { dispose: () => {} };
    },
  };
}

// src/extension.ts
// Extension entry point. Wires up commands, CodeLens provider, status bar,
// configuration-change listener, and a one-time welcome notification.

import * as vscode from "vscode";
import { t } from "./l10n";
import { registerPasteCommand } from "./commands/paste";
import { registerUploadLocalImagesCommand } from "./commands/uploadLocalImages";
import { registerUploadOneImageCommand } from "./commands/uploadOneImage";
import { registerConfigCommands } from "./commands/configure";
import { createStatusBar } from "./statusBar";
import { getConfig } from "./config";
import { registerCodeLensProvider } from "./codeLens";

const WELCOME_SHOWN_KEY = "picfast.welcome.shown";

export function activate(context: vscode.ExtensionContext): void {
  const config = getConfig();
  const status = createStatusBar(config.showStatusBar);

  context.subscriptions.push(
    status,
    registerPasteCommand({
      setStatus: (text, _kind) => status.set(text, _kind),
    }),
    registerUploadLocalImagesCommand(),
    registerUploadOneImageCommand(),
    ...registerConfigCommands(),
    registerCodeLensProvider(context),
  );

  // React to config changes for the showStatusBar toggle.
  context.subscriptions.push(
    vscode.workspace.onDidChangeConfiguration((e) => {
      if (e.affectsConfiguration("picfast.showStatusBar")) {
        const next = getConfig().showStatusBar;
        if (next) {
          status.show();
        } else {
          status.hide();
        }
      }
    }),
  );

  // One-time welcome notification on first activation, explaining the
  // discoverable entry points (CodeLens + command palette + shortcut).
  if (!context.globalState.get(WELCOME_SHOWN_KEY)) {
    void context.globalState.update(WELCOME_SHOWN_KEY, true);
    const openSettings = t("picfast.error.openSettings");
    void vscode.window
      .showInformationMessage(
        t("picfast.ui.welcomeNotification"),
        openSettings,
      )
      .then((choice) => {
        if (choice === openSettings) {
          void vscode.commands.executeCommand("picfast.setBaseUrl");
        }
      });
  }
}

export function deactivate(): void {
  // Nothing to clean up; subscriptions handle their own disposal.
}

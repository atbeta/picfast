// src/commands/configure.ts
// Helper commands for setting baseUrl / apiToken / opening docs.

import * as vscode from "vscode";
import { t } from "../l10n";
import { CONFIG_SECTION } from "../config";

export function registerConfigCommands(): vscode.Disposable[] {
  return [
    vscode.commands.registerCommand("picfast.setBaseUrl", async () => {
      const current = vscode.workspace.getConfiguration(CONFIG_SECTION).get<string>("baseUrl") ?? "";
      const value = await vscode.window.showInputBox({
        title: t("picfast.input.baseUrl.title"),
        prompt: t("picfast.input.baseUrl.prompt"),
        value: current,
        placeHolder: t("picfast.input.baseUrl.placeholder"),
        validateInput: (v) => {
          if (!v.trim()) return t("picfast.input.baseUrl.empty");
          if (!/^https?:\/\/.+/.test(v.trim())) {
            return t("picfast.input.baseUrl.invalid");
          }
          return null;
        },
      });
      if (value !== undefined) {
        await vscode.workspace
          .getConfiguration(CONFIG_SECTION)
          .update("baseUrl", value.trim(), vscode.ConfigurationTarget.Global);
        void vscode.window.showInformationMessage(
          t("picfast.notify.baseUrlSet", value.trim()),
        );
      }
    }),

    vscode.commands.registerCommand("picfast.setApiToken", async () => {
      const value = await vscode.window.showInputBox({
        title: t("picfast.input.token.title"),
        prompt: t("picfast.input.token.prompt"),
        password: true,
        placeHolder: t("picfast.input.token.placeholder"),
      });
      if (value !== undefined) {
        await vscode.workspace
          .getConfiguration(CONFIG_SECTION)
          .update("apiToken", value, vscode.ConfigurationTarget.Global);
        void vscode.window.showInformationMessage(
          value
            ? t("picfast.notify.tokenSet")
            : t("picfast.notify.tokenCleared"),
        );
      }
    }),

    vscode.commands.registerCommand("picfast.openDocs", async () => {
      await vscode.env.openExternal(vscode.Uri.parse("https://github.com/atbeta/picfast"));
    }),
  ];
}

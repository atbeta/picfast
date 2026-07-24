// src/codeLens.ts
// CodeLens provider that surfaces an "Upload to PicFast" link above every
// local image reference in a markdown/HTML file. This is the main
// discoverability mechanism for the extension — users see the link the
// moment they open a file that contains `![alt](./local.png)` and don't
// have to remember a keyboard shortcut.

import * as vscode from "vscode";
import { t } from "./l10n";
import { findImageRefs, isLocalPath } from "./markdown";

const COMMAND_UPLOAD_ONE = "picfast.uploadOneImage";

class PicFastCodeLensProvider implements vscode.CodeLensProvider {
  private readonly onChangeEmitter = new vscode.EventEmitter<void>();

  readonly onDidChangeCodeLenses = this.onChangeEmitter.event;

  refresh(): void {
    this.onChangeEmitter.fire();
  }

  provideCodeLenses(document: vscode.TextDocument): vscode.CodeLens[] {
    // Only surface lenses in `.md` / `.mdx` documents. HTML is
    // excluded: VS Code's own Markdown preview and image-paste
    // extensions already handle `<img src="…">` there, and surfacing
    // a PicFast action in raw `.html` files invites accidental
    // uploads of unrelated paths.
    const ext = document.uri.fsPath.toLowerCase();
    if (!ext.endsWith(".md") && !ext.endsWith(".mdx")) {
      return [];
    }

    const text = document.getText();
    const refs = findImageRefs(text);
    const lenses: vscode.CodeLens[] = [];
    for (const ref of refs) {
      if (!isLocalPath(ref.rawPath)) continue;
      const start = document.positionAt(ref.range.start);
      const end = document.positionAt(ref.range.end);
      const range = new vscode.Range(
        // Anchor the lens on the line above the reference so the click
        // target is unambiguous and the lens doesn't overlap the
        // reference text itself.
        start.with(start.line, 0),
        end,
      );
      lenses.push(
        new vscode.CodeLens(range, {
          title: t("picfast.ui.codeLens.title"),
          command: COMMAND_UPLOAD_ONE,
          arguments: [ref, document.uri],
          tooltip:
            t("picfast.ui.codeLens.tooltip", ref.rawPath) +
            "\n\n" +
            t("picfast.ui.codeLens.rebindTip"),
        }),
      );
    }
    return lenses;
  }
}

export function registerCodeLensProvider(context: vscode.ExtensionContext): vscode.Disposable {
  const provider = new PicFastCodeLensProvider();
  const disposable = vscode.languages.registerCodeLensProvider(
    [
      { pattern: "**/*.md", scheme: "file" },
      { pattern: "**/*.mdx", scheme: "file" },
    ],
    provider,
  );

  // Refresh lenses when the document content or config changes.
  context.subscriptions.push(
    disposable,
    vscode.workspace.onDidChangeTextDocument((e) => {
      if (isRelevantDoc(e.document)) provider.refresh();
    }),
    vscode.workspace.onDidChangeConfiguration((e) => {
      if (e.affectsConfiguration("picfast.baseUrl")) provider.refresh();
    }),
  );

  return disposable;
}

function isRelevantDoc(doc: vscode.TextDocument): boolean {
  const ext = doc.uri.fsPath.toLowerCase();
  return ext.endsWith(".md") || ext.endsWith(".mdx");
}

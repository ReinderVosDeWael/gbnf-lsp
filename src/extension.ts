import * as path from "path";
import * as vscode from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from "vscode-languageclient/node";
import * as fs from "fs";

let client: LanguageClient;
let outputChannel: vscode.OutputChannel;

export function activate(context: vscode.ExtensionContext) {
  // Create output channel for logging
  outputChannel = vscode.window.createOutputChannel("GBNF LSP");
  outputChannel.show();

  outputChannel.appendLine("Extension activating...");

  try {
    // Path to your Go LSP executable
    const serverPath = path.join(context.extensionPath, "/build/go-src.exe");

    // Make sure the LSP executable exists and is executable
    if (!fs.existsSync(serverPath)) {
      outputChannel.appendLine(`ERROR: LSP server not found at ${serverPath}`);
      vscode.window.showErrorMessage(
        `GBNF LSP server not found at ${serverPath}`
      );
      return;
    }

    outputChannel.appendLine(`Found LSP server at: ${serverPath}`);

    // Server options - defining how to start your LSP
    const serverOptions: ServerOptions = {
      run: { command: serverPath, args: [], options: { shell: true } },
      debug: {
        command: serverPath,
        args: ["--debug"],
        options: { shell: true },
      },
    };

    // Client options
    const clientOptions: LanguageClientOptions = {
      documentSelector: [{ scheme: "file", language: "gbnf" }],
      outputChannel: outputChannel,
    };

    // Create client
    client = new LanguageClient(
      "gbnfLanguageServer",
      "GBNF Language Server",
      serverOptions,
      clientOptions
    );

    outputChannel.appendLine("Starting LSP client...");

    // Start client with proper error handling
    client.start().then(
      () => {
        outputChannel.appendLine("LSP client started successfully");
        vscode.window.showInformationMessage("GBNF LSP connected successfully");
      },
      (error) => {
        outputChannel.appendLine(`Error starting LSP client: ${error.message}`);
        outputChannel.appendLine(error.stack || "No stack trace available");
        vscode.window.showErrorMessage(
          `Failed to start GBNF LSP: ${error.message}`
        );
      }
    );
  } catch (error: any) {
    outputChannel.appendLine(`Activation error: ${error.message}`);
    outputChannel.appendLine(error.stack || "No stack trace available");
    vscode.window.showErrorMessage(
      `GBNF LSP activation error: ${error.message}`
    );
  }
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}

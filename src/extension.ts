import * as os from "os";
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

function getPlatformBinary(context: vscode.ExtensionContext): string {
  const supportedPlatforms = ["win32", "linux", "darwin"];
  if (supportedPlatforms.find((plat) => plat === os.platform()) === undefined) {
    throw new Error(`Unsupported platform: ${os.platform()}`);
  }

  const supportedArchitectures = ["x64", "arm64"];
  if (supportedArchitectures.find((plat) => plat === os.arch()) === undefined) {
    throw new Error(`Unsupported platform: ${os.arch()}`);
  }

  let binName = os.platform() + "-" + os.arch() + "-" + "gbnf-engine";
  if (os.platform() === "win32") {
    binName += ".exe";
  }

  return path.join(context.extensionPath, "bin", binName);
}

export function activate(context: vscode.ExtensionContext) {
  // Create output channel for logging
  outputChannel = vscode.window.createOutputChannel("GBNF LSP");
  outputChannel.show();

  outputChannel.appendLine("Extension activating...");

  try {
    const serverPath = getPlatformBinary(context);

    if (!fs.existsSync(serverPath)) {
      outputChannel.appendLine(`ERROR: LSP server not found at ${serverPath}`);
      vscode.window.showErrorMessage(
        `GBNF LSP server not found at ${serverPath}`
      );
      return;
    }

    outputChannel.appendLine(`Found LSP server at: ${serverPath}`);

    const serverOptions: ServerOptions = {
      run: { command: serverPath, args: [], options: { shell: true } },
      debug: {
        command: serverPath,
        args: ["--debug"],
        options: { shell: true },
      },
    };

    const clientOptions: LanguageClientOptions = {
      documentSelector: [{ scheme: "file", language: "gbnf" }],
      outputChannel: outputChannel,
      traceOutputChannel: outputChannel,
    };

    client = new LanguageClient(
      "gbnfLanguageServer",
      "GBNF Language Server",
      serverOptions,
      clientOptions
    );

    outputChannel.appendLine("Starting LSP client...");

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

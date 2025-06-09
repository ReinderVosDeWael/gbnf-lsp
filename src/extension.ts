import * as os from "os";
import * as path from "path";
import * as vscode from "vscode";
import { https } from "follow-redirects";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from "vscode-languageclient/node";
import * as fs from "fs";
import { promisify } from "util";

const chmod = promisify(fs.chmod);
let client: LanguageClient;
let outputChannel: vscode.OutputChannel;

async function downloadFile(url: string, dest: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https
      .get(url, (response) => {
        if (response.statusCode !== 200) {
          return reject(
            new Error(`Failed to get '${url}' (${response.statusCode})`)
          );
        }
        response.pipe(file);
        file.on("finish", () => {
          file.close(async () => {
            try {
              await chmod(dest, 0o755);
              resolve();
            } catch (err) {
              reject(err);
            }
          });
        });
      })
      .on("error", (err) => {
        fs.unlink(dest, () => reject(err));
      });
  });
}

function getExtensionVersion(context: vscode.ExtensionContext): string {
  const packageJsonPath = path.join(context.extensionPath, "package.json");
  if (!fs.existsSync(packageJsonPath)) {
    throw new Error(`package.json not found at ${packageJsonPath}`);
  }

  const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, "utf-8"));
  if (!packageJson.version) {
    throw new Error("Version not found in package.json");
  }

  return `v${packageJson.version}`;
}

async function getPlatformBinary(
  context: vscode.ExtensionContext
): Promise<string> {
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

  const bindir = path.join(context.extensionPath, "bin");
  if (!fs.existsSync(bindir)) {
    fs.mkdirSync(bindir);
  }
  const binpath = path.join(bindir, binName);
  if (!fs.existsSync(binpath)) {
    const version = getExtensionVersion(context);
    const url = `https://github.com/ReinderVosDeWael/gbnf-lsp/releases/download/${version}/${binName}`;

    try {
      outputChannel.appendLine(`Downloading binary from ${url}...`);
      await downloadFile(url, binpath)
        .then(() => {
          if (os.platform() !== "win32") {
            fs.chmodSync(binpath, 0o755);
          }
          outputChannel.appendLine(`Downloaded ${binName} successfully.`);
        })
        .catch((err) => {
          throw new Error(`Failed to download binary: ${err.message}`);
        });
    } catch (err) {
      throw new Error(`Error while fetching binary: ${err}`);
    }
  }
  return binpath;
}

export async function activate(context: vscode.ExtensionContext) {
  // Create output channel for logging
  outputChannel = vscode.window.createOutputChannel("GBNF LSP");
  outputChannel.show();

  outputChannel.appendLine("Extension activating...");

  try {
    const serverPath = await getPlatformBinary(context);

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

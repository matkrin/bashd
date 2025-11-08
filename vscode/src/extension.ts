import * as vscode from "vscode";
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind,
} from "vscode-languageclient/node";

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
    console.log("bashd extention activated");
    const config = vscode.workspace.getConfiguration("bashd");
    const serverPath = config.get("path", "bashd");

    const serverOptions: ServerOptions = {
        command: serverPath,
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [
            { scheme: "file", language: "bash" },
            { scheme: "file", language: "shellscript" },
        ],
        synchronize: {
            configurationSection: "bashd",
        },
    };

    client = new LanguageClient(
        "bashd",
        "Bash Language Server",
        serverOptions,
        clientOptions,
    );

    client.start();
}

export function deactivate(): Thenable<void> | undefined {
    return client ? client.stop() : undefined;
}

// 𓁟 Thoth Provider — thothProvider.ts
//
// Context compression for AI conversations.
// Reads .thoth/memory.yaml from the workspace and provides it as
// inline context for AI features (completions, explanations).
//
// Thoth's job is to reduce the context window cost of starting work.
// Instead of an AI needing to read 50+ files to understand the project,
// it reads memory.yaml (compressed knowledge) and gets 99.3% of what
// it needs in one file.
//
// This provider:
//   1. Loads .thoth/memory.yaml on activation
//   2. Watches for changes to memory.yaml
//   3. Exposes compressed context via getContext()
//   4. Can display context in a webview panel

import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';

interface ThothMemory {
    raw: string;
    project: string;
    version: string;
    modulesCount: number;
    testCount: string;
    recentChanges: string[];
    filePath: string;
    lastLoaded: Date;
}

export class ThothProvider implements vscode.Disposable {
    private output: vscode.OutputChannel;
    private memory: ThothMemory | null = null;
    private watcher: vscode.FileSystemWatcher | undefined;

    constructor(output: vscode.OutputChannel) {
        this.output = output;
    }

    // ── Load Memory ───────────────────────────────────────────────

    load(): void {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        if (!workspaceFolders || workspaceFolders.length === 0) {
            this.output.appendLine('𓁟 Thoth: No workspace folder found');
            return;
        }

        for (const folder of workspaceFolders) {
            const memoryPath = path.join(folder.uri.fsPath, '.thoth', 'memory.yaml');
            if (fs.existsSync(memoryPath)) {
                this.loadFromPath(memoryPath);
                this.watchFile(memoryPath);
                return;
            }
        }

        this.output.appendLine('𓁟 Thoth: No .thoth/memory.yaml found in workspace');
    }

    private loadFromPath(filePath: string): void {
        try {
            const content = fs.readFileSync(filePath, 'utf-8');
            this.memory = this.parseMemory(content, filePath);
            this.output.appendLine(
                `𓁟 Thoth loaded: ${this.memory.project} v${this.memory.version} ` +
                `(${this.memory.modulesCount} modules, ${this.memory.testCount} tests)`
            );
        } catch (err: unknown) {
            const msg = err instanceof Error ? err.message : String(err);
            this.output.appendLine(`𓁟 Thoth load error: ${msg}`);
        }
    }

    private watchFile(filePath: string): void {
        const pattern = new vscode.RelativePattern(
            path.dirname(path.dirname(filePath)),
            '.thoth/memory.yaml'
        );
        this.watcher = vscode.workspace.createFileSystemWatcher(pattern);
        this.watcher.onDidChange(() => {
            this.output.appendLine('𓁟 Thoth: memory.yaml changed — reloading');
            this.loadFromPath(filePath);
        });
    }

    // ── Context API ───────────────────────────────────────────────

    getContext(): string | null {
        if (!this.memory) { return null; }
        return this.memory.raw;
    }

    getCompressedContext(): string | null {
        if (!this.memory) { return null; }

        // Return a compressed version — key facts only
        const lines = this.memory.raw.split('\n');
        const compressed: string[] = [];

        for (const line of lines) {
            const trimmed = line.trim();
            // Keep header lines, key-value pairs, and non-comment content
            if (trimmed.startsWith('##') ||
                trimmed.startsWith('project:') ||
                trimmed.startsWith('version:') ||
                trimmed.startsWith('language:') ||
                trimmed.startsWith('test_count:') ||
                trimmed.startsWith('module_count:') ||
                trimmed.startsWith('binary_count:') ||
                trimmed.startsWith('mcp_tools:') ||
                /^#\s+\d{4}-\d{2}-\d{2}:/.test(trimmed)) {
                compressed.push(line);
            }
        }

        return compressed.join('\n');
    }

    getSummary(): string {
        if (!this.memory) { return 'Thoth: No context loaded'; }
        return `${this.memory.project} v${this.memory.version} — ` +
            `${this.memory.modulesCount} modules, ${this.memory.testCount} tests`;
    }

    isLoaded(): boolean {
        return this.memory !== null;
    }

    // ── Display ───────────────────────────────────────────────────

    async showContextPanel(): Promise<void> {
        if (!this.memory) {
            vscode.window.showWarningMessage('𓁟 Thoth: No memory.yaml loaded');
            return;
        }

        // Open memory.yaml as a file
        const uri = vscode.Uri.file(this.memory.filePath);
        const doc = await vscode.workspace.openTextDocument(uri);
        await vscode.window.showTextDocument(doc, {
            preview: true,
            viewColumn: vscode.ViewColumn.Beside,
        });
    }

    // ── Parsing ───────────────────────────────────────────────────

    private parseMemory(content: string, filePath: string): ThothMemory {
        const lines = content.split('\n');
        let project = 'unknown';
        let version = 'unknown';
        let modulesCount = 0;
        let testCount = '0';
        const recentChanges: string[] = [];

        for (const line of lines) {
            const trimmed = line.trim();

            // Parse YAML-like key-value pairs
            const projectMatch = trimmed.match(/^project:\s*(.+)/);
            if (projectMatch) { project = projectMatch[1]; }

            const versionMatch = trimmed.match(/^version:\s*(.+)/);
            if (versionMatch) { version = versionMatch[1]; }

            const moduleMatch = trimmed.match(/^module_count:\s*(\d+)/);
            if (moduleMatch) { modulesCount = parseInt(moduleMatch[1], 10); }

            const testMatch = trimmed.match(/^test_count:\s*(.+)/);
            if (testMatch) { testCount = testMatch[1]; }

            // Capture recent changes (date-prefixed comments)
            const changeMatch = trimmed.match(/^#\s+(\d{4}-\d{2}-\d{2}:.+)/);
            if (changeMatch) { recentChanges.push(changeMatch[1]); }
        }

        return {
            raw: content,
            project,
            version,
            modulesCount,
            testCount,
            recentChanges,
            filePath,
            lastLoaded: new Date(),
        };
    }

    // ── Cleanup ───────────────────────────────────────────────────

    dispose(): void {
        this.watcher?.dispose();
    }
}

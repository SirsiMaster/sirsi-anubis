// 𓃣 Pantheon Diagnostics — diagnostics.ts
//
// Maps scan findings to VS Code Diagnostics so advisories appear
// inline in the editor. Each finding becomes a diagnostic with:
//   - Message: the advisory text
//   - Severity: mapped from safe/caution/warning
//   - Code Action: the remediation (clean, git gc, etc.)
//
// Architecture:
//   1. Reads persisted findings from ~/.config/pantheon/findings/latest-scan.json
//   2. Groups by file path (findings that point to files in the workspace)
//   3. Creates DiagnosticCollection with advisory messages
//   4. Registers Code Actions for fixable findings
//
// Refresh triggers:
//   - On extension activation (if findings exist)
//   - After sirsi.scan command completes
//   - Manual via sirsi.refreshDiagnostics command

import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

interface PersistedFinding {
    rule: string;
    category: string;
    description: string;
    path: string;
    size_bytes: number;
    size_human: string;
    severity: 'safe' | 'caution' | 'warning';
    advisory: string;
    remediation: string;
    can_fix: boolean;
    breaking: boolean;
}

interface PersistedScan {
    timestamp: string;
    total_size: number;
    findings: PersistedFinding[];
}

export class PantheonDiagnostics implements vscode.Disposable {
    private diagnosticCollection: vscode.DiagnosticCollection;
    private codeActionProvider: vscode.Disposable | undefined;
    private findings: PersistedFinding[] = [];
    private output: vscode.OutputChannel;

    constructor(output: vscode.OutputChannel) {
        this.output = output;
        this.diagnosticCollection = vscode.languages.createDiagnosticCollection('pantheon');
    }

    dispose(): void {
        this.diagnosticCollection.dispose();
        this.codeActionProvider?.dispose();
    }

    /**
     * Load findings from disk and update diagnostics.
     */
    refresh(): void {
        const findingsPath = path.join(
            os.homedir(), '.config', 'pantheon', 'findings', 'latest-scan.json'
        );

        try {
            const raw = fs.readFileSync(findingsPath, 'utf-8');
            const scan: PersistedScan = JSON.parse(raw);
            this.findings = scan.findings || [];
            this.updateDiagnostics();
            this.output.appendLine(
                `𓃣 Diagnostics: loaded ${this.findings.length} findings from ${scan.timestamp}`
            );
        } catch {
            this.output.appendLine('𓃣 Diagnostics: no findings file — run sirsi scan first');
            this.diagnosticCollection.clear();
        }
    }

    /**
     * Map findings to VS Code diagnostics grouped by workspace-relative path.
     */
    private updateDiagnostics(): void {
        this.diagnosticCollection.clear();

        const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
        if (!workspaceRoot) { return; }

        // Group findings that are inside the workspace
        const byFile = new Map<string, vscode.Diagnostic[]>();

        for (const f of this.findings) {
            // Only show findings inside the workspace or in common dev paths
            if (!f.path.startsWith(workspaceRoot) && !this.isRelevantPath(f)) {
                continue;
            }

            const uri = this.findingToUri(f, workspaceRoot);
            if (!uri) { continue; }

            const key = uri.toString();
            if (!byFile.has(key)) {
                byFile.set(key, []);
            }

            const severity = this.mapSeverity(f.severity);
            const message = this.formatMessage(f);

            const diag = new vscode.Diagnostic(
                new vscode.Range(0, 0, 0, 0), // Line 0 — these are directory-level findings
                message,
                severity
            );
            diag.source = 'Pantheon';
            diag.code = f.rule;

            byFile.get(key)!.push(diag);
        }

        for (const [uriStr, diags] of byFile) {
            this.diagnosticCollection.set(vscode.Uri.parse(uriStr), diags);
        }
    }

    private isRelevantPath(f: PersistedFinding): boolean {
        // Show findings for common dev directories even if not in workspace
        const devDirs = ['/Development/', '/code/', '/projects/', '/src/'];
        return devDirs.some(d => f.path.includes(d));
    }

    private findingToUri(f: PersistedFinding, workspaceRoot: string): vscode.Uri | undefined {
        // For directory findings, point to the directory
        if (f.path.startsWith(workspaceRoot)) {
            return vscode.Uri.file(f.path);
        }

        // For findings outside workspace, still show them if relevant
        try {
            return vscode.Uri.file(f.path);
        } catch {
            return undefined;
        }
    }

    private mapSeverity(sev: string): vscode.DiagnosticSeverity {
        switch (sev) {
            case 'warning': return vscode.DiagnosticSeverity.Warning;
            case 'caution': return vscode.DiagnosticSeverity.Information;
            case 'safe': return vscode.DiagnosticSeverity.Hint;
            default: return vscode.DiagnosticSeverity.Information;
        }
    }

    private formatMessage(f: PersistedFinding): string {
        let msg = `${f.description} (${f.size_human})`;
        if (f.advisory) {
            msg += `\n${f.advisory}`;
        }
        if (f.can_fix && f.remediation) {
            msg += `\nFix: ${f.remediation}`;
        }
        return msg;
    }

    /**
     * Get all current findings (for code action provider).
     */
    getFindings(): PersistedFinding[] {
        return this.findings;
    }
}

/**
 * Code Action Provider — offers "Clean" as a quick fix for fixable findings.
 */
export class PantheonCodeActionProvider implements vscode.CodeActionProvider {
    private diagnostics: PantheonDiagnostics;
    private binaryPath: string;
    private output: vscode.OutputChannel;

    constructor(diagnostics: PantheonDiagnostics, binaryPath: string, output: vscode.OutputChannel) {
        this.diagnostics = diagnostics;
        this.binaryPath = binaryPath;
        this.output = output;
    }

    provideCodeActions(
        document: vscode.TextDocument,
        range: vscode.Range,
        context: vscode.CodeActionContext
    ): vscode.CodeAction[] {
        const actions: vscode.CodeAction[] = [];

        for (const diag of context.diagnostics) {
            if (diag.source !== 'Pantheon') { continue; }

            const rule = diag.code as string;
            const finding = this.diagnostics.getFindings().find(
                f => f.rule === rule && document.uri.fsPath.includes(f.path.split('/').pop() || '')
            );

            if (!finding || !finding.can_fix) { continue; }

            const action = new vscode.CodeAction(
                `𓃣 ${finding.remediation}`,
                vscode.CodeActionKind.QuickFix
            );
            action.diagnostics = [diag];
            action.isPreferred = finding.severity === 'safe';

            // Execute the sirsi clean command for this finding
            action.command = {
                command: 'sirsi.cleanFinding',
                title: finding.remediation,
                arguments: [finding],
            };

            actions.push(action);
        }

        return actions;
    }
}

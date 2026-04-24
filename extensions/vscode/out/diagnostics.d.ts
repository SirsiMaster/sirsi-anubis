import * as vscode from 'vscode';
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
export declare class PantheonDiagnostics implements vscode.Disposable {
    private diagnosticCollection;
    private codeActionProvider;
    private findings;
    private output;
    constructor(output: vscode.OutputChannel);
    dispose(): void;
    /**
     * Load findings from disk and update diagnostics.
     */
    refresh(): void;
    /**
     * Map findings to VS Code diagnostics grouped by workspace-relative path.
     */
    private updateDiagnostics;
    private isRelevantPath;
    private findingToUri;
    private mapSeverity;
    private formatMessage;
    /**
     * Get all current findings (for code action provider).
     */
    getFindings(): PersistedFinding[];
}
/**
 * Code Action Provider — offers "Clean" as a quick fix for fixable findings.
 */
export declare class PantheonCodeActionProvider implements vscode.CodeActionProvider {
    private diagnostics;
    private binaryPath;
    private output;
    constructor(diagnostics: PantheonDiagnostics, binaryPath: string, output: vscode.OutputChannel);
    provideCodeActions(document: vscode.TextDocument, range: vscode.Range, context: vscode.CodeActionContext): vscode.CodeAction[];
}
export {};
//# sourceMappingURL=diagnostics.d.ts.map
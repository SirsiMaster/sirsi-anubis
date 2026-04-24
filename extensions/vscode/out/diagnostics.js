"use strict";
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
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.PantheonCodeActionProvider = exports.PantheonDiagnostics = void 0;
const vscode = __importStar(require("vscode"));
const fs = __importStar(require("fs"));
const path = __importStar(require("path"));
const os = __importStar(require("os"));
class PantheonDiagnostics {
    constructor(output) {
        this.findings = [];
        this.output = output;
        this.diagnosticCollection = vscode.languages.createDiagnosticCollection('pantheon');
    }
    dispose() {
        this.diagnosticCollection.dispose();
        this.codeActionProvider?.dispose();
    }
    /**
     * Load findings from disk and update diagnostics.
     */
    refresh() {
        const findingsPath = path.join(os.homedir(), '.config', 'pantheon', 'findings', 'latest-scan.json');
        try {
            const raw = fs.readFileSync(findingsPath, 'utf-8');
            const scan = JSON.parse(raw);
            this.findings = scan.findings || [];
            this.updateDiagnostics();
            this.output.appendLine(`𓃣 Diagnostics: loaded ${this.findings.length} findings from ${scan.timestamp}`);
        }
        catch {
            this.output.appendLine('𓃣 Diagnostics: no findings file — run sirsi scan first');
            this.diagnosticCollection.clear();
        }
    }
    /**
     * Map findings to VS Code diagnostics grouped by workspace-relative path.
     */
    updateDiagnostics() {
        this.diagnosticCollection.clear();
        const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
        if (!workspaceRoot) {
            return;
        }
        // Group findings that are inside the workspace
        const byFile = new Map();
        for (const f of this.findings) {
            // Only show findings inside the workspace or in common dev paths
            if (!f.path.startsWith(workspaceRoot) && !this.isRelevantPath(f)) {
                continue;
            }
            const uri = this.findingToUri(f, workspaceRoot);
            if (!uri) {
                continue;
            }
            const key = uri.toString();
            if (!byFile.has(key)) {
                byFile.set(key, []);
            }
            const severity = this.mapSeverity(f.severity);
            const message = this.formatMessage(f);
            const diag = new vscode.Diagnostic(new vscode.Range(0, 0, 0, 0), // Line 0 — these are directory-level findings
            message, severity);
            diag.source = 'Pantheon';
            diag.code = f.rule;
            byFile.get(key).push(diag);
        }
        for (const [uriStr, diags] of byFile) {
            this.diagnosticCollection.set(vscode.Uri.parse(uriStr), diags);
        }
    }
    isRelevantPath(f) {
        // Show findings for common dev directories even if not in workspace
        const devDirs = ['/Development/', '/code/', '/projects/', '/src/'];
        return devDirs.some(d => f.path.includes(d));
    }
    findingToUri(f, workspaceRoot) {
        // For directory findings, point to the directory
        if (f.path.startsWith(workspaceRoot)) {
            return vscode.Uri.file(f.path);
        }
        // For findings outside workspace, still show them if relevant
        try {
            return vscode.Uri.file(f.path);
        }
        catch {
            return undefined;
        }
    }
    mapSeverity(sev) {
        switch (sev) {
            case 'warning': return vscode.DiagnosticSeverity.Warning;
            case 'caution': return vscode.DiagnosticSeverity.Information;
            case 'safe': return vscode.DiagnosticSeverity.Hint;
            default: return vscode.DiagnosticSeverity.Information;
        }
    }
    formatMessage(f) {
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
    getFindings() {
        return this.findings;
    }
}
exports.PantheonDiagnostics = PantheonDiagnostics;
/**
 * Code Action Provider — offers "Clean" as a quick fix for fixable findings.
 */
class PantheonCodeActionProvider {
    constructor(diagnostics, binaryPath, output) {
        this.diagnostics = diagnostics;
        this.binaryPath = binaryPath;
        this.output = output;
    }
    provideCodeActions(document, range, context) {
        const actions = [];
        for (const diag of context.diagnostics) {
            if (diag.source !== 'Pantheon') {
                continue;
            }
            const rule = diag.code;
            const finding = this.diagnostics.getFindings().find(f => f.rule === rule && document.uri.fsPath.includes(f.path.split('/').pop() || ''));
            if (!finding || !finding.can_fix) {
                continue;
            }
            const action = new vscode.CodeAction(`𓃣 ${finding.remediation}`, vscode.CodeActionKind.QuickFix);
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
exports.PantheonCodeActionProvider = PantheonCodeActionProvider;
//# sourceMappingURL=diagnostics.js.map
"use strict";
// 𓃣 Pantheon VS Code Extension — extension.ts
//
// Entry point. Activates on workspace open (onStartupFinished).
// Starts Guardian (always-on renice), status bar ankh, Thoth provider,
// Thoth Accountability Engine, and registers all command palette entries.
//
// Architecture:
//   activate() → starts Guardian background loop
//             → creates status bar ankh with live metrics
//             → registers Command Palette commands
//             → loads Thoth context from .thoth/memory.yaml
//             → runs Thoth Accountability Engine (cold-start benchmark)
//
// The Anubis Suite operates without oversight.
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
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = __importStar(require("vscode"));
const guardian_1 = require("./guardian");
const statusBar_1 = require("./statusBar");
const commands_1 = require("./commands");
const thothProvider_1 = require("./thothProvider");
const thothAccountability_1 = require("./thothAccountability");
const crashpadMonitor_1 = require("./crashpadMonitor");
const diagnostics_1 = require("./diagnostics");
let guardian;
let statusBar;
let thothProvider;
let accountabilityEngine;
let crashpadMonitor;
let diagnostics;
function activate(context) {
    const outputChannel = vscode.window.createOutputChannel('Pantheon');
    outputChannel.appendLine('𓃣 Pantheon extension activating...');
    // ── Resolve binary path ───────────────────────────────────────────
    const config = vscode.workspace.getConfiguration('sirsi');
    const binaryPath = config.get('binaryPath', 'sirsi');
    // ── Status Bar (Ankh) ─────────────────────────────────────────────
    statusBar = new statusBar_1.PantheonStatusBar(binaryPath, outputChannel);
    context.subscriptions.push(statusBar);
    // ── Guardian (Always-On) ──────────────────────────────────────────
    const guardianEnabled = config.get('guardian.enabled', true);
    if (guardianEnabled) {
        const reniceDelay = config.get('guardian.reniceDelay', 30);
        const pollInterval = config.get('guardian.pollInterval', 5);
        const autoRenice = config.get('guardian.autoRenice', true);
        guardian = new guardian_1.Guardian(binaryPath, outputChannel, {
            reniceDelaySec: reniceDelay,
            pollIntervalSec: pollInterval,
            autoRenice,
        });
        guardian.start();
        context.subscriptions.push(guardian);
        outputChannel.appendLine(`𓁵 Guardian armed — renice in ${reniceDelay}s, poll every ${pollInterval}s`);
    }
    else {
        outputChannel.appendLine('𓁵 Guardian disabled by configuration');
    }
    // ── Thoth Context Provider ────────────────────────────────────────
    const thothEnabled = config.get('thoth.enabled', true);
    if (thothEnabled) {
        thothProvider = new thothProvider_1.ThothProvider(outputChannel);
        thothProvider.load();
        context.subscriptions.push(thothProvider);
        outputChannel.appendLine('𓁟 Thoth context provider loaded');
    }
    // ── Thoth Accountability Engine ───────────────────────────────────
    const accountabilityEnabled = config.get('thoth.accountability', true);
    if (accountabilityEnabled) {
        accountabilityEngine = new thothAccountability_1.ThothAccountabilityEngine(context, outputChannel);
        context.subscriptions.push(accountabilityEngine);
        // Run benchmark async — don't block activation
        accountabilityEngine.activate().catch(err => {
            const msg = err instanceof Error ? err.message : String(err);
            outputChannel.appendLine(`𓁟 Accountability Engine error: ${msg}`);
        });
        outputChannel.appendLine('𓁟 Thoth Accountability Engine armed');
    }
    // ── Crashpad Monitor ──────────────────────────────────────────────
    crashpadMonitor = new crashpadMonitor_1.CrashpadMonitor(outputChannel);
    crashpadMonitor.start();
    context.subscriptions.push(crashpadMonitor);
    outputChannel.appendLine('𓁵 Crashpad Monitor armed — tracking IDE stability');
    // ── Diagnostics Provider (Advisory Intelligence) ───────────────────
    diagnostics = new diagnostics_1.PantheonDiagnostics(outputChannel);
    diagnostics.refresh(); // Load existing findings on startup
    context.subscriptions.push(diagnostics);
    // Code Actions — "Clean" quick fixes for fixable findings
    const codeActionProvider = vscode.languages.registerCodeActionsProvider({ scheme: 'file' }, new diagnostics_1.PantheonCodeActionProvider(diagnostics, binaryPath, outputChannel), { providedCodeActionKinds: [vscode.CodeActionKind.QuickFix] });
    context.subscriptions.push(codeActionProvider);
    // Refresh diagnostics after scan completes
    context.subscriptions.push(vscode.commands.registerCommand('sirsi.refreshDiagnostics', () => {
        diagnostics?.refresh();
    }));
    // Clean a specific finding via code action
    context.subscriptions.push(vscode.commands.registerCommand('sirsi.cleanFinding', async (finding) => {
        const { execFile } = require('child_process');
        const { promisify } = require('util');
        const execFileAsync = promisify(execFile);
        try {
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: `𓃣 ${finding.remediation}...`,
            }, async () => {
                await execFileAsync(binaryPath, ['clean', 'safe'], { timeout: 30000 });
                diagnostics?.refresh();
            });
            vscode.window.showInformationMessage(`𓃣 ${finding.description}: ${finding.remediation} ✓`);
        }
        catch (err) {
            vscode.window.showErrorMessage(`𓃣 Clean failed: ${err.message}`);
        }
    }));
    outputChannel.appendLine(`𓃣 Diagnostics armed — ${diagnostics.getFindings().length} findings loaded`);
    // ── Command Palette Registration ──────────────────────────────────
    (0, commands_1.registerCommands)(context, binaryPath, outputChannel, statusBar, thothProvider, guardian, accountabilityEngine, crashpadMonitor);
    // ── Workspace Optimization ────────────────────────────────────────
    const autoOptimize = config.get('workspace.autoOptimize', false);
    if (autoOptimize) {
        applyOptimalSettings(outputChannel);
    }
    // ── Start metric refresh loop ─────────────────────────────────────
    const pollInterval = config.get('guardian.pollInterval', 5);
    statusBar.startMetricLoop(pollInterval * 1000);
    outputChannel.appendLine('𓃣 Pantheon extension activated — the Anubis Suite is operational');
    // Show welcome notification on first install
    const hasShownWelcome = context.globalState.get('pantheon.welcomeShown');
    if (!hasShownWelcome) {
        vscode.window.showInformationMessage('𓃣 Pantheon activated. Guardian is monitoring your workspace.', 'Show Metrics', 'Dismiss').then(choice => {
            if (choice === 'Show Metrics') {
                vscode.commands.executeCommand('sirsi.showMetrics');
            }
        });
        context.globalState.update('pantheon.welcomeShown', true);
    }
}
function deactivate() {
    guardian?.dispose();
    statusBar?.dispose();
    thothProvider?.dispose();
    accountabilityEngine?.dispose();
    crashpadMonitor?.dispose();
    diagnostics?.dispose();
}
// ── Workspace Settings ────────────────────────────────────────────────
function applyOptimalSettings(outputChannel) {
    const wsConfig = vscode.workspace.getConfiguration();
    // gopls directory filters — exclude non-Go directories from analysis
    const goplsFilters = wsConfig.get('gopls.directoryFilters');
    if (!goplsFilters || goplsFilters.length === 0) {
        wsConfig.update('gopls.directoryFilters', [
            '-**/node_modules',
            '-**/.git',
            '-**/vendor',
            '-**/.vscode-test',
            '-**/dist',
        ], vscode.ConfigurationTarget.Workspace);
        outputChannel.appendLine('𓃣 Applied gopls directory filters');
    }
    // File watcher exclusions — reduce inotify/kqueue pressure
    const existingExcludes = wsConfig.get('files.watcherExclude') || {};
    const extraExcludes = {
        '**/node_modules/**': true,
        '**/.git/objects/**': true,
        '**/.git/subtree-cache/**': true,
        '**/dist/**': true,
        '**/coverage/**': true,
    };
    let needsUpdate = false;
    for (const [pattern, value] of Object.entries(extraExcludes)) {
        if (!(pattern in existingExcludes)) {
            existingExcludes[pattern] = value;
            needsUpdate = true;
        }
    }
    if (needsUpdate) {
        wsConfig.update('files.watcherExclude', existingExcludes, vscode.ConfigurationTarget.Workspace);
        outputChannel.appendLine('𓃣 Applied file watcher exclusions');
    }
    // Disable shell integration if causing issues
    const shellIntegration = wsConfig.get('terminal.integrated.shellIntegration.enabled');
    if (shellIntegration !== false) {
        wsConfig.update('terminal.integrated.shellIntegration.enabled', false, vscode.ConfigurationTarget.Workspace);
        outputChannel.appendLine('𓃣 Disabled shell integration (reduces Extension Host CPU)');
    }
}
//# sourceMappingURL=extension.js.map
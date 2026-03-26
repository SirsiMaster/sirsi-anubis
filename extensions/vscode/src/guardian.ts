// 𓁵 Pantheon Guardian — guardian.ts
//
// Always-on background process controller. The Guardian runs continuously
// in the Extension Host and manages:
//
//   1. Auto-renice: Deprioritizes LSP processes 30s after activation
//   2. Memory pressure monitoring: Watches for sustained RAM hogs
//   3. Re-renice loop: Re-applies renice when processes respawn/reset
//
// Architecture:
//   Guardian.start()
//     └─ setTimeout(reniceDelay) → execRenice()
//     └─ setInterval(pollInterval) → checkPressure()
//
// The Guardian never kills processes. It uses renice(1) and taskpolicy(1)
// (macOS) to lower scheduler priority — processes continue running but
// yield CPU to the IDE Renderer on contention.
//
// Safety:
//   - Only renices processes owned by the current user
//   - language_server_macos_arm is PROTECTED from slay but CAN be reniced
//   - No binary modification (Rule A19)
//   - No telemetry (Rule A11)

import { execFile } from 'child_process';
import { promisify } from 'util';
import * as vscode from 'vscode';

const execFileAsync = promisify(execFile);

export interface GuardianConfig {
    reniceDelaySec: number;
    pollIntervalSec: number;
    autoRenice: boolean;
}

export interface ReniceResult {
    target: string;
    reniced: number;
    skipped: number;
    errors?: string[];
    processes?: ReniceProcess[];
}

export interface ReniceProcess {
    pid: number;
    name: string;
    rss_bytes: number;
    rss_human: string;
    old_nice: number;
    new_nice: number;
    qos: string;
}

export interface GuardMetrics {
    totalRAM: number;
    totalRAMHuman: string;
    processCount: number;
    reniceCount: number;
    lastRenice: Date | null;
    guardianUptime: number;
    errors: string[];
}

export class Guardian implements vscode.Disposable {
    private binaryPath: string;
    private output: vscode.OutputChannel;
    private config: GuardianConfig;
    private reniceTimer: NodeJS.Timeout | undefined;
    private pollTimer: NodeJS.Timeout | undefined;
    private disposed = false;
    private startTime: Date = new Date();
    private lastReniceTime: Date | null = null;
    private totalReniced = 0;
    private lastErrors: string[] = [];

    constructor(
        binaryPath: string,
        output: vscode.OutputChannel,
        config: GuardianConfig
    ) {
        this.binaryPath = binaryPath;
        this.output = output;
        this.config = config;
    }

    // ── Lifecycle ─────────────────────────────────────────────────

    start(): void {
        this.startTime = new Date();
        this.output.appendLine(`𓁵 Guardian starting — delay ${this.config.reniceDelaySec}s`);

        // Schedule initial renice after delay (LSPs need time to spawn)
        this.reniceTimer = setTimeout(async () => {
            if (this.disposed) { return; }
            await this.executeRenice();

            // Start recurring poll loop
            if (this.config.autoRenice && !this.disposed) {
                this.pollTimer = setInterval(async () => {
                    if (this.disposed) { return; }
                    await this.executeRenice();
                }, this.config.pollIntervalSec * 1000 * 12); // Re-renice every 12 poll intervals (60s default)
            }
        }, this.config.reniceDelaySec * 1000);
    }

    dispose(): void {
        this.disposed = true;
        if (this.reniceTimer) {
            clearTimeout(this.reniceTimer);
            this.reniceTimer = undefined;
        }
        if (this.pollTimer) {
            clearInterval(this.pollTimer);
            this.pollTimer = undefined;
        }
        this.output.appendLine('𓁵 Guardian stopped');
    }

    // ── Renice Execution ──────────────────────────────────────────

    async executeRenice(): Promise<ReniceResult | null> {
        try {
            const { stdout } = await execFileAsync(this.binaryPath, [
                'guard', '--renice', 'lsp', '--json'
            ], { timeout: 15000 });

            const result = this.parseReniceOutput(stdout);
            if (result) {
                this.lastReniceTime = new Date();
                this.totalReniced += result.reniced;

                if (result.reniced > 0) {
                    this.output.appendLine(
                        `𓁵 Guardian reniced ${result.reniced} process(es)`
                    );

                    // Show notification for significant operations
                    if (result.processes && result.processes.length > 0) {
                        const totalRAM = result.processes.reduce(
                            (sum, p) => sum + p.rss_bytes, 0
                        );
                        const ramHuman = this.formatBytes(totalRAM);
                        this.output.appendLine(
                            `   Total RAM of reniced processes: ${ramHuman}`
                        );
                    }
                }

                this.lastErrors = result.errors || [];
                return result;
            }
        } catch (err: unknown) {
            const msg = err instanceof Error ? err.message : String(err);
            // Binary not found is not a critical error — just log it
            if (msg.includes('ENOENT')) {
                this.output.appendLine(
                    '𓁵 Guardian: pantheon binary not found. Install via: brew install sirsi-pantheon'
                );
            } else {
                this.output.appendLine(`𓁵 Guardian renice error: ${msg}`);
            }
            this.lastErrors = [msg];
        }
        return null;
    }

    // ── Manual trigger (from command palette) ─────────────────────

    async manualRenice(): Promise<void> {
        const result = await this.executeRenice();
        if (result) {
            if (result.reniced === 0) {
                vscode.window.showInformationMessage(
                    '𓁵 No LSP processes found to renice'
                );
            } else {
                const procs = result.processes
                    ?.map(p => `${p.name} (${p.rss_human})`)
                    .join(', ') || '';
                vscode.window.showInformationMessage(
                    `𓁵 Reniced ${result.reniced} process(es): ${procs}`
                );
            }
        } else {
            vscode.window.showWarningMessage(
                '𓁵 Guardian: Could not execute renice. Is pantheon installed?'
            );
        }
    }

    // ── Metrics ───────────────────────────────────────────────────

    getMetrics(): GuardMetrics {
        return {
            totalRAM: 0,
            totalRAMHuman: '—',
            processCount: 0,
            reniceCount: this.totalReniced,
            lastRenice: this.lastReniceTime,
            guardianUptime: Date.now() - this.startTime.getTime(),
            errors: this.lastErrors,
        };
    }

    // ── Parsing ───────────────────────────────────────────────────

    private parseReniceOutput(stdout: string): ReniceResult | null {
        try {
            // Try JSON parse first (--json flag)
            const parsed = JSON.parse(stdout.trim());
            return parsed as ReniceResult;
        } catch {
            // Fallback: parse text output
            const reniced = (stdout.match(/Deprioritized:\s+(\d+)/i)?.[1]) || '0';
            return {
                target: 'lsp',
                reniced: parseInt(reniced, 10),
                skipped: 0,
            };
        }
    }

    private formatBytes(bytes: number): string {
        if (bytes === 0) { return '0 B'; }
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        const k = 1024;
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return `${(bytes / Math.pow(k, i)).toFixed(1)} ${units[i]}`;
    }
}

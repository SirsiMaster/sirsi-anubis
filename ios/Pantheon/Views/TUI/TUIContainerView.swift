import SwiftUI

/// Terminal emulator view — renders Pantheon output in the familiar CLI style.
/// Gold-on-black monospace text with command input, mimicking the BubbleTea TUI.
struct TUIContainerView: View {
    @EnvironmentObject var appState: AppState
    @StateObject private var terminal = TerminalState()

    var body: some View {
        VStack(spacing: 0) {
            // Terminal output
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 2) {
                        ForEach(terminal.lines) { line in
                            TerminalLine(line: line)
                                .id(line.id)
                        }
                    }
                    .padding(12)
                }
                .onChange(of: terminal.lines.count) { _, _ in
                    if let last = terminal.lines.last {
                        proxy.scrollTo(last.id, anchor: .bottom)
                    }
                }
            }
            .background(PantheonTheme.tuiBackground)

            Divider().background(PantheonTheme.gold.opacity(0.3))

            // Command input
            HStack(spacing: 8) {
                Text("𓃣")
                    .font(.system(size: 16))
                    .foregroundStyle(PantheonTheme.tuiPrompt)

                TextField("sirsi command...", text: $terminal.inputText)
                    .font(PantheonTheme.tuiFont)
                    .foregroundStyle(PantheonTheme.tuiText)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                    .onSubmit {
                        Task { await terminal.execute(bridge: appState.bridge) }
                    }

                Button {
                    Task { await terminal.execute(bridge: appState.bridge) }
                } label: {
                    Image(systemName: "return")
                        .foregroundStyle(PantheonTheme.gold)
                }
                .disabled(terminal.inputText.isEmpty || terminal.isExecuting)
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 10)
            .background(PantheonTheme.tuiBackground)
        }
    }
}

// MARK: - Terminal State

@MainActor
final class TerminalState: ObservableObject {
    @Published var lines: [TUILine] = [
        TUILine(text: "𓁢 Pantheon v0.15.0-ios", style: .gold),
        TUILine(text: "Type a command: scan, ghosts, thoth, hardware, seshat", style: .dim),
        TUILine(text: "─────────────────────────────────────────", style: .dim),
    ]
    @Published var inputText = ""
    @Published var isExecuting = false

    func execute(bridge: PantheonBridge) async {
        let command = inputText.trimmingCharacters(in: .whitespaces)
        guard !command.isEmpty else { return }

        lines.append(TUILine(text: "𓃣 \(command)", style: .prompt))
        inputText = ""
        isExecuting = true
        defer { isExecuting = false }

        await dispatch(command: command, bridge: bridge)
    }

    private func dispatch(command: String, bridge: PantheonBridge) async {
        let parts = command.lowercased().split(separator: " ")
        guard let cmd = parts.first else { return }

        switch cmd {
        case "scan", "anubis":
            await runAnubisScan(bridge: bridge)
        case "ghosts", "ka":
            await runKaHunt(bridge: bridge)
        case "hardware", "seba":
            await runSebaDetect(bridge: bridge)
        case "thoth":
            emit("Thoth requires a project path. Use GUI mode for folder selection.", style: .warning)
        case "seshat":
            await runSeshatList(bridge: bridge)
        case "version":
            emit("Pantheon \(bridge.version())", style: .normal)
        case "help":
            emit("Available commands:", style: .gold)
            emit("  scan      — Run Anubis infrastructure scan", style: .normal)
            emit("  ghosts    — Hunt Ka ghost residuals", style: .normal)
            emit("  hardware  — Seba hardware detection", style: .normal)
            emit("  thoth     — Project memory (use GUI)", style: .normal)
            emit("  seshat    — Knowledge bridge sources", style: .normal)
            emit("  version   — Show version", style: .normal)
            emit("  clear     — Clear terminal", style: .normal)
        case "clear":
            lines.removeAll()
        default:
            emit("Unknown command: \(command). Type 'help' for commands.", style: .error)
        }
    }

    private func runAnubisScan(bridge: PantheonBridge) async {
        emit("𓁢 Starting Anubis scan...", style: .gold)
        do {
            let root = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first?.path ?? "/"
            let result = try await bridge.anubisScan(rootPath: root)
            emit("Scan complete: \(result.findings.count) findings", style: .success)
            emit("Total reclaimable: \(ByteCountFormatter.string(fromByteCount: result.totalSize, countStyle: .file))", style: .gold)
            for finding in result.findings.prefix(20) {
                emit("  [\(finding.severity)] \(finding.description) — \(finding.formattedSize)", style: .normal)
            }
            if result.findings.count > 20 {
                emit("  ... and \(result.findings.count - 20) more", style: .dim)
            }
        } catch {
            emit("Error: \(error.localizedDescription)", style: .error)
        }
    }

    private func runKaHunt(bridge: PantheonBridge) async {
        emit("𓂓 Hunting ghosts...", style: .gold)
        do {
            let ghosts = try await bridge.kaHunt()
            if ghosts.isEmpty {
                emit("No ghosts found — system is clean.", style: .success)
            } else {
                emit("\(ghosts.count) ghost(s) detected:", style: .warning)
                for ghost in ghosts {
                    emit("  𓂓 \(ghost.appName) (\(ghost.bundleId)) — \(ghost.formattedSize)", style: .normal)
                    for residual in ghost.residuals.prefix(3) {
                        emit("    └─ \(residual.path)", style: .dim)
                    }
                }
            }
        } catch {
            emit("Error: \(error.localizedDescription)", style: .error)
        }
    }

    private func runSebaDetect(bridge: PantheonBridge) async {
        emit("𓇽 Detecting hardware...", style: .gold)
        do {
            let hw = try await bridge.sebaDetectHardware()
            emit("CPU: \(hw.cpuModel) (\(hw.cpuArch), \(hw.cpuCores) cores)", style: .normal)
            emit("RAM: \(hw.formattedRAM)", style: .normal)
            if let gpu = hw.gpu {
                emit("GPU: \(gpu.name) [\(gpu.type)]", style: .normal)
                if let metal = gpu.metalFamily {
                    emit("Metal: \(metal)", style: .success)
                }
            }
            if let ne = hw.neuralEngine, ne {
                emit("Neural Engine: Available", style: .success)
            }
        } catch {
            emit("Error: \(error.localizedDescription)", style: .error)
        }
    }

    private func runSeshatList(bridge: PantheonBridge) async {
        emit("𓁆 Knowledge sources:", style: .gold)
        do {
            let sources = try bridge.seshatListSources()
            for source in sources {
                emit("  • \(source.name) — \(source.description)", style: .normal)
            }
        } catch {
            emit("Error: \(error.localizedDescription)", style: .error)
        }
    }

    private func emit(_ text: String, style: TUILine.Style) {
        lines.append(TUILine(text: text, style: style))
    }
}

// MARK: - TUI Line Model

struct TUILine: Identifiable {
    let id = UUID()
    let text: String
    let style: Style

    enum Style {
        case normal, gold, prompt, dim, success, warning, error
    }
}

// MARK: - Terminal Line View

struct TerminalLine: View {
    let line: TUILine

    var body: some View {
        Text(line.text)
            .font(PantheonTheme.tuiFont)
            .foregroundStyle(color)
            .textSelection(.enabled)
    }

    private var color: Color {
        switch line.style {
        case .normal:  return PantheonTheme.tuiText
        case .gold:    return PantheonTheme.gold
        case .prompt:  return PantheonTheme.tuiPrompt
        case .dim:     return PantheonTheme.textSecondary
        case .success: return PantheonTheme.success
        case .warning: return PantheonTheme.warning
        case .error:   return PantheonTheme.error
        }
    }
}

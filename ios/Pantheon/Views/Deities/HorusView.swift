import SwiftUI

/// 𓂀 Horus — Code graph browser view.
/// Parse Go projects and explore structural symbol outlines.
struct HorusView: View {
    @EnvironmentObject var appState: AppState
    @State private var projectRoot = ""
    @State private var graph: HorusSymbolGraph?
    @State private var matchedSymbols: [HorusSymbol]?
    @State private var searchPattern = ""
    @State private var isParsing = false
    @State private var errorMessage: String?
    @State private var selectedSymbol: HorusSymbol?
    @State private var symbolContext: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                // Header
                DeityHeader(
                    glyph: "\u{13080}",
                    name: "Horus",
                    subtitle: "The All-Seeing",
                    description: "Extract structural code symbols. Serve outlines instead of full files."
                )

                // Project root input
                VStack(alignment: .leading, spacing: 8) {
                    Text("Project Root")
                        .font(.caption)
                        .foregroundStyle(PantheonTheme.textSecondary)

                    TextField("/path/to/project", text: $projectRoot)
                        .textFieldStyle(.plain)
                        .padding(10)
                        .background(PantheonTheme.surfaceElevated)
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                        .font(.system(.caption, design: .monospaced))
                }

                // Parse button
                Button {
                    Task { await parseProject() }
                } label: {
                    HStack {
                        Image(systemName: isParsing ? "progress.indicator" : "eye.fill")
                        Text(isParsing ? "Parsing..." : "Parse Project")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(PantheonTheme.gold)
                    .foregroundStyle(.black)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .font(.headline)
                }
                .disabled(isParsing || projectRoot.isEmpty)

                // Error
                if let errorMessage {
                    ErrorRetryView(message: errorMessage) { await parseProject() }
                }

                // Loading
                if isParsing {
                    ScanResultSkeleton()
                }

                // Graph stats
                if let graph = graph, !isParsing {
                    HorusStatsCard(stats: graph.stats, packageCount: graph.packages.count)

                    // Symbol search
                    HStack(spacing: 8) {
                        TextField("Search symbols (e.g., Filter*)", text: $searchPattern)
                            .textFieldStyle(.plain)
                            .padding(10)
                            .background(PantheonTheme.surfaceElevated)
                            .clipShape(RoundedRectangle(cornerRadius: 8))

                        Button {
                            Task { await searchSymbols() }
                        } label: {
                            Image(systemName: "magnifyingglass")
                                .padding(10)
                                .background(PantheonTheme.gold)
                                .foregroundStyle(.black)
                                .clipShape(RoundedRectangle(cornerRadius: 8))
                        }
                        .disabled(searchPattern.isEmpty)
                    }

                    // Symbol list (from search or full graph)
                    let displaySymbols = matchedSymbols ?? graph.symbols.filter { $0.exported }
                    let label = matchedSymbols != nil
                        ? "\(displaySymbols.count) match\(displaySymbols.count == 1 ? "" : "es")"
                        : "\(displaySymbols.count) exported symbols"

                    Text(label)
                        .font(.caption)
                        .foregroundStyle(PantheonTheme.textSecondary)

                    ForEach(displaySymbols.prefix(100)) { symbol in
                        HorusSymbolRow(symbol: symbol) {
                            selectedSymbol = symbol
                            Task { await loadContext(for: symbol) }
                        }
                    }

                    if displaySymbols.count > 100 {
                        Text("Showing first 100 of \(displaySymbols.count) symbols")
                            .font(.caption)
                            .foregroundStyle(PantheonTheme.textSecondary)
                    }
                }

                // Symbol context detail
                if let context = symbolContext, let sym = selectedSymbol {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Context: \(sym.name)")
                            .font(.headline)
                            .foregroundStyle(PantheonTheme.gold)

                        Text(context)
                            .font(.system(.caption, design: .monospaced))
                            .padding(8)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .background(PantheonTheme.surfaceElevated)
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                    }
                }
            }
            .padding()
        }
        .background(PantheonTheme.background)
    }

    private func parseProject() async {
        isParsing = true
        errorMessage = nil
        matchedSymbols = nil
        symbolContext = nil
        selectedSymbol = nil
        defer { isParsing = false }

        do {
            graph = try await appState.bridge.horusParseDir(root: projectRoot)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func searchSymbols() async {
        guard !searchPattern.isEmpty else { return }
        do {
            matchedSymbols = try await appState.bridge.horusMatchSymbols(
                root: projectRoot, pattern: searchPattern
            )
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func loadContext(for symbol: HorusSymbol) async {
        do {
            let result: HorusContextResult = try await appState.bridge.horusContextFor(
                root: projectRoot, symbolName: symbol.name
            )
            symbolContext = result.context
        } catch {
            symbolContext = "Error: \(error.localizedDescription)"
        }
    }
}

// MARK: - Subviews

struct HorusStatsCard: View {
    let stats: HorusGraphStats
    let packageCount: Int

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Code Graph")
                .font(.headline)
                .foregroundStyle(PantheonTheme.gold)

            HStack(spacing: 24) {
                StatPill(label: "Files", value: "\(stats.files)")
                StatPill(label: "Packages", value: "\(packageCount)")
                StatPill(label: "Types", value: "\(stats.types)")
            }

            HStack(spacing: 24) {
                StatPill(label: "Functions", value: "\(stats.functions)")
                StatPill(label: "Methods", value: "\(stats.methods)")
                StatPill(label: "Lines", value: "\(stats.totalLines)")
            }
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct HorusSymbolRow: View {
    let symbol: HorusSymbol
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            HStack {
                Text(symbol.kindIcon)
                    .font(.system(.caption, design: .monospaced).bold())
                    .frame(width: 32)
                    .foregroundStyle(PantheonTheme.gold)

                VStack(alignment: .leading, spacing: 2) {
                    Text(symbol.parent != nil ? "\(symbol.parent!).\(symbol.name)" : symbol.name)
                        .font(.subheadline)
                        .foregroundStyle(PantheonTheme.textPrimary)
                    Text(symbol.signature)
                        .font(.caption)
                        .foregroundStyle(PantheonTheme.textSecondary)
                        .lineLimit(1)
                }

                Spacer()

                Text("\(symbol.file):\(symbol.line)")
                    .font(.caption2)
                    .foregroundStyle(PantheonTheme.textSecondary)
            }
            .padding()
            .background(PantheonTheme.surface)
            .clipShape(RoundedRectangle(cornerRadius: 8))
        }
        .buttonStyle(.plain)
    }
}

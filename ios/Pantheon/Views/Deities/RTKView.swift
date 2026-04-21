import SwiftUI

/// ⚡ RTK — Output filter view.
/// Filter raw tool output to reduce AI context window consumption.
struct RTKView: View {
    @EnvironmentObject var appState: AppState
    @State private var rawInput = ""
    @State private var filterResult: FilterResult?
    @State private var filterConfig: FilterConfig?
    @State private var isFiltering = false
    @State private var errorMessage: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                // Header
                DeityHeader(
                    glyph: "⚡",
                    name: "RTK",
                    subtitle: "The Sieve",
                    description: "Strip noise from tool output before it consumes your AI context window."
                )

                // Config pills
                if let config = filterConfig {
                    HStack(spacing: 12) {
                        StatPill(label: "ANSI Strip", value: config.stripAnsi ? "ON" : "OFF")
                        StatPill(label: "Dedup", value: config.dedup ? "ON" : "OFF")
                        StatPill(label: "Tail", value: "\(config.tailLines)")
                    }
                }

                // Input area
                VStack(alignment: .leading, spacing: 8) {
                    Text("Raw Output")
                        .font(.caption)
                        .foregroundStyle(PantheonTheme.textSecondary)

                    TextEditor(text: $rawInput)
                        .frame(minHeight: 120)
                        .scrollContentBackground(.hidden)
                        .padding(8)
                        .background(PantheonTheme.surfaceElevated)
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                        .font(.system(.caption, design: .monospaced))
                }

                // Filter button
                Button {
                    Task { await runFilter() }
                } label: {
                    HStack {
                        Image(systemName: isFiltering ? "progress.indicator" : "bolt.fill")
                        Text(isFiltering ? "Filtering..." : "Apply Filter")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(PantheonTheme.gold)
                    .foregroundStyle(.black)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .font(.headline)
                }
                .disabled(isFiltering || rawInput.isEmpty)

                // Error
                if let errorMessage {
                    ErrorRetryView(message: errorMessage) { await runFilter() }
                }

                // Loading skeleton
                if isFiltering {
                    ScanResultSkeleton()
                }

                // Results
                if let result = filterResult, !isFiltering {
                    RTKResultCard(result: result)

                    // Filtered output
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Filtered Output")
                            .font(.caption)
                            .foregroundStyle(PantheonTheme.textSecondary)

                        Text(result.output)
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
        .task { await loadConfig() }
    }

    private func loadConfig() async {
        do {
            filterConfig = try appState.bridge.rtkDefaultConfig()
        } catch {
            // Non-fatal — config display is optional.
        }
    }

    private func runFilter() async {
        isFiltering = true
        errorMessage = nil
        defer { isFiltering = false }

        do {
            filterResult = try await appState.bridge.rtkFilter(rawOutput: rawInput)
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

// MARK: - Subviews

struct RTKResultCard: View {
    let result: FilterResult

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Filter Complete")
                .font(.headline)
                .foregroundStyle(PantheonTheme.gold)

            HStack(spacing: 24) {
                StatPill(label: "Original", value: result.formattedOriginal)
                StatPill(label: "Filtered", value: result.formattedFiltered)
                StatPill(label: "Saved", value: result.reductionPercent)
            }

            HStack(spacing: 24) {
                StatPill(label: "Lines Removed", value: "\(result.linesRemoved)")
                StatPill(label: "Dups Collapsed", value: "\(result.dupsCollapsed)")
                if result.truncated {
                    StatPill(label: "Truncated", value: "Yes")
                }
            }
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

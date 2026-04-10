import SwiftUI

/// 𓁆 Seshat — Knowledge bridge view.
/// Ingest knowledge from external sources, export to targets.
struct SeshatView: View {
    @EnvironmentObject var appState: AppState
    @State private var sources: [KnowledgeSource] = []
    @State private var selectedSources: Set<String> = []
    @State private var ingestResults: [IngestResult] = []
    @State private var isIngesting = false
    @State private var sinceDays = 7
    @State private var errorMessage: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                DeityHeader(
                    glyph: "𓁆",
                    name: "Seshat",
                    subtitle: "The Knowledge Keeper",
                    description: "Bridge knowledge between sources — Chrome, Notes, Gemini, Claude — and export to Thoth, NotebookLM."
                )

                // Source selection
                if !sources.isEmpty {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Knowledge Sources")
                            .font(.headline)
                            .foregroundStyle(PantheonTheme.gold)

                        ForEach(sources) { source in
                            SourceToggle(
                                source: source,
                                isSelected: selectedSources.contains(source.name),
                                onToggle: {
                                    if selectedSources.contains(source.name) {
                                        selectedSources.remove(source.name)
                                    } else {
                                        selectedSources.insert(source.name)
                                    }
                                }
                            )
                        }
                    }
                    .padding()
                    .background(PantheonTheme.surface)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }

                // Time range
                VStack(alignment: .leading, spacing: 8) {
                    Text("Time Range")
                        .font(.subheadline)
                        .foregroundStyle(PantheonTheme.textSecondary)

                    Picker("Since", selection: $sinceDays) {
                        Text("24 hours").tag(1)
                        Text("7 days").tag(7)
                        Text("30 days").tag(30)
                        Text("90 days").tag(90)
                    }
                    .pickerStyle(.segmented)
                }

                // Ingest button
                Button {
                    Task { await runIngest() }
                } label: {
                    HStack {
                        Image(systemName: isIngesting ? "progress.indicator" : "arrow.down.doc")
                        Text(isIngesting ? "Ingesting..." : "Ingest Knowledge")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(selectedSources.isEmpty ? PantheonTheme.surfaceElevated : PantheonTheme.gold)
                    .foregroundStyle(selectedSources.isEmpty ? PantheonTheme.textSecondary : .black)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .font(.headline)
                }
                .disabled(isIngesting || selectedSources.isEmpty)

                if let errorMessage {
                    ErrorBanner(message: errorMessage)
                }

                // Results
                if !ingestResults.isEmpty {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Ingestion Results")
                            .font(.headline)
                            .foregroundStyle(PantheonTheme.gold)

                        ForEach(ingestResults) { result in
                            HStack {
                                Image(systemName: result.error == nil ? "checkmark.circle.fill" : "xmark.circle.fill")
                                    .foregroundStyle(result.error == nil ? PantheonTheme.success : PantheonTheme.error)
                                Text(result.source)
                                    .font(.subheadline)
                                Spacer()
                                if let error = result.error {
                                    Text(error)
                                        .font(.caption)
                                        .foregroundStyle(PantheonTheme.error)
                                } else {
                                    Text("\(result.count) items")
                                        .font(.subheadline.bold())
                                        .foregroundStyle(PantheonTheme.gold)
                                }
                            }
                            .padding(.vertical, 4)
                        }
                    }
                    .padding()
                    .background(PantheonTheme.surface)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
            }
            .padding()
        }
        .background(PantheonTheme.background)
        .task { loadSources() }
    }

    private func loadSources() {
        do {
            sources = try appState.bridge.seshatListSources()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func runIngest() async {
        isIngesting = true
        errorMessage = nil
        defer { isIngesting = false }

        do {
            ingestResults = try await appState.bridge.seshatIngest(
                sources: Array(selectedSources),
                sinceDays: sinceDays
            )
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct SourceToggle: View {
    let source: KnowledgeSource
    let isSelected: Bool
    let onToggle: () -> Void

    var body: some View {
        Button(action: onToggle) {
            HStack {
                Image(systemName: isSelected ? "checkmark.square.fill" : "square")
                    .foregroundStyle(isSelected ? PantheonTheme.gold : PantheonTheme.textSecondary)
                VStack(alignment: .leading) {
                    Text(source.name)
                        .font(.subheadline)
                        .foregroundStyle(PantheonTheme.textPrimary)
                    Text(source.description)
                        .font(.caption)
                        .foregroundStyle(PantheonTheme.textSecondary)
                }
                Spacer()
            }
        }
    }
}

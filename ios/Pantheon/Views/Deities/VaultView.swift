import SwiftUI

/// 🏛️ Vault — Context sandbox view.
/// Store and search large tool output outside the AI context window.
struct VaultView: View {
    @EnvironmentObject var appState: AppState
    @State private var searchQuery = ""
    @State private var searchResults: VaultSearchResult?
    @State private var storeStats: VaultStoreStats?
    @State private var isSearching = false
    @State private var isLoadingStats = false
    @State private var errorMessage: String?

    // Store form fields
    @State private var storeSource = ""
    @State private var storeTag = ""
    @State private var storeContent = ""
    @State private var showStoreSheet = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                // Header
                DeityHeader(
                    glyph: "🏛️",
                    name: "Vault",
                    subtitle: "The Keeper",
                    description: "Sandbox large output in FTS5 storage, searchable by BM25 ranking."
                )

                // Stats card
                if let stats = storeStats {
                    VaultStatsCard(stats: stats)
                }

                // Search bar
                HStack(spacing: 8) {
                    TextField("Search vault...", text: $searchQuery)
                        .textFieldStyle(.plain)
                        .padding(10)
                        .background(PantheonTheme.surfaceElevated)
                        .clipShape(RoundedRectangle(cornerRadius: 8))

                    Button {
                        Task { await runSearch() }
                    } label: {
                        Image(systemName: "magnifyingglass")
                            .padding(10)
                            .background(PantheonTheme.gold)
                            .foregroundStyle(.black)
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                    }
                    .disabled(isSearching || searchQuery.isEmpty)
                }

                // Store button
                Button {
                    showStoreSheet = true
                } label: {
                    HStack {
                        Image(systemName: "plus.circle.fill")
                        Text("Store Entry")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(PantheonTheme.surface)
                    .foregroundStyle(PantheonTheme.gold)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .font(.headline)
                }

                // Error
                if let errorMessage {
                    ErrorRetryView(message: errorMessage) { await runSearch() }
                }

                // Loading
                if isSearching {
                    ScanResultSkeleton()
                }

                // Search results
                if let results = searchResults, !isSearching {
                    Text("\(results.totalHits) result\(results.totalHits == 1 ? "" : "s") for \"\(results.query)\"")
                        .font(.caption)
                        .foregroundStyle(PantheonTheme.textSecondary)

                    ForEach(results.entries) { entry in
                        VaultEntryRow(entry: entry)
                    }
                }
            }
            .padding()
        }
        .background(PantheonTheme.background)
        .task { await loadStats() }
        .sheet(isPresented: $showStoreSheet) {
            VaultStoreSheet(
                source: $storeSource,
                tag: $storeTag,
                content: $storeContent,
                onStore: { await storeEntry() }
            )
        }
    }

    private func loadStats() async {
        isLoadingStats = true
        defer { isLoadingStats = false }

        do {
            storeStats = try await appState.bridge.vaultStats()
        } catch {
            // Non-fatal on initial load.
        }
    }

    private func runSearch() async {
        isSearching = true
        errorMessage = nil
        defer { isSearching = false }

        do {
            searchResults = try await appState.bridge.vaultSearch(query: searchQuery, limit: 20)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func storeEntry() async {
        do {
            let _: VaultEntry = try await appState.bridge.vaultStore(
                source: storeSource, tag: storeTag, content: storeContent, tokens: 0
            )
            showStoreSheet = false
            storeSource = ""
            storeTag = ""
            storeContent = ""
            await loadStats()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

// MARK: - Subviews

struct VaultStatsCard: View {
    let stats: VaultStoreStats

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Vault Statistics")
                .font(.headline)
                .foregroundStyle(PantheonTheme.gold)

            HStack(spacing: 24) {
                StatPill(label: "Entries", value: "\(stats.totalEntries)")
                StatPill(label: "Size", value: stats.formattedBytes)
                StatPill(label: "Tokens", value: "\(stats.totalTokens)")
            }
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct VaultEntryRow: View {
    let entry: VaultEntry

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                HStack(spacing: 6) {
                    Text(entry.source)
                        .font(.subheadline.bold())
                        .foregroundStyle(PantheonTheme.textPrimary)
                    Text(entry.tag)
                        .font(.caption2)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(PantheonTheme.gold.opacity(0.2))
                        .foregroundStyle(PantheonTheme.gold)
                        .clipShape(Capsule())
                }
                Text(entry.snippet ?? entry.content ?? "")
                    .font(.caption)
                    .foregroundStyle(PantheonTheme.textSecondary)
                    .lineLimit(2)
            }

            Spacer()

            VStack(alignment: .trailing, spacing: 4) {
                Text("\(entry.tokens) tok")
                    .font(.subheadline.bold())
                    .foregroundStyle(PantheonTheme.gold)
                if let created = entry.createdAt {
                    Text(created)
                        .font(.caption2)
                        .foregroundStyle(PantheonTheme.textSecondary)
                }
            }
        }
        .padding()
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}

struct VaultStoreSheet: View {
    @Binding var source: String
    @Binding var tag: String
    @Binding var content: String
    let onStore: () async -> Void

    @State private var isStoring = false

    var body: some View {
        NavigationStack {
            Form {
                Section("Metadata") {
                    TextField("Source (e.g., go build)", text: $source)
                    TextField("Tag (e.g., build_output)", text: $tag)
                }
                Section("Content") {
                    TextEditor(text: $content)
                        .frame(minHeight: 150)
                        .font(.system(.caption, design: .monospaced))
                }
            }
            .navigationTitle("Store Entry")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button {
                        isStoring = true
                        Task {
                            await onStore()
                            isStoring = false
                        }
                    } label: {
                        if isStoring {
                            ProgressView()
                        } else {
                            Text("Store")
                        }
                    }
                    .disabled(source.isEmpty || content.isEmpty || isStoring)
                }
            }
        }
    }
}

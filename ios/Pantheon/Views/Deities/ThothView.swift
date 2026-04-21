import SwiftUI

/// 𓁟 Thoth — Memory & Context management view.
/// Tabbed interface consolidating project memory, output filtering (RTK), and context vault.
struct ThothView: View {
    @EnvironmentObject var appState: AppState

    enum ThothTab: String, CaseIterable {
        case memory = "Memory"
        case filter = "Filter"
        case vault = "Vault"
    }

    @State private var selectedTab: ThothTab = .memory

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                DeityHeader(
                    glyph: "\u{1305F}",
                    name: "Thoth",
                    subtitle: "Memory & Context",
                    description: "Persistent AI project memory, output filtering, and context search."
                )

                Picker("Tab", selection: $selectedTab) {
                    ForEach(ThothTab.allCases, id: \.self) { tab in
                        Text(tab.rawValue).tag(tab)
                    }
                }
                .pickerStyle(.segmented)

                switch selectedTab {
                case .memory:
                    ThothMemoryContent()
                case .filter:
                    ThothFilterContent()
                case .vault:
                    ThothVaultContent()
                }
            }
            .padding()
        }
        .background(PantheonTheme.background)
    }
}

// MARK: - Memory Tab (formerly ThothView body)

struct ThothMemoryContent: View {
    @EnvironmentObject var appState: AppState
    @State private var projectInfo: ProjectInfo?
    @State private var selectedProject: URL?
    @State private var isWorking = false
    @State private var statusMessage: String?
    @State private var errorMessage: String?
    @State private var showDocumentPicker = false

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            // Project selector
            Button {
                showDocumentPicker = true
            } label: {
                HStack {
                    Image(systemName: "folder.badge.plus")
                    Text(selectedProject?.lastPathComponent ?? "Select Project Folder")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(PantheonTheme.surfaceElevated)
                .foregroundStyle(PantheonTheme.textPrimary)
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }
            .sheet(isPresented: $showDocumentPicker) {
                DocumentPickerView(selectedURL: $selectedProject)
            }

            // Project info card
            if let info = projectInfo {
                ProjectInfoCard(info: info)
            }

            // Action buttons
            if selectedProject != nil {
                HStack(spacing: 12) {
                    ThothActionButton(title: "Init", icon: "plus.circle", isWorking: isWorking) {
                        await thothInit()
                    }
                    ThothActionButton(title: "Sync", icon: "arrow.triangle.2.circlepath", isWorking: isWorking) {
                        await thothSync()
                    }
                    ThothActionButton(title: "Compact", icon: "rectangle.compress.vertical", isWorking: isWorking) {
                        await thothCompact()
                    }
                }
            }

            if let statusMessage {
                Text(statusMessage)
                    .font(.subheadline)
                    .foregroundStyle(PantheonTheme.success)
                    .padding()
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(PantheonTheme.success.opacity(0.1))
                    .clipShape(RoundedRectangle(cornerRadius: 8))
            }

            if let errorMessage {
                ErrorRetryView(message: errorMessage) { await thothSync() }
            }
        }
    }

    private func thothInit() async {
        guard let path = selectedProject?.path else { return }
        isWorking = true
        errorMessage = nil
        defer { isWorking = false }

        do {
            projectInfo = try await appState.bridge.thothInit(projectRoot: path)
            statusMessage = ".thoth/ initialized successfully"
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func thothSync() async {
        guard let path = selectedProject?.path else { return }
        isWorking = true
        errorMessage = nil
        defer { isWorking = false }

        do {
            try await appState.bridge.thothSync(root: path)
            statusMessage = "Project memory synced"
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func thothCompact() async {
        guard let path = selectedProject?.path else { return }
        isWorking = true
        errorMessage = nil
        defer { isWorking = false }

        do {
            try await appState.bridge.thothCompact(root: path)
            statusMessage = "Context compacted"
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

// MARK: - Filter Tab (formerly RTKView body)

struct ThothFilterContent: View {
    @EnvironmentObject var appState: AppState
    @State private var rawInput = ""
    @State private var filterResult: FilterResult?
    @State private var filterConfig: FilterConfig?
    @State private var isFiltering = false
    @State private var errorMessage: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
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

// MARK: - Vault Tab (formerly VaultView body)

struct ThothVaultContent: View {
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
        VStack(alignment: .leading, spacing: 16) {
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

// MARK: - Shared Subviews

struct ThothActionButton: View {
    let title: String
    let icon: String
    let isWorking: Bool
    let action: () async -> Void

    var body: some View {
        Button {
            Task { await action() }
        } label: {
            VStack(spacing: 6) {
                Image(systemName: icon)
                    .font(.title2)
                Text(title)
                    .font(.caption)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(PantheonTheme.gold)
            .foregroundStyle(.black)
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
        .disabled(isWorking)
    }
}

struct ProjectInfoCard: View {
    let info: ProjectInfo

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Detected Project")
                .font(.headline)
                .foregroundStyle(PantheonTheme.gold)

            if let name = info.name {
                InfoRow(label: "Name", value: name)
            }
            if let language = info.language {
                InfoRow(label: "Language", value: language)
            }
            if let version = info.version {
                InfoRow(label: "Version", value: version)
            }
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct InfoRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label)
                .font(.subheadline)
                .foregroundStyle(PantheonTheme.textSecondary)
            Spacer()
            Text(value)
                .font(.subheadline.bold())
                .foregroundStyle(PantheonTheme.textPrimary)
        }
    }
}

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

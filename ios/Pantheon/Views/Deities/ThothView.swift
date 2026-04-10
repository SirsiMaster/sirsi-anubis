import SwiftUI

/// 𓁟 Thoth — Project memory management view.
/// Initialize, sync, and compact AI project memory.
struct ThothView: View {
    @EnvironmentObject var appState: AppState
    @State private var projectInfo: ProjectInfo?
    @State private var selectedProject: URL?
    @State private var isWorking = false
    @State private var statusMessage: String?
    @State private var errorMessage: String?
    @State private var showDocumentPicker = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                DeityHeader(
                    glyph: "𓁟",
                    name: "Thoth",
                    subtitle: "The Scribe",
                    description: "Persistent AI project memory — init, sync, and compact knowledge across sessions."
                )

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
                    ErrorBanner(message: errorMessage)
                }
            }
            .padding()
        }
        .background(PantheonTheme.background)
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

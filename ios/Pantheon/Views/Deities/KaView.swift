import SwiftUI

/// 𓂓 Ka — Ghost detection view.
/// Detects residuals from uninstalled applications.
struct KaView: View {
    @EnvironmentObject var appState: AppState
    @State private var ghosts: [GhostApp] = []
    @State private var isScanning = false
    @State private var errorMessage: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                DeityHeader(
                    glyph: "𓂓",
                    name: "Ka",
                    subtitle: "The Spirit Finder",
                    description: "Detect dead app remnants — preferences, caches, and support files left behind by uninstalled apps."
                )

                Button {
                    Task { await huntGhosts() }
                } label: {
                    HStack {
                        Image(systemName: isScanning ? "progress.indicator" : "eye.trianglebadge.exclamationmark")
                        Text(isScanning ? "Hunting..." : "Hunt Ghosts")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(PantheonTheme.gold)
                    .foregroundStyle(.black)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .font(.headline)
                }
                .disabled(isScanning)

                if let errorMessage {
                    ErrorBanner(message: errorMessage)
                }

                if !ghosts.isEmpty {
                    GhostSummaryCard(ghosts: ghosts)
                }

                ForEach(ghosts) { ghost in
                    GhostRow(ghost: ghost)
                }

                if !isScanning && ghosts.isEmpty && errorMessage == nil {
                    ContentUnavailableView(
                        "No Ghosts Detected",
                        systemImage: "checkmark.shield.fill",
                        description: Text("Your system is clean — no app residuals found.")
                    )
                    .foregroundStyle(PantheonTheme.textSecondary)
                }
            }
            .padding()
        }
        .background(PantheonTheme.background)
    }

    private func huntGhosts() async {
        isScanning = true
        errorMessage = nil
        defer { isScanning = false }

        do {
            ghosts = try await appState.bridge.kaHunt()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct GhostSummaryCard: View {
    let ghosts: [GhostApp]

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("𓂓 \(ghosts.count) Ghost\(ghosts.count == 1 ? "" : "s") Found")
                .font(.headline)
                .foregroundStyle(PantheonTheme.warning)

            let totalSize = ghosts.reduce(Int64(0)) { $0 + $1.totalSize }
            Text("Total reclaimable: \(ByteCountFormatter.string(fromByteCount: totalSize, countStyle: .file))")
                .font(.subheadline)
                .foregroundStyle(PantheonTheme.textSecondary)
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct GhostRow: View {
    let ghost: GhostApp
    @State private var isExpanded = false

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Button { isExpanded.toggle() } label: {
                HStack {
                    VStack(alignment: .leading, spacing: 4) {
                        Text(ghost.appName)
                            .font(.subheadline.bold())
                            .foregroundStyle(PantheonTheme.textPrimary)
                        Text(ghost.bundleId)
                            .font(.caption)
                            .foregroundStyle(PantheonTheme.textSecondary)
                    }

                    Spacer()

                    VStack(alignment: .trailing) {
                        Text(ghost.formattedSize)
                            .font(.subheadline.bold())
                            .foregroundStyle(PantheonTheme.gold)
                        Text("\(ghost.residuals.count) residuals")
                            .font(.caption)
                            .foregroundStyle(PantheonTheme.textSecondary)
                    }

                    Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                        .foregroundStyle(PantheonTheme.textSecondary)
                        .font(.caption)
                }
            }

            if isExpanded {
                ForEach(ghost.residuals) { residual in
                    HStack {
                        Image(systemName: "folder.fill")
                            .foregroundStyle(PantheonTheme.textSecondary)
                            .font(.caption)
                        Text(residual.path)
                            .font(.caption)
                            .foregroundStyle(PantheonTheme.textSecondary)
                            .lineLimit(1)
                        Spacer()
                        Text(residual.formattedSize)
                            .font(.caption)
                            .foregroundStyle(PantheonTheme.gold)
                    }
                    .padding(.leading, 8)
                }
            }
        }
        .padding()
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}

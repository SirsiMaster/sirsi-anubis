import Foundation

// MARK: - Ka (Ghost Detection)

struct GhostApp: Codable, Identifiable {
    var id: String { appName }

    let appName: String
    let bundleId: String
    let residuals: [Residual]
    let totalSize: Int64
    let totalFiles: Int
    let inLaunchServices: Bool
    let detectionMethod: String?

    var formattedSize: String {
        ByteCountFormatter.string(fromByteCount: totalSize, countStyle: .file)
    }
}

struct Residual: Codable, Identifiable {
    var id: String { path }

    let path: String
    let type: String
    let sizeBytes: Int64
    let fileCount: Int
    let requiresSudo: Bool

    var formattedSize: String {
        ByteCountFormatter.string(fromByteCount: sizeBytes, countStyle: .file)
    }
}

struct InstalledApp: Codable, Identifiable {
    var id: String { bundleId }

    let name: String
    let bundleId: String
    let path: String
    let version: String?
    let source: String?
    let size: Int64
    let lastUsed: String?
    let isRunning: Bool
    let hasGhosts: Bool
    let ghostCount: Int
    let ghostSize: Int64
}

import Foundation

// MARK: - Vault (Context Keeper)

struct VaultEntry: Codable, Identifiable {
    let id: Int64
    let source: String
    let tag: String
    let content: String?
    let tokens: Int
    let createdAt: String?
    let snippet: String?
}

struct VaultSearchResult: Codable {
    let query: String
    let totalHits: Int
    let entries: [VaultEntry]
}

struct VaultStoreStats: Codable {
    let totalEntries: Int
    let totalBytes: Int64
    let totalTokens: Int64
    let oldestEntry: String?
    let newestEntry: String?
    let tagCounts: [String: Int]?

    var formattedBytes: String {
        ByteCountFormatter.string(fromByteCount: totalBytes, countStyle: .file)
    }
}

struct VaultPruneResult: Codable {
    let pruned: Int
}

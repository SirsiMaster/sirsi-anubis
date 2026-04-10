import Foundation

// MARK: - Anubis (Jackal Scanner)

struct ScanCategory: Codable, Identifiable {
    let id: String
    let displayName: String
}

struct ScanResult: Codable {
    let findings: [Finding]
    let totalSize: Int64
    let rulesRan: Int
    let rulesWithFindings: Int
    let errors: [RuleError]?
    let byCategory: [String: CategorySummary]?
}

struct Finding: Codable, Identifiable {
    var id: String { "\(ruleName):\(path)" }

    let ruleName: String
    let category: String
    let description: String
    let path: String
    let sizeBytes: Int64
    let fileCount: Int
    let severity: String
    let lastModified: String?
    let requiresSudo: Bool
    let isDir: Bool

    var formattedSize: String {
        ByteCountFormatter.string(fromByteCount: sizeBytes, countStyle: .file)
    }
}

struct RuleError: Codable, Identifiable {
    var id: String { ruleName }
    let ruleName: String
    let error: String
}

struct CategorySummary: Codable {
    let totalSize: Int64
    let findingCount: Int
}

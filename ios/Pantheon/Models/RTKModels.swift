import Foundation

// MARK: - RTK (Output Filter)

struct FilterConfig: Codable {
    let stripAnsi: Bool
    let dedup: Bool
    let dedupWindow: Int
    let maxLines: Int
    let maxBytes: Int
    let tailLines: Int
    let collapseBlank: Bool
}

struct FilterResult: Codable {
    let output: String
    let originalBytes: Int
    let filteredBytes: Int
    let linesRemoved: Int
    let dupsCollapsed: Int
    let truncated: Bool
    let ratio: Double

    var reductionPercent: String {
        let pct = (1.0 - ratio) * 100.0
        return String(format: "%.1f%%", pct)
    }

    var formattedOriginal: String {
        ByteCountFormatter.string(fromByteCount: Int64(originalBytes), countStyle: .file)
    }

    var formattedFiltered: String {
        ByteCountFormatter.string(fromByteCount: Int64(filteredBytes), countStyle: .file)
    }
}

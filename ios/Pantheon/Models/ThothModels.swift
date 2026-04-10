import Foundation

// MARK: - Thoth (Project Memory)

struct ProjectInfo: Codable {
    let name: String?
    let language: String?
    let version: String?
    let root: String?
}

struct JournalEntry: Codable, Identifiable {
    var id: Int { number }

    let number: Int
    let date: String
    let title: String
    let commits: [CommitInfo]?
}

struct CommitInfo: Codable, Identifiable {
    var id: String { hash }

    let hash: String
    let subject: String
    let date: String
    let files: Int?
    let adds: Int?
    let dels: Int?
}

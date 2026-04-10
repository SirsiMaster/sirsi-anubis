import Foundation

// MARK: - Seshat (Knowledge Bridge)

struct KnowledgeSource: Codable, Identifiable {
    var id: String { name }

    let name: String
    let description: String
}

struct IngestResult: Codable, Identifiable {
    var id: String { source }

    let source: String
    let count: Int
    let error: String?
}

struct KnowledgeItem: Codable, Identifiable {
    var id: String { title }

    let title: String
    let summary: String?
    let references: [KIReference]?
}

struct KIReference: Codable {
    let type: String
    let value: String
}

struct Conversation: Codable, Identifiable {
    let id: String
    let title: String?
    let startedAt: String?
    let messageCount: Int?
    let messages: [ConversationMessage]?
}

struct ConversationMessage: Codable, Identifiable {
    var id: String { "\(role):\(timestamp ?? "unknown")" }

    let role: String
    let content: String
    let timestamp: String?
}

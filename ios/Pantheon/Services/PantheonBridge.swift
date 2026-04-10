import Foundation
import PantheonCore  // gomobile-generated framework

/// Bridge between Swift and the Go mobile package.
/// All Go functions return JSON strings; this layer deserializes them into Swift types.
final class PantheonBridge: Sendable {

    // MARK: - Response envelope (matches mobile.Response in Go)

    struct BridgeResponse<T: Decodable>: Decodable {
        let ok: Bool
        let data: T?
        let error: String?
    }

    // MARK: - Anubis

    func anubisCategories() throws -> [ScanCategory] {
        let json = MobileAnubisCategories()
        return try decode(json)
    }

    func anubisScan(rootPath: String, categories: [String] = []) async throws -> ScanResult {
        let options = try JSONEncoder().encode(["categories": categories])
        let optionsStr = String(data: options, encoding: .utf8) ?? ""

        return try await Task.detached(priority: .userInitiated) {
            let json = MobileAnubisScan(rootPath, optionsStr)
            return try self.decode(json) as ScanResult
        }.value
    }

    // MARK: - Ka

    func kaHunt(includeSudo: Bool = false) async throws -> [GhostApp] {
        return try await Task.detached(priority: .userInitiated) {
            let json = MobileKaHunt(includeSudo)
            return try self.decode(json) as [GhostApp]
        }.value
    }

    func kaEnumerateApps() async throws -> [InstalledApp] {
        return try await Task.detached(priority: .userInitiated) {
            let json = MobileKaEnumerateApps()
            return try self.decode(json) as [InstalledApp]
        }.value
    }

    // MARK: - Thoth

    func thothInit(projectRoot: String) async throws -> ProjectInfo {
        return try await Task.detached(priority: .userInitiated) {
            let json = MobileThothInit(projectRoot)
            return try self.decode(json) as ProjectInfo
        }.value
    }

    func thothSync(root: String) async throws {
        let options = try JSONEncoder().encode(["root": root])
        let optionsStr = String(data: options, encoding: .utf8) ?? ""

        try await Task.detached(priority: .userInitiated) {
            let json = MobileThothSync(optionsStr)
            let _: [String: String] = try self.decode(json)
        }.value
    }

    func thothCompact(root: String) async throws {
        let options = try JSONEncoder().encode(["repo_root": root])
        let optionsStr = String(data: options, encoding: .utf8) ?? ""

        try await Task.detached(priority: .userInitiated) {
            let json = MobileThothCompact(optionsStr)
            let _: [String: String] = try self.decode(json)
        }.value
    }

    func thothDetectProject(root: String) throws -> ProjectInfo {
        let json = MobileThothDetectProject(root)
        return try decode(json)
    }

    // MARK: - Seba

    func sebaDetectHardware() async throws -> HardwareProfile {
        return try await Task.detached(priority: .userInitiated) {
            let json = MobileSebaDetectHardware()
            return try self.decode(json) as HardwareProfile
        }.value
    }

    func sebaDetectAccelerators() async throws -> AcceleratorProfile {
        return try await Task.detached(priority: .userInitiated) {
            let json = MobileSebaDetectAccelerators()
            return try self.decode(json) as AcceleratorProfile
        }.value
    }

    // MARK: - Seshat

    func seshatListSources() throws -> [KnowledgeSource] {
        let json = MobileSeshatListSources()
        return try decode(json)
    }

    func seshatIngest(sources: [String], sinceDays: Int = 7) async throws -> [IngestResult] {
        struct IngestOptions: Encodable {
            let sources: [String]
            let sinceDays: Int

            enum CodingKeys: String, CodingKey {
                case sources
                case sinceDays = "since_days"
            }
        }
        let options = try JSONEncoder().encode(IngestOptions(sources: sources, sinceDays: sinceDays))
        let optionsStr = String(data: options, encoding: .utf8) ?? ""

        return try await Task.detached(priority: .userInitiated) {
            let json = MobileSeshatIngest(optionsStr)
            return try self.decode(json) as [IngestResult]
        }.value
    }

    // MARK: - Version

    func version() -> String {
        return MobileVersion()
    }

    // MARK: - JSON Decoding

    private func decode<T: Decodable>(_ json: String) throws -> T {
        guard let data = json.data(using: .utf8) else {
            throw BridgeError.invalidJSON
        }

        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase

        let response = try decoder.decode(BridgeResponse<T>.self, from: data)

        guard response.ok, let result = response.data else {
            throw BridgeError.goError(response.error ?? "unknown error")
        }

        return result
    }
}

// MARK: - Bridge Errors

enum BridgeError: LocalizedError {
    case invalidJSON
    case goError(String)

    var errorDescription: String? {
        switch self {
        case .invalidJSON:
            return "Invalid JSON response from Pantheon core"
        case .goError(let message):
            return "Pantheon: \(message)"
        }
    }
}

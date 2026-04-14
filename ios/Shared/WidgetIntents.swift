import AppIntents
import WidgetKit
import PantheonCore

// MARK: - Anubis: Trigger Scan from Widget

struct AnubisWidgetScanIntent: AppIntent {
    static let title: LocalizedStringResource = "Scan for Waste"
    static let description: IntentDescription = "Run an Anubis infrastructure scan."
    static let isDiscoverable = false

    func perform() async throws -> some IntentResult {
        let json = MobileAnubisScan("", "")
        if let data = json.data(using: .utf8),
           let response = try? JSONDecoder().decode(WidgetScanResponse.self, from: data),
           response.ok, let scan = response.data {
            let sorted = scan.findings.sorted { $0.sizeBytes > $1.sizeBytes }
            let top = sorted.prefix(3).map { ($0.description, $0.sizeBytes) }
            SharedDataManager.saveScanResults(
                findingCount: scan.findings.count,
                totalSize: scan.totalSize,
                topFindings: top,
                rulesRan: scan.rulesRan
            )
        }
        return .result()
    }
}

// MARK: - Seba: Refresh Hardware from Widget

struct SebaWidgetRefreshIntent: AppIntent {
    static let title: LocalizedStringResource = "Refresh Hardware"
    static let description: IntentDescription = "Refresh Seba hardware profile."
    static let isDiscoverable = false

    func perform() async throws -> some IntentResult {
        let json = MobileSebaDetectHardware()
        SharedDataManager.saveHardwareJSON(json)
        return .result()
    }
}

// MARK: - Decodable helpers (shared between widget intents)

struct WidgetScanResponse: Decodable {
    let ok: Bool
    let data: WidgetScanData?
    let error: String?
}

struct WidgetScanData: Decodable {
    let findings: [WidgetFinding]
    let totalSize: Int64
    let rulesRan: Int

    enum CodingKeys: String, CodingKey {
        case findings
        case totalSize = "TotalSize"
        case rulesRan = "RulesRan"
    }
}

struct WidgetFinding: Decodable {
    let description: String
    let sizeBytes: Int64

    enum CodingKeys: String, CodingKey {
        case description = "Description"
        case sizeBytes = "SizeBytes"
    }
}

import Foundation
import WidgetKit

/// Manages shared data between the main app and widget extensions via App Group container.
/// Widgets cannot call PantheonCore directly in the background — the app writes scan results
/// to the shared container, and widgets read from it.
enum SharedDataManager {
    static let appGroupID = "group.ai.sirsi.pantheon"

    private static var defaults: UserDefaults? {
        UserDefaults(suiteName: appGroupID)
    }

    // MARK: - Keys

    private enum Key {
        static let lastScanDate = "lastScanDate"
        static let scanFindingCount = "scanFindingCount"
        static let scanTotalSize = "scanTotalSize"
        static let scanTopFindings = "scanTopFindings"
        static let scanRulesRan = "scanRulesRan"
        static let hardwareJSON = "hardwareJSON"
    }

    // MARK: - Write (Main App)

    static func saveScanResults(
        findingCount: Int,
        totalSize: Int64,
        topFindings: [(name: String, size: Int64)],
        rulesRan: Int
    ) {
        guard let defaults else { return }
        defaults.set(Date(), forKey: Key.lastScanDate)
        defaults.set(findingCount, forKey: Key.scanFindingCount)
        defaults.set(totalSize, forKey: Key.scanTotalSize)
        defaults.set(rulesRan, forKey: Key.scanRulesRan)

        let encoded = topFindings.map { ["name": $0.name, "size": String($0.size)] }
        defaults.set(encoded, forKey: Key.scanTopFindings)

        WidgetCenter.shared.reloadTimelines(ofKind: "ai.sirsi.pantheon.anubis")
    }

    static func saveHardwareJSON(_ json: String) {
        defaults?.set(json, forKey: Key.hardwareJSON)
        WidgetCenter.shared.reloadTimelines(ofKind: "ai.sirsi.pantheon.seba")
    }

    // MARK: - Read (Widgets)

    struct ScanSnapshot {
        let date: Date?
        let findingCount: Int
        let totalSize: Int64
        let topFindings: [(name: String, size: Int64)]
        let rulesRan: Int
        var isStale: Bool {
            guard let date else { return true }
            return Date().timeIntervalSince(date) > 3600
        }
    }

    static func loadScanSnapshot() -> ScanSnapshot? {
        guard let defaults else { return nil }
        guard defaults.object(forKey: Key.scanFindingCount) != nil else { return nil }

        let findings: [(String, Int64)]
        if let raw = defaults.array(forKey: Key.scanTopFindings) as? [[String: String]] {
            findings = raw.compactMap { dict in
                guard let name = dict["name"],
                      let sizeStr = dict["size"],
                      let size = Int64(sizeStr) else { return nil }
                return (name, size)
            }
        } else {
            findings = []
        }

        return ScanSnapshot(
            date: defaults.object(forKey: Key.lastScanDate) as? Date,
            findingCount: defaults.integer(forKey: Key.scanFindingCount),
            totalSize: Int64(defaults.integer(forKey: Key.scanTotalSize)),
            topFindings: findings,
            rulesRan: defaults.integer(forKey: Key.scanRulesRan)
        )
    }

    static func loadHardwareJSON() -> String? {
        defaults?.string(forKey: Key.hardwareJSON)
    }
}

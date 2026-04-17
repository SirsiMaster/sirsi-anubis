import SwiftUI
import Combine

/// Central app state shared across all views.
/// Manages the active deity, view mode (TUI vs GUI), and bridge lifecycle.
@MainActor
final class AppState: ObservableObject {
    enum ViewMode: String, CaseIterable {
        case gui = "GUI"
        case tui = "TUI"
    }

    enum ActiveDeity: String, CaseIterable, Identifiable {
        case anubis = "Anubis"
        case ka = "Ka"
        case thoth = "Thoth"
        case seba = "Seba"
        case seshat = "Seshat"

        var id: String { rawValue }

        var glyph: String {
            switch self {
            case .anubis: return "𓁢"
            case .ka:     return "𓂓"
            case .thoth:  return "𓁟"
            case .seba:   return "𓇽"
            case .seshat: return "𓁆"
            }
        }

        var subtitle: String {
            switch self {
            case .anubis: return "Infrastructure Scanner"
            case .ka:     return "Ghost Detection"
            case .thoth:  return "Project Memory"
            case .seba:   return "Hardware Profiling"
            case .seshat: return "Knowledge Bridge"
            }
        }

        var iconName: String {
            switch self {
            case .anubis: return "magnifyingglass.circle.fill"
            case .ka:     return "eye.trianglebadge.exclamationmark.fill"
            case .thoth:  return "brain.head.profile.fill"
            case .seba:   return "cpu.fill"
            case .seshat: return "books.vertical.fill"
            }
        }
    }

    @Published var viewMode: ViewMode = .gui
    @Published var activeDeity: ActiveDeity = .anubis
    @Published var isRunning = false

    let bridge = PantheonBridge()

    // MARK: - Deep Links (sirsi://deity/{name})

    func handleDeepLink(_ url: URL) {
        guard url.scheme == "sirsi" else { return }
        switch url.host {
        case "anubis": activeDeity = .anubis
        case "ka":     activeDeity = .ka
        case "thoth":  activeDeity = .thoth
        case "seba":   activeDeity = .seba
        case "seshat": activeDeity = .seshat
        default: break
        }
        viewMode = .gui
    }
}

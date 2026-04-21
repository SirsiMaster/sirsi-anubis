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
        case brain = "Brain"
        case horus = "Horus"
        case stele = "Stele"

        var id: String { rawValue }

        var glyph: String {
            switch self {
            case .anubis: return "\u{13062}"
            case .ka:     return "\u{13093}"
            case .thoth:  return "\u{1305F}"
            case .seba:   return "\u{133BD}"
            case .seshat: return "\u{13046}"
            case .brain:  return "\u{130A7}"
            case .horus:  return "\u{13080}"
            case .stele:  return "\u{130BD}"
            }
        }

        var subtitle: String {
            switch self {
            case .anubis: return "Infrastructure Scanner"
            case .ka:     return "Ghost Detection"
            case .thoth:  return "Memory & Context"
            case .seba:   return "Hardware Profiling"
            case .seshat: return "Knowledge Bridge"
            case .brain:  return "Neural Classification"
            case .horus:  return "Code Graph"
            case .stele:  return "Event Ledger"
            }
        }

        var iconName: String {
            switch self {
            case .anubis: return "magnifyingglass.circle.fill"
            case .ka:     return "eye.trianglebadge.exclamationmark.fill"
            case .thoth:  return "brain.head.profile.fill"
            case .seba:   return "cpu.fill"
            case .seshat: return "books.vertical.fill"
            case .brain:  return "brain.fill"
            case .horus:  return "eye.circle.fill"
            case .stele:  return "list.bullet.rectangle.fill"
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
        case "brain":  activeDeity = .brain
        case "rtk":    activeDeity = .thoth  // RTK folded into Thoth
        case "vault":  activeDeity = .thoth  // Vault folded into Thoth
        case "horus":  activeDeity = .horus
        case "stele":  activeDeity = .stele
        default: break
        }
        viewMode = .gui
    }
}

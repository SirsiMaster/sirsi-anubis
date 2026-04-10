import SwiftUI

/// Pantheon brand theme — Egyptian gold on black, matching the CLI terminal output.
enum PantheonTheme {
    // MARK: - Core Colors
    static let gold = Color(hex: 0xC8A951)
    static let deepLapis = Color(hex: 0x1A1A5E)
    static let background = Color(hex: 0x0F0F0F)
    static let surface = Color(hex: 0x1A1A1A)
    static let surfaceElevated = Color(hex: 0x252525)
    static let textPrimary = Color.white
    static let textSecondary = Color.white.opacity(0.7)
    static let success = Color(hex: 0x4CAF50)
    static let warning = Color(hex: 0xFFA726)
    static let error = Color(hex: 0xEF5350)

    // MARK: - Severity Colors (scan results)
    static func severityColor(_ severity: String) -> Color {
        switch severity.lowercased() {
        case "safe":    return success
        case "caution": return warning
        case "warning": return error
        default:        return textSecondary
        }
    }

    // MARK: - TUI Terminal
    static let tuiBackground = Color.black
    static let tuiText = Color(hex: 0xE0E0E0)
    static let tuiPrompt = gold
    static let tuiFont: Font = .system(size: 13, design: .monospaced)
}

// MARK: - Color hex initializer
extension Color {
    init(hex: UInt, alpha: Double = 1.0) {
        self.init(
            .sRGB,
            red: Double((hex >> 16) & 0xFF) / 255.0,
            green: Double((hex >> 8) & 0xFF) / 255.0,
            blue: Double(hex & 0xFF) / 255.0,
            opacity: alpha
        )
    }
}

import SwiftUI

// MARK: - Preview AppState

@MainActor
private func previewAppState(_ deity: AppState.ActiveDeity = .anubis) -> AppState {
    let state = AppState()
    state.activeDeity = deity
    return state
}

// MARK: - Content View Previews

#Preview("iPhone — Anubis") {
    ContentView()
        .environmentObject(previewAppState(.anubis))
        .preferredColorScheme(.dark)
}

#Preview("iPhone — Thoth") {
    ContentView()
        .environmentObject(previewAppState(.thoth))
        .preferredColorScheme(.dark)
}

#Preview("iPad — Sidebar", traits: .landscapeLeft) {
    ContentView()
        .environmentObject(previewAppState(.seba))
        .preferredColorScheme(.dark)
        .previewDevice("iPad Pro (12.9-inch) (6th generation)")
}

// MARK: - Individual Deity Previews

#Preview("Anubis") {
    NavigationStack {
        AnubisView()
            .environmentObject(previewAppState(.anubis))
    }
    .preferredColorScheme(.dark)
}

#Preview("Ka") {
    NavigationStack {
        KaView()
            .environmentObject(previewAppState(.ka))
    }
    .preferredColorScheme(.dark)
}

#Preview("Thoth") {
    NavigationStack {
        ThothView()
            .environmentObject(previewAppState(.thoth))
    }
    .preferredColorScheme(.dark)
}

#Preview("Seba") {
    NavigationStack {
        SebaView()
            .environmentObject(previewAppState(.seba))
    }
    .preferredColorScheme(.dark)
}

#Preview("Seshat") {
    NavigationStack {
        SeshatView()
            .environmentObject(previewAppState(.seshat))
    }
    .preferredColorScheme(.dark)
}

// MARK: - Shared Component Previews

#Preview("Deity Header") {
    DeityHeader(
        glyph: "𓁢",
        name: "Anubis",
        subtitle: "Weigh. Judge. Purge.",
        description: "Scan for infrastructure waste — caches, build artifacts, orphaned data."
    )
    .padding()
    .background(PantheonTheme.background)
    .preferredColorScheme(.dark)
}

#Preview("Error Banner") {
    ErrorBanner(message: "Failed to connect to PantheonCore bridge")
        .padding()
        .background(PantheonTheme.background)
        .preferredColorScheme(.dark)
}

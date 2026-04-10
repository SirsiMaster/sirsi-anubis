import SwiftUI

/// Root content view with tab bar switching between deity views.
/// Toggle between GUI (native SwiftUI) and TUI (terminal emulator) modes.
struct ContentView: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Mode toggle
                Picker("Mode", selection: $appState.viewMode) {
                    ForEach(AppState.ViewMode.allCases, id: \.self) { mode in
                        Text(mode.rawValue).tag(mode)
                    }
                }
                .pickerStyle(.segmented)
                .padding(.horizontal)
                .padding(.top, 8)

                // Main content
                Group {
                    switch appState.viewMode {
                    case .gui:
                        GUIContainerView()
                    case .tui:
                        TUIContainerView()
                    }
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
            .background(PantheonTheme.background)
            .navigationTitle(deityTitle)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .principal) {
                    HStack(spacing: 6) {
                        Text(appState.activeDeity.glyph)
                            .font(.title2)
                        Text(appState.activeDeity.rawValue)
                            .font(.headline)
                            .foregroundStyle(PantheonTheme.gold)
                    }
                }
            }
        }
        .tint(PantheonTheme.gold)
    }

    private var deityTitle: String {
        "\(appState.activeDeity.glyph) \(appState.activeDeity.rawValue)"
    }
}

/// GUI mode: native SwiftUI views with tab bar for deity selection.
struct GUIContainerView: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        TabView(selection: $appState.activeDeity) {
            AnubisView()
                .tabItem {
                    Label("Anubis", systemImage: "magnifyingglass.circle.fill")
                }
                .tag(AppState.ActiveDeity.anubis)

            KaView()
                .tabItem {
                    Label("Ka", systemImage: "eye.trianglebadge.exclamationmark.fill")
                }
                .tag(AppState.ActiveDeity.ka)

            ThothView()
                .tabItem {
                    Label("Thoth", systemImage: "brain.head.profile.fill")
                }
                .tag(AppState.ActiveDeity.thoth)

            SebaView()
                .tabItem {
                    Label("Seba", systemImage: "cpu.fill")
                }
                .tag(AppState.ActiveDeity.seba)

            SeshatView()
                .tabItem {
                    Label("Seshat", systemImage: "books.vertical.fill")
                }
                .tag(AppState.ActiveDeity.seshat)
        }
    }
}

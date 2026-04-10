# Pantheon iOS

Native SwiftUI app with Go core via gomobile. Supports both GUI (native views) and TUI (terminal emulator) modes.

## Deities Available

| Deity | Glyph | Function | iOS Status |
|-------|-------|----------|------------|
| Anubis | 𓁢 | Infrastructure scan | Sandbox-scoped |
| Ka | 𓂓 | Ghost detection | Sandbox-scoped |
| Thoth | 𓁟 | Project memory | Full support |
| Seba | 𓇽 | Hardware profiling | Full support |
| Seshat | 𓁆 | Knowledge bridge | Source-dependent |

## Architecture

```
┌─────────────────────────────┐
│  SwiftUI App (Views, Nav)   │  100% native iOS
├─────────────────────────────┤
│  PantheonBridge.swift       │  JSON-based Go↔Swift bridge
├─────────────────────────────┤
│  PantheonCore.xcframework   │  Go core via gomobile
│  (jackal, ka, thoth, seba,  │
│   seshat)                   │
└─────────────────────────────┘
```

## Prerequisites

```bash
# Go 1.24+
go version

# gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Xcode 16+ with iOS SDK
xcode-select -p
```

## Build

### 1. Build the Go framework

```bash
make ios-framework
```

This produces `bin/ios/PantheonCore.xcframework`.

### 2. Set up Xcode project

1. Open Xcode → File → New → Project → iOS App
2. Product Name: `Pantheon`, Organization: `ai.sirsi`
3. Interface: SwiftUI, Language: Swift
4. Add `PantheonCore.xcframework` to Frameworks, Libraries
5. Drag the `ios/Pantheon/` source files into the project

### 3. Build the app

```bash
make ios
```

Or build directly from Xcode for simulator/device.

## Project Structure

```
ios/Pantheon/
├── App/
│   ├── PantheonApp.swift         # @main entry point
│   ├── AppState.swift            # Shared state, deity registry
│   └── ContentView.swift         # Root view, GUI/TUI toggle
├── Views/
│   ├── Deities/
│   │   ├── AnubisView.swift      # 𓁢 Scanner GUI
│   │   ├── KaView.swift          # 𓂓 Ghost detection GUI
│   │   ├── ThothView.swift       # 𓁟 Project memory GUI
│   │   ├── SebaView.swift        # 𓇽 Hardware profiling GUI
│   │   └── SeshatView.swift      # 𓁆 Knowledge bridge GUI
│   ├── TUI/
│   │   └── TUIContainerView.swift  # Terminal emulator mode
│   └── Shared/
│       └── SharedComponents.swift  # DeityHeader, ErrorBanner, etc.
├── Services/
│   └── PantheonBridge.swift      # Go↔Swift JSON bridge
├── Models/
│   ├── AnubisModels.swift        # Finding, ScanResult
│   ├── KaModels.swift            # GhostApp, Residual
│   ├── ThothModels.swift         # ProjectInfo, JournalEntry
│   ├── SebaModels.swift          # HardwareProfile, AcceleratorProfile
│   └── SeshatModels.swift        # KnowledgeSource, Conversation
└── Theme/
    └── PantheonTheme.swift       # Gold/black brand colors

mobile/                           # Go bridge package (gomobile exports)
├── mobile.go                     # Response envelope, version
├── anubis.go                     # AnubisScan(), AnubisCategories()
├── ka.go                         # KaHunt(), KaEnumerateApps()
├── thoth.go                      # ThothInit(), ThothSync(), ThothCompact()
├── seba.go                       # SebaDetectHardware(), SebaDetectAccelerators()
└── seshat.go                     # SeshatIngest(), SeshatListSources()

internal/platform/ios.go          # iOS Platform implementation (build-tagged)
```

## iOS Sandbox Limitations

iOS sandboxing restricts what Pantheon can access compared to macOS:

- **Anubis**: Can only scan the app's own Documents/Caches, not system-wide
- **Ka**: Ghost detection limited to app sandbox residuals
- **Thoth**: Full functionality on repos cloned/opened via Files app
- **Seba**: Full hardware detection via sysctl and Metal APIs
- **Seshat**: Source availability depends on iOS permissions (no direct Chrome DB access)

## TUI Mode

The TUI mode provides a terminal-like interface within the app. Type commands like:
- `scan` — Run Anubis scan
- `ghosts` — Hunt Ka ghosts
- `hardware` — Seba hardware profile
- `seshat` — List knowledge sources
- `help` — Show all commands

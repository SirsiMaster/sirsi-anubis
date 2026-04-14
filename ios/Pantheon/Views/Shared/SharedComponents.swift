import SwiftUI

// MARK: - Deity Header

struct DeityHeader: View {
    let glyph: String
    let name: String
    let subtitle: String
    let description: String

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 10) {
                Text(glyph)
                    .font(.system(size: 36))
                VStack(alignment: .leading, spacing: 2) {
                    Text(name)
                        .font(.title2.bold())
                        .foregroundStyle(PantheonTheme.gold)
                    Text(subtitle)
                        .font(.caption)
                        .foregroundStyle(PantheonTheme.textSecondary)
                        .italic()
                }
            }
            Text(description)
                .font(.subheadline)
                .foregroundStyle(PantheonTheme.textSecondary)
        }
        .padding(.bottom, 4)
    }
}

// MARK: - Error Banner

struct ErrorBanner: View {
    let message: String

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundStyle(PantheonTheme.error)
            Text(message)
                .font(.subheadline)
                .foregroundStyle(PantheonTheme.error)
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PantheonTheme.error.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}

// MARK: - Stat Pill

struct StatPill: View {
    let label: String
    let value: String

    var body: some View {
        VStack(spacing: 4) {
            Text(value)
                .font(.headline)
                .foregroundStyle(PantheonTheme.textPrimary)
            Text(label)
                .font(.caption)
                .foregroundStyle(PantheonTheme.textSecondary)
        }
    }
}

// MARK: - Shimmer Loading Modifier

struct ShimmerModifier: ViewModifier {
    @State private var phase: CGFloat = -1

    func body(content: Content) -> some View {
        content
            .redacted(reason: .placeholder)
            .overlay {
                GeometryReader { geo in
                    LinearGradient(
                        colors: [.clear, .white.opacity(0.15), .clear],
                        startPoint: .leading,
                        endPoint: .trailing
                    )
                    .frame(width: geo.size.width * 0.4)
                    .offset(x: phase * geo.size.width)
                    .onAppear {
                        withAnimation(.linear(duration: 1.5).repeatForever(autoreverses: false)) {
                            phase = 1.5
                        }
                    }
                }
                .clipped()
            }
    }
}

extension View {
    func shimmer() -> some View {
        modifier(ShimmerModifier())
    }
}

// MARK: - Error View with Retry

struct ErrorRetryView: View {
    let message: String
    let retryAction: () async -> Void

    var body: some View {
        VStack(spacing: 12) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.largeTitle)
                .foregroundStyle(PantheonTheme.error)
            Text(message)
                .font(.subheadline)
                .foregroundStyle(PantheonTheme.textSecondary)
                .multilineTextAlignment(.center)
            Button {
                Task { await retryAction() }
            } label: {
                HStack(spacing: 6) {
                    Image(systemName: "arrow.clockwise")
                    Text("Retry")
                }
                .padding(.horizontal, 20)
                .padding(.vertical, 10)
                .background(PantheonTheme.gold)
                .foregroundStyle(.black)
                .clipShape(Capsule())
                .font(.subheadline.bold())
            }
        }
        .padding()
        .frame(maxWidth: .infinity)
        .background(PantheonTheme.error.opacity(0.05))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - Scan Result Skeleton (loading placeholder)

struct ScanResultSkeleton: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            VStack(alignment: .leading, spacing: 8) {
                Text("Scan Complete")
                    .font(.headline)
                HStack(spacing: 24) {
                    StatPill(label: "Findings", value: "12")
                    StatPill(label: "Reclaimable", value: "3.4 GB")
                    StatPill(label: "Rules", value: "58")
                }
            }
            .padding()
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(PantheonTheme.surface)
            .clipShape(RoundedRectangle(cornerRadius: 12))

            ForEach(0..<3, id: \.self) { _ in
                HStack {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Loading finding name here")
                            .font(.subheadline)
                        Text("/path/to/something/loading")
                            .font(.caption)
                    }
                    Spacer()
                    Text("1.2 GB")
                        .font(.subheadline.bold())
                }
                .padding()
                .background(PantheonTheme.surface)
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }
        }
        .shimmer()
    }
}

// MARK: - Hardware Skeleton

struct HardwareSkeleton: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Device Hardware")
                .font(.headline)
            InfoRow(label: "CPU", value: "Apple M4 Pro")
            InfoRow(label: "Architecture", value: "arm64")
            InfoRow(label: "Cores", value: "12")
            InfoRow(label: "RAM", value: "48 GB")
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shimmer()
    }
}

// MARK: - Document Picker (iOS)

struct DocumentPickerView: UIViewControllerRepresentable {
    @Binding var selectedURL: URL?

    func makeUIViewController(context: Context) -> UIDocumentPickerViewController {
        let picker = UIDocumentPickerViewController(forOpeningContentTypes: [.folder])
        picker.allowsMultipleSelection = false
        picker.delegate = context.coordinator
        return picker
    }

    func updateUIViewController(_ uiViewController: UIDocumentPickerViewController, context: Context) {}

    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }

    class Coordinator: NSObject, UIDocumentPickerDelegate {
        let parent: DocumentPickerView

        init(_ parent: DocumentPickerView) {
            self.parent = parent
        }

        func documentPicker(_ controller: UIDocumentPickerViewController, didPickDocumentsAt urls: [URL]) {
            parent.selectedURL = urls.first
        }
    }
}

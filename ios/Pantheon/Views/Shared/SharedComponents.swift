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

import SwiftUI
import UniformTypeIdentifiers

/// 𓂧 Brain — Neural classification inference view.
/// Classify files using the on-device inference backend (Stub/Spotlight/CoreML).
struct BrainView: View {
    @EnvironmentObject var appState: AppState
    @State private var modelInfo: ModelInfo?
    @State private var isLoadingModel = false

    // Quick Classify
    @State private var showFilePicker = false
    @State private var classifyResults: [FileClassification] = []
    @State private var isClassifying = false

    // Batch Analysis
    @State private var showDirectoryPicker = false
    @State private var batchResult: BatchClassificationResult?
    @State private var isBatchRunning = false

    // Legend
    @State private var legendExpanded = false

    // Errors
    @State private var errorMessage: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                DeityHeader(
                    glyph: "𓂧",
                    name: "Brain",
                    subtitle: "Neural Classification",
                    description: "Classify files using on-device inference — heuristic, Spotlight, or CoreML backends."
                )

                modelStatusCard
                quickClassifySection
                batchAnalysisSection
                classificationLegend

                if let errorMessage {
                    ErrorRetryView(message: errorMessage) { await detectModel() }
                }
            }
            .padding()
        }
        .background(PantheonTheme.background)
        .task { await detectModel() }
        .sheet(isPresented: $showFilePicker) {
            MultiDocumentPickerView(selectedURLs: Binding(
                get: { [] },
                set: { urls in
                    Task { await classifyFiles(urls) }
                }
            ))
        }
        .sheet(isPresented: $showDirectoryPicker) {
            DocumentPickerView(selectedURL: Binding(
                get: { nil },
                set: { url in
                    if let url { Task { await scanDirectory(url) } }
                }
            ))
        }
    }

    // MARK: - A. Model Status Card

    private var modelStatusCard: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: modelInfo?.iconName ?? "questionmark.circle")
                    .font(.title2)
                    .foregroundStyle(PantheonTheme.gold)
                VStack(alignment: .leading, spacing: 2) {
                    Text("Model Backend")
                        .font(.headline)
                        .foregroundStyle(PantheonTheme.textPrimary)
                    if let info = modelInfo {
                        Text(info.displayName)
                            .font(.subheadline)
                            .foregroundStyle(PantheonTheme.textSecondary)
                    }
                }
                Spacer()
                modelStatusBadge
            }

            if let info = modelInfo {
                HStack(spacing: 16) {
                    StatPill(label: "Backend", value: info.type.capitalized)
                    StatPill(label: "Status", value: info.loaded ? "Loaded" : "Offline")
                    StatPill(label: "Engine", value: info.name)
                }
            }

            Button {
                Task { await detectModel() }
            } label: {
                HStack(spacing: 6) {
                    if isLoadingModel {
                        ProgressView()
                            .progressViewStyle(.circular)
                            .tint(.black)
                            .scaleEffect(0.8)
                    } else {
                        Image(systemName: "arrow.triangle.2.circlepath")
                    }
                    Text("Detect Model")
                }
                .font(.subheadline.bold())
                .frame(maxWidth: .infinity)
                .padding(.vertical, 10)
                .background(PantheonTheme.gold)
                .foregroundStyle(.black)
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }
            .disabled(isLoadingModel)
        }
        .padding()
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var modelStatusBadge: some View {
        Group {
            if isLoadingModel {
                ProgressView()
                    .progressViewStyle(.circular)
                    .tint(PantheonTheme.gold)
            } else if let info = modelInfo {
                Text(info.loaded ? "ACTIVE" : "OFFLINE")
                    .font(.caption2.bold())
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(info.loaded ? PantheonTheme.success.opacity(0.2) : PantheonTheme.error.opacity(0.2))
                    .foregroundStyle(info.loaded ? PantheonTheme.success : PantheonTheme.error)
                    .clipShape(Capsule())
            }
        }
    }

    // MARK: - B. Quick Classify

    private var quickClassifySection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: "doc.text.magnifyingglass")
                    .foregroundStyle(PantheonTheme.gold)
                Text("Quick Classify")
                    .font(.headline)
                    .foregroundStyle(PantheonTheme.textPrimary)
            }

            Button {
                showFilePicker = true
            } label: {
                HStack {
                    Image(systemName: isClassifying ? "progress.indicator" : "folder.badge.plus")
                    Text(isClassifying ? "Classifying..." : "Pick Files")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(PantheonTheme.gold)
                .foregroundStyle(.black)
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .font(.headline)
            }
            .disabled(isClassifying)

            if isClassifying {
                ClassificationSkeleton()
            }

            if !classifyResults.isEmpty && !isClassifying {
                ForEach(classifyResults) { result in
                    ClassificationRow(classification: result)
                }
            }
        }
        .padding()
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - C. Batch Analysis

    private var batchAnalysisSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: "folder.fill.badge.gearshape")
                    .foregroundStyle(PantheonTheme.gold)
                Text("Batch Analysis")
                    .font(.headline)
                    .foregroundStyle(PantheonTheme.textPrimary)
            }

            Button {
                showDirectoryPicker = true
            } label: {
                HStack {
                    Image(systemName: isBatchRunning ? "progress.indicator" : "folder.badge.questionmark")
                    Text(isBatchRunning ? "Scanning..." : "Scan Directory")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(PantheonTheme.gold)
                .foregroundStyle(.black)
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .font(.headline)
            }
            .disabled(isBatchRunning)

            if isBatchRunning {
                BatchSkeleton()
            }

            if let batch = batchResult, !isBatchRunning {
                batchSummaryCard(batch)
                classBreakdownChart(batch)
                batchGroupedResults(batch)
            }
        }
        .padding()
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func batchSummaryCard(_ batch: BatchClassificationResult) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Scan Complete")
                .font(.headline)
                .foregroundStyle(PantheonTheme.gold)

            HStack(spacing: 24) {
                StatPill(label: "Processed", value: "\(batch.filesProcessed)")
                StatPill(label: "Skipped", value: "\(batch.filesSkipped)")
                StatPill(label: "Model", value: batch.modelUsed)
            }
        }
        .padding()
        .background(PantheonTheme.surfaceElevated)
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private func classBreakdownChart(_ batch: BatchClassificationResult) -> some View {
        let grouped = Dictionary(grouping: batch.classifications, by: \.fileClass)
        let sorted = grouped.sorted { $0.value.count > $1.value.count }
        let total = max(batch.filesProcessed, 1)

        return VStack(alignment: .leading, spacing: 8) {
            Text("Class Breakdown")
                .font(.subheadline.bold())
                .foregroundStyle(PantheonTheme.textPrimary)

            ForEach(sorted, id: \.key) { fileClass, items in
                HStack(spacing: 8) {
                    Text(fileClass)
                        .font(.caption)
                        .foregroundStyle(PantheonTheme.textSecondary)
                        .frame(width: 60, alignment: .leading)

                    GeometryReader { geo in
                        RoundedRectangle(cornerRadius: 4)
                            .fill(FileClassColors.color(for: fileClass))
                            .frame(width: geo.size.width * CGFloat(items.count) / CGFloat(total))
                    }
                    .frame(height: 16)

                    Text("\(items.count)")
                        .font(.caption.bold())
                        .foregroundStyle(PantheonTheme.textPrimary)
                        .frame(width: 30, alignment: .trailing)
                }
            }
        }
        .padding()
        .background(PantheonTheme.surfaceElevated)
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private func batchGroupedResults(_ batch: BatchClassificationResult) -> some View {
        let grouped = Dictionary(grouping: batch.classifications, by: \.fileClass)
        let sorted = grouped.sorted { $0.value.count > $1.value.count }

        return VStack(alignment: .leading, spacing: 12) {
            ForEach(sorted, id: \.key) { fileClass, items in
                DisclosureGroup {
                    ForEach(items) { item in
                        ClassificationRow(classification: item)
                    }
                } label: {
                    HStack(spacing: 8) {
                        ClassBadge(fileClass: fileClass)
                        Text("\(items.count) file\(items.count == 1 ? "" : "s")")
                            .font(.subheadline)
                            .foregroundStyle(PantheonTheme.textSecondary)
                    }
                }
                .tint(PantheonTheme.gold)
            }
        }
    }

    // MARK: - D. Classification Legend

    private var classificationLegend: some View {
        DisclosureGroup(isExpanded: $legendExpanded) {
            VStack(alignment: .leading, spacing: 8) {
                ForEach(FileClassInfo.allClasses) { info in
                    HStack(spacing: 10) {
                        ClassBadge(fileClass: info.id)
                        VStack(alignment: .leading, spacing: 2) {
                            Text(info.displayName)
                                .font(.subheadline)
                                .foregroundStyle(PantheonTheme.textPrimary)
                            Text(info.description)
                                .font(.caption)
                                .foregroundStyle(PantheonTheme.textSecondary)
                        }
                    }
                    .padding(.vertical, 2)
                }
            }
            .padding(.top, 8)
        } label: {
            HStack(spacing: 8) {
                Image(systemName: "paintpalette.fill")
                    .foregroundStyle(PantheonTheme.gold)
                Text("Classification Legend")
                    .font(.headline)
                    .foregroundStyle(PantheonTheme.textPrimary)
            }
        }
        .tint(PantheonTheme.gold)
        .padding()
        .background(PantheonTheme.surface)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - Actions

    private func detectModel() async {
        isLoadingModel = true
        errorMessage = nil
        defer { isLoadingModel = false }

        do {
            modelInfo = try appState.bridge.brainModelInfo()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func classifyFiles(_ urls: [URL]) async {
        guard !urls.isEmpty else { return }
        isClassifying = true
        errorMessage = nil
        classifyResults = []
        defer { isClassifying = false }

        // Classify each file individually for granular results
        for url in urls {
            let secured = url.startAccessingSecurityScopedResource()
            defer { if secured { url.stopAccessingSecurityScopedResource() } }

            do {
                let result = try await appState.bridge.brainClassify(filePath: url.path)
                classifyResults.append(result)
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }

    private func scanDirectory(_ url: URL) async {
        let secured = url.startAccessingSecurityScopedResource()
        defer { if secured { url.stopAccessingSecurityScopedResource() } }

        isBatchRunning = true
        errorMessage = nil
        batchResult = nil
        defer { isBatchRunning = false }

        // Enumerate files in the directory
        let fm = FileManager.default
        guard let enumerator = fm.enumerator(at: url, includingPropertiesForKeys: [.isRegularFileKey], options: [.skipsHiddenFiles]) else {
            errorMessage = "Could not enumerate directory"
            return
        }

        var paths: [String] = []
        for case let fileURL as URL in enumerator {
            do {
                let resourceValues = try fileURL.resourceValues(forKeys: [.isRegularFileKey])
                if resourceValues.isRegularFile == true {
                    paths.append(fileURL.path)
                }
            } catch {
                continue
            }
            // Cap at 500 files to avoid UI overload
            if paths.count >= 500 { break }
        }

        guard !paths.isEmpty else {
            errorMessage = "No files found in directory"
            return
        }

        do {
            batchResult = try await appState.bridge.brainClassifyBatch(paths: paths, workers: 4)
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

// MARK: - Classification Row

struct ClassificationRow: View {
    let classification: FileClassification

    var body: some View {
        HStack(spacing: 10) {
            VStack(alignment: .leading, spacing: 4) {
                Text(classification.fileName)
                    .font(.subheadline)
                    .foregroundStyle(PantheonTheme.textPrimary)
                    .lineLimit(1)

                Text(classification.truncatedPath)
                    .font(.caption)
                    .foregroundStyle(PantheonTheme.textSecondary)
                    .lineLimit(1)
            }

            Spacer()

            VStack(alignment: .trailing, spacing: 6) {
                ClassBadge(fileClass: classification.fileClass)
                ConfidenceBar(percent: classification.confidencePercent)
            }
        }
        .padding(10)
        .background(PantheonTheme.surfaceElevated)
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}

// MARK: - Class Badge

struct ClassBadge: View {
    let fileClass: String

    var body: some View {
        Text(fileClass)
            .font(.caption2.bold())
            .textCase(.uppercase)
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(FileClassColors.color(for: fileClass).opacity(0.2))
            .foregroundStyle(FileClassColors.color(for: fileClass))
            .clipShape(Capsule())
    }
}

// MARK: - Confidence Bar

struct ConfidenceBar: View {
    let percent: Int

    var body: some View {
        HStack(spacing: 4) {
            GeometryReader { geo in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 3)
                        .fill(PantheonTheme.surfaceElevated)
                    RoundedRectangle(cornerRadius: 3)
                        .fill(confidenceColor)
                        .frame(width: geo.size.width * CGFloat(percent) / 100)
                }
            }
            .frame(width: 40, height: 6)

            Text("\(percent)%")
                .font(.caption2)
                .foregroundStyle(PantheonTheme.textSecondary)
                .frame(width: 30, alignment: .trailing)
        }
    }

    private var confidenceColor: Color {
        switch percent {
        case 80...100: return PantheonTheme.success
        case 50..<80:  return PantheonTheme.warning
        default:       return PantheonTheme.error
        }
    }
}

// MARK: - Loading Skeletons

struct ClassificationSkeleton: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            ForEach(0..<3, id: \.self) { _ in
                HStack {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("filename.ext")
                            .font(.subheadline)
                        Text(".../path/to/file")
                            .font(.caption)
                    }
                    Spacer()
                    Text("PROJECT")
                        .font(.caption2)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 3)
                        .background(Color.gray.opacity(0.2))
                        .clipShape(Capsule())
                }
                .padding(10)
                .background(PantheonTheme.surfaceElevated)
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }
        }
        .shimmer()
    }
}

struct BatchSkeleton: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 24) {
                StatPill(label: "Processed", value: "42")
                StatPill(label: "Skipped", value: "3")
                StatPill(label: "Model", value: "stub")
            }
            .padding()
            .background(PantheonTheme.surfaceElevated)
            .clipShape(RoundedRectangle(cornerRadius: 8))

            ForEach(0..<4, id: \.self) { _ in
                HStack(spacing: 8) {
                    Text("class")
                        .font(.caption)
                        .frame(width: 60, alignment: .leading)
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.gray.opacity(0.3))
                        .frame(height: 16)
                    Text("12")
                        .font(.caption.bold())
                        .frame(width: 30, alignment: .trailing)
                }
            }
            .padding()
            .background(PantheonTheme.surfaceElevated)
            .clipShape(RoundedRectangle(cornerRadius: 8))
        }
        .shimmer()
    }
}

// MARK: - File Class Colors

enum FileClassColors {
    static func color(for fileClass: String) -> Color {
        switch fileClass {
        case "junk":      return PantheonTheme.error
        case "essential": return PantheonTheme.success
        case "project":   return Color(hex: 0x42A5F5)
        case "model":     return Color(hex: 0xAB47BC)
        case "data":      return Color(hex: 0x26C6DA)
        case "media":     return Color(hex: 0xFFA726)
        case "archive":   return Color(hex: 0x8D6E63)
        case "config":    return PantheonTheme.gold
        default:          return PantheonTheme.textSecondary
        }
    }
}

// MARK: - File Class Info (Legend Data)

struct FileClassInfo: Identifiable {
    let id: String
    let displayName: String
    let description: String

    static let allClasses: [FileClassInfo] = [
        FileClassInfo(id: "junk", displayName: "Junk", description: "Temporary files, caches, build artifacts"),
        FileClassInfo(id: "essential", displayName: "Essential", description: "System or application critical files"),
        FileClassInfo(id: "project", displayName: "Project", description: "Source code, documentation, project files"),
        FileClassInfo(id: "model", displayName: "Model", description: "ML model weights and checkpoints"),
        FileClassInfo(id: "data", displayName: "Data", description: "Datasets, databases, structured data"),
        FileClassInfo(id: "media", displayName: "Media", description: "Images, video, audio files"),
        FileClassInfo(id: "archive", displayName: "Archive", description: "Compressed files and disk images"),
        FileClassInfo(id: "config", displayName: "Config", description: "Configuration and settings files"),
        FileClassInfo(id: "unknown", displayName: "Unknown", description: "Unclassified files"),
    ]
}

// MARK: - Multi-Document Picker (for file selection)

struct MultiDocumentPickerView: UIViewControllerRepresentable {
    @Binding var selectedURLs: [URL]

    func makeUIViewController(context: Context) -> UIDocumentPickerViewController {
        let picker = UIDocumentPickerViewController(forOpeningContentTypes: [.item])
        picker.allowsMultipleSelection = true
        picker.delegate = context.coordinator
        return picker
    }

    func updateUIViewController(_ uiViewController: UIDocumentPickerViewController, context: Context) {}

    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }

    class Coordinator: NSObject, UIDocumentPickerDelegate {
        let parent: MultiDocumentPickerView

        init(_ parent: MultiDocumentPickerView) {
            self.parent = parent
        }

        func documentPicker(_ controller: UIDocumentPickerViewController, didPickDocumentsAt urls: [URL]) {
            parent.selectedURLs = urls
        }
    }
}

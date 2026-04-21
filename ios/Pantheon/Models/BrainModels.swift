import Foundation

// MARK: - Brain (Inference / Classification)

struct FileClassification: Codable, Identifiable {
    var id: String { path }

    let path: String
    let fileClass: String
    let confidence: Double
    let modelUsed: String

    enum CodingKeys: String, CodingKey {
        case path
        case fileClass = "class"
        case confidence
        case modelUsed
    }

    /// Display name for the file (last path component).
    var fileName: String {
        (path as NSString).lastPathComponent
    }

    /// Truncated path for display (keeps last 3 components).
    var truncatedPath: String {
        let components = path.split(separator: "/")
        if components.count <= 3 {
            return path
        }
        return ".../" + components.suffix(3).joined(separator: "/")
    }

    /// Confidence as a percentage integer (0-100).
    var confidencePercent: Int {
        Int(confidence * 100)
    }
}

struct BatchClassificationResult: Codable {
    let classifications: [FileClassification]
    let filesProcessed: Int
    let filesSkipped: Int
    let modelUsed: String
}

struct ModelInfo: Codable {
    let name: String
    let loaded: Bool
    let type: String

    /// Human-readable display name for the backend.
    var displayName: String {
        switch type {
        case "stub":      return "Heuristic (Stub)"
        case "spotlight": return "Spotlight (ANE)"
        case "coreml":    return "CoreML (ANE)"
        case "onnx":      return "ONNX Runtime"
        default:          return name
        }
    }

    /// SF Symbol for the backend type.
    var iconName: String {
        switch type {
        case "stub":      return "cpu"
        case "spotlight":  return "sparkle.magnifyingglass"
        case "coreml":    return "brain"
        case "onnx":      return "square.grid.3x3.middle.filled"
        default:          return "questionmark.circle"
        }
    }
}

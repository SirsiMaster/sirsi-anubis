package ai.sirsi.pantheon.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// --- Brain (Classification) ---

@Serializable
enum class FileClass {
    @SerialName("source") SOURCE,
    @SerialName("test") TEST,
    @SerialName("config") CONFIG,
    @SerialName("docs") DOCS,
    @SerialName("build") BUILD,
    @SerialName("data") DATA,
    @SerialName("generated") GENERATED,
    @SerialName("binary") BINARY,
    @SerialName("media") MEDIA,
    @SerialName("unknown") UNKNOWN,
}

@Serializable
data class FileClassification(
    val path: String,
    @SerialName("file_class") val fileClass: FileClass,
    val confidence: Double,
    val language: String? = null,
    val tokens: Int? = null,
) {
    val displayClass: String
        get() = when (fileClass) {
            FileClass.SOURCE -> "Source"
            FileClass.TEST -> "Test"
            FileClass.CONFIG -> "Config"
            FileClass.DOCS -> "Docs"
            FileClass.BUILD -> "Build"
            FileClass.DATA -> "Data"
            FileClass.GENERATED -> "Generated"
            FileClass.BINARY -> "Binary"
            FileClass.MEDIA -> "Media"
            FileClass.UNKNOWN -> "Unknown"
        }
    val formattedConfidence: String get() = "%.0f%%".format(confidence * 100)
    val id: String get() = path
}

@Serializable
data class BatchResult(
    val results: List<FileClassification>,
    @SerialName("total_files") val totalFiles: Int,
    @SerialName("elapsed_ms") val elapsedMs: Long,
    val errors: List<String>? = null,
)

@Serializable
data class ModelInfo(
    @SerialName("backend_name") val backendName: String,
    val loaded: Boolean,
    val version: String? = null,
)

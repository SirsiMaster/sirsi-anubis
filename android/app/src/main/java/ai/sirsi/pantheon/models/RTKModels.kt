package ai.sirsi.pantheon.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// --- RTK (Output Filter) ---

@Serializable
data class FilterConfig(
    @SerialName("strip_ansi") val stripAnsi: Boolean,
    val dedup: Boolean,
    @SerialName("dedup_window") val dedupWindow: Int,
    @SerialName("max_lines") val maxLines: Int,
    @SerialName("max_bytes") val maxBytes: Int,
    @SerialName("tail_lines") val tailLines: Int,
    @SerialName("collapse_blank") val collapseBlank: Boolean,
)

@Serializable
data class FilterResult(
    val output: String,
    @SerialName("original_bytes") val originalBytes: Int,
    @SerialName("filtered_bytes") val filteredBytes: Int,
    @SerialName("lines_removed") val linesRemoved: Int,
    @SerialName("dups_collapsed") val dupsCollapsed: Int,
    val truncated: Boolean,
    val ratio: Double,
) {
    val reductionPercent: String
        get() = "%.1f%%".format((1.0 - ratio) * 100.0)

    val formattedOriginal: String get() = formatBytes(originalBytes.toLong())
    val formattedFiltered: String get() = formatBytes(filteredBytes.toLong())
}

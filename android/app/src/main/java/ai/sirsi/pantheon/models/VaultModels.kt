package ai.sirsi.pantheon.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// --- Vault (Context Keeper) ---

@Serializable
data class VaultEntry(
    val id: Long,
    val source: String,
    val tag: String,
    val content: String? = null,
    val tokens: Int,
    val createdAt: String? = null,
    val snippet: String? = null,
) {
    val displaySnippet: String
        get() = snippet ?: content?.take(120)?.let { if (content.length > 120) "$it..." else it } ?: ""
    val formattedTokens: String get() = "$tokens tok"
}

@Serializable
data class VaultSearchResult(
    val query: String,
    val totalHits: Int,
    val entries: List<VaultEntry>,
)

@Serializable
data class VaultStoreStats(
    val totalEntries: Int,
    val totalBytes: Long,
    val totalTokens: Long,
    val oldestEntry: String? = null,
    val newestEntry: String? = null,
    val tagCounts: Map<String, Int>? = null,
) {
    val formattedTokens: String get() = "%,d".format(totalTokens)
    val formattedBytes: String get() = formatBytes(totalBytes)
}

@Serializable
data class VaultPruneResult(
    val pruned: Int,
)

package ai.sirsi.pantheon.models

import kotlinx.serialization.Serializable

// --- Horus (Code Graph) ---

@Serializable
data class HorusSymbol(
    val name: String,
    val kind: String,
    val file: String,
    val line: Int,
    val endLine: Int = 0,
    val signature: String = "",
    val doc: String? = null,
    val exported: Boolean = false,
    val parent: String? = null,
) {
    val id: String get() = "$file:$name:$line"

    val kindIcon: String
        get() = when (kind) {
            "func"      -> "f()"
            "method"    -> "m()"
            "type"      -> "T"
            "struct"    -> "S"
            "interface" -> "I"
            "const"     -> "C"
            "var"       -> "V"
            "package"   -> "P"
            else        -> "?"
        }

    val displayKind: String
        get() = when (kind) {
            "func"      -> "Function"
            "method"    -> "Method"
            "type"      -> "Type"
            "struct"    -> "Struct"
            "interface" -> "Interface"
            "const"     -> "Constant"
            "var"       -> "Variable"
            "package"   -> "Package"
            else        -> kind
        }
}

@Serializable
data class HorusGraphStats(
    val files: Int,
    val packages: Int,
    val types: Int,
    val functions: Int,
    val methods: Int,
    val interfaces: Int,
    val totalLines: Int,
)

@Serializable
data class HorusSymbolGraph(
    val root: String,
    val packages: List<String>,
    val symbols: List<HorusSymbol>,
    val stats: HorusGraphStats,
    val builtAt: String,
)

@Serializable
data class HorusOutlineResult(
    val outline: String,
)

@Serializable
data class HorusContextResult(
    val context: String,
)

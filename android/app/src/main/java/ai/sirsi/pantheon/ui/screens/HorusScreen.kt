package ai.sirsi.pantheon.ui.screens

import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.launch
import ai.sirsi.pantheon.R
import ai.sirsi.pantheon.bridge.PantheonBridge
import ai.sirsi.pantheon.models.HorusSymbol
import ai.sirsi.pantheon.models.HorusSymbolGraph
import ai.sirsi.pantheon.ui.components.DeityHeader
import ai.sirsi.pantheon.ui.components.StatPill
import ai.sirsi.pantheon.ui.theme.PantheonError
import ai.sirsi.pantheon.ui.theme.PantheonGold
import ai.sirsi.pantheon.ui.theme.PantheonLapis
import ai.sirsi.pantheon.ui.theme.PantheonSurface
import ai.sirsi.pantheon.ui.theme.PantheonTextSecondary

/**
 * Horus code graph screen. Parse a project directory to extract
 * structural symbols (types, functions, methods, interfaces) and
 * browse the resulting symbol graph.
 */
@Composable
fun HorusScreen() {
    var projectRoot by remember { mutableStateOf("") }
    var isParsing by remember { mutableStateOf(false) }
    var graph by remember { mutableStateOf<HorusSymbolGraph?>(null) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    var filterText by remember { mutableStateOf("") }
    var selectedKind by remember { mutableStateOf<String?>(null) }
    var selectedSymbolContext by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 16.dp, vertical = 24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        item {
            DeityHeader(
                glyph = "\uD80C\uDC80",
                name = stringResource(R.string.horus_name),
                subtitle = stringResource(R.string.horus_subtitle),
                description = stringResource(R.string.horus_description),
            )
        }

        // Project root input
        item {
            OutlinedTextField(
                value = projectRoot,
                onValueChange = { projectRoot = it },
                label = { Text(stringResource(R.string.horus_root_hint)) },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = PantheonGold,
                    unfocusedBorderColor = PantheonTextSecondary,
                    cursorColor = PantheonGold,
                    focusedLabelColor = PantheonGold,
                ),
            )
        }

        // Parse button
        item {
            Button(
                onClick = {
                    scope.launch {
                        isParsing = true
                        errorMessage = null
                        selectedSymbolContext = null
                        try {
                            graph = PantheonBridge.horusParseDir(projectRoot)
                        } catch (e: Exception) {
                            errorMessage = e.message ?: "Parse failed"
                        } finally {
                            isParsing = false
                        }
                    }
                },
                enabled = !isParsing && projectRoot.isNotBlank(),
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.buttonColors(containerColor = PantheonGold),
                shape = RoundedCornerShape(8.dp),
            ) {
                if (isParsing) {
                    CircularProgressIndicator(
                        color = MaterialTheme.colorScheme.onPrimary,
                        strokeWidth = 2.dp,
                    )
                } else {
                    Text(stringResource(R.string.action_parse))
                }
            }
        }

        // Error
        if (errorMessage != null) {
            item {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = PantheonError.copy(alpha = 0.1f)),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    Text(
                        text = errorMessage!!,
                        modifier = Modifier.padding(16.dp),
                        color = PantheonError,
                        style = MaterialTheme.typography.bodyMedium,
                    )
                }
            }
        }

        // Graph stats
        if (graph != null) {
            item {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = PantheonSurface),
                    shape = RoundedCornerShape(12.dp),
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = stringResource(R.string.horus_stats_title),
                            style = MaterialTheme.typography.titleMedium,
                            color = PantheonGold,
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .horizontalScroll(rememberScrollState()),
                            horizontalArrangement = Arrangement.SpaceEvenly,
                        ) {
                            StatPill(label = stringResource(R.string.horus_files), value = "${graph!!.stats.files}")
                            StatPill(label = stringResource(R.string.horus_packages), value = "${graph!!.stats.packages}")
                            StatPill(label = stringResource(R.string.horus_types), value = "${graph!!.stats.types}")
                            StatPill(label = stringResource(R.string.horus_functions), value = "${graph!!.stats.functions}")
                            StatPill(label = stringResource(R.string.horus_methods), value = "${graph!!.stats.methods}")
                            StatPill(label = stringResource(R.string.horus_interfaces), value = "${graph!!.stats.interfaces}")
                            StatPill(label = stringResource(R.string.horus_lines), value = "${graph!!.stats.totalLines}")
                        }
                    }
                }
            }

            // Filter chips — string-based kind filtering
            item {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    FilterChip(
                        selected = selectedKind == null,
                        onClick = { selectedKind = null },
                        label = { Text(stringResource(R.string.horus_filter_all)) },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = PantheonGold,
                            selectedLabelColor = PantheonSurface,
                        ),
                    )
                    listOf(
                        "type" to R.string.horus_filter_types,
                        "func" to R.string.horus_filter_functions,
                        "method" to R.string.horus_filter_methods,
                        "interface" to R.string.horus_filter_interfaces,
                    ).forEach { (kind, labelRes) ->
                        FilterChip(
                            selected = selectedKind == kind,
                            onClick = { selectedKind = if (selectedKind == kind) null else kind },
                            label = { Text(stringResource(labelRes)) },
                            colors = FilterChipDefaults.filterChipColors(
                                selectedContainerColor = PantheonGold,
                                selectedLabelColor = PantheonSurface,
                            ),
                        )
                    }
                }
            }

            // Search filter
            item {
                OutlinedTextField(
                    value = filterText,
                    onValueChange = { filterText = it },
                    label = { Text(stringResource(R.string.horus_search_hint)) },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = PantheonGold,
                        unfocusedBorderColor = PantheonTextSecondary,
                        cursorColor = PantheonGold,
                        focusedLabelColor = PantheonGold,
                    ),
                )
            }
        }

        // Context detail (shown when a symbol is tapped)
        if (selectedSymbolContext != null) {
            item {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = PantheonLapis.copy(alpha = 0.3f)),
                    shape = RoundedCornerShape(12.dp),
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = stringResource(R.string.horus_context_title),
                            style = MaterialTheme.typography.titleMedium,
                            color = PantheonGold,
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = selectedSymbolContext!!,
                            style = MaterialTheme.typography.bodySmall.copy(
                                fontFamily = FontFamily.Monospace,
                            ),
                            color = PantheonTextSecondary,
                        )
                    }
                }
            }
        }

        // Filtered symbol list
        val filteredSymbols = graph?.symbols.orEmpty().filter { symbol ->
            val kindMatch = selectedKind == null || symbol.kind == selectedKind
            val textMatch = filterText.isBlank() || symbol.name.contains(filterText, ignoreCase = true)
            kindMatch && textMatch
        }

        items(filteredSymbols, key = { it.id }) { symbol ->
            SymbolRow(
                symbol = symbol,
                onClick = {
                    scope.launch {
                        try {
                            val result = PantheonBridge.horusContextFor(
                                graph!!.root,
                                symbol.name,
                            )
                            selectedSymbolContext = result.context
                        } catch (e: Exception) {
                            errorMessage = e.message ?: "Failed to load context"
                        }
                    }
                },
            )
        }
    }
}

@Composable
private fun SymbolRow(
    symbol: HorusSymbol,
    onClick: () -> Unit,
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 1.dp),
        colors = CardDefaults.cardColors(containerColor = PantheonSurface),
        shape = RoundedCornerShape(8.dp),
        onClick = onClick,
    ) {
        Row(
            modifier = Modifier.padding(12.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = if (symbol.parent != null) "${symbol.parent}.${symbol.name}" else symbol.name,
                    style = MaterialTheme.typography.bodyLarge,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Text(
                        text = symbol.displayKind,
                        style = MaterialTheme.typography.bodySmall,
                        color = PantheonGold,
                    )
                    if (symbol.signature.isNotEmpty()) {
                        Text(
                            text = symbol.signature,
                            style = MaterialTheme.typography.bodySmall,
                            color = PantheonTextSecondary,
                            maxLines = 1,
                        )
                    }
                }
                Text(
                    text = "${symbol.file}:${symbol.line}",
                    style = MaterialTheme.typography.labelSmall,
                    color = PantheonTextSecondary,
                )
            }
            if (symbol.exported) {
                Text(
                    text = "pub",
                    style = MaterialTheme.typography.labelSmall,
                    color = PantheonGold,
                )
            }
        }
    }
}

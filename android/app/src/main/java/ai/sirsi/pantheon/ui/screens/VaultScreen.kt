package ai.sirsi.pantheon.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.launch
import ai.sirsi.pantheon.R
import ai.sirsi.pantheon.bridge.PantheonBridge
import ai.sirsi.pantheon.models.VaultEntry
import ai.sirsi.pantheon.models.VaultSearchResult
import ai.sirsi.pantheon.models.VaultStoreStats
import ai.sirsi.pantheon.ui.components.DeityHeader
import ai.sirsi.pantheon.ui.components.StatPill
import ai.sirsi.pantheon.ui.theme.PantheonError
import ai.sirsi.pantheon.ui.theme.PantheonGold
import ai.sirsi.pantheon.ui.theme.PantheonSurface
import ai.sirsi.pantheon.ui.theme.PantheonTextSecondary

/**
 * Vault context storage screen. Search, browse, and store
 * indexed code and output in the SQLite FTS5 vault.
 */
@Composable
fun VaultScreen() {
    var searchQuery by remember { mutableStateOf("") }
    var isSearching by remember { mutableStateOf(false) }
    var searchResult by remember { mutableStateOf<VaultSearchResult?>(null) }
    var stats by remember { mutableStateOf<VaultStoreStats?>(null) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    var showStoreDialog by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()

    // Load stats on mount
    LaunchedEffect(Unit) {
        try {
            stats = PantheonBridge.vaultStats()
        } catch (_: Exception) {
            // Stats are optional; vault may not be initialized yet
        }
    }

    Scaffold(
        floatingActionButton = {
            FloatingActionButton(
                onClick = { showStoreDialog = true },
                containerColor = PantheonGold,
            ) {
                Text("+", style = MaterialTheme.typography.headlineSmall)
            }
        },
        containerColor = androidx.compose.ui.graphics.Color.Transparent,
    ) { innerPadding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
                .padding(horizontal = 16.dp, vertical = 24.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            item {
                DeityHeader(
                    glyph = "\uD83C\uDFDB",
                    name = stringResource(R.string.vault_name),
                    subtitle = stringResource(R.string.vault_subtitle),
                    description = stringResource(R.string.vault_description),
                )
            }

            // Stats card
            if (stats != null) {
                item {
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        colors = CardDefaults.cardColors(containerColor = PantheonSurface),
                        shape = RoundedCornerShape(12.dp),
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(
                                text = stringResource(R.string.vault_stats_title),
                                style = MaterialTheme.typography.titleMedium,
                                color = PantheonGold,
                            )
                            Spacer(modifier = Modifier.height(12.dp))
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.SpaceEvenly,
                            ) {
                                StatPill(
                                    label = stringResource(R.string.vault_entries),
                                    value = "${stats!!.totalEntries}",
                                )
                                StatPill(
                                    label = stringResource(R.string.vault_tokens),
                                    value = stats!!.formattedTokens,
                                )
                                StatPill(
                                    label = stringResource(R.string.vault_bytes),
                                    value = stats!!.formattedBytes,
                                )
                            }
                        }
                    }
                }
            }

            // Search bar
            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    OutlinedTextField(
                        value = searchQuery,
                        onValueChange = { searchQuery = it },
                        label = { Text(stringResource(R.string.vault_search_hint)) },
                        modifier = Modifier.weight(1f),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = PantheonGold,
                            unfocusedBorderColor = PantheonTextSecondary,
                            cursorColor = PantheonGold,
                            focusedLabelColor = PantheonGold,
                        ),
                        singleLine = true,
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Button(
                        onClick = {
                            scope.launch {
                                isSearching = true
                                errorMessage = null
                                try {
                                    searchResult = PantheonBridge.vaultSearch(searchQuery)
                                } catch (e: Exception) {
                                    errorMessage = e.message ?: "Search failed"
                                } finally {
                                    isSearching = false
                                }
                            }
                        },
                        enabled = !isSearching && searchQuery.isNotBlank(),
                        colors = ButtonDefaults.buttonColors(containerColor = PantheonGold),
                        shape = RoundedCornerShape(8.dp),
                    ) {
                        if (isSearching) {
                            CircularProgressIndicator(
                                color = MaterialTheme.colorScheme.onPrimary,
                                strokeWidth = 2.dp,
                            )
                        } else {
                            Text(stringResource(R.string.action_search))
                        }
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

            // Results
            val entries = searchResult?.entries.orEmpty()
            if (entries.isEmpty() && searchResult != null) {
                item {
                    Text(
                        text = stringResource(R.string.vault_no_results),
                        style = MaterialTheme.typography.bodyMedium,
                        color = PantheonTextSecondary,
                        modifier = Modifier.padding(vertical = 8.dp),
                    )
                }
            }

            items(entries, key = { it.id }) { entry ->
                VaultEntryRow(entry)
            }
        }
    }

    // Store dialog
    if (showStoreDialog) {
        VaultStoreDialog(
            onDismiss = { showStoreDialog = false },
            onStore = { source, tag, content, tokens ->
                scope.launch {
                    try {
                        PantheonBridge.vaultStore(source, tag, content, tokens)
                        showStoreDialog = false
                        // Refresh stats
                        stats = PantheonBridge.vaultStats()
                    } catch (e: Exception) {
                        errorMessage = e.message ?: "Store failed"
                    }
                }
            },
        )
    }
}

@Composable
private fun VaultEntryRow(entry: VaultEntry) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = PantheonSurface),
        shape = RoundedCornerShape(8.dp),
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.Top,
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = entry.source,
                        style = MaterialTheme.typography.bodyLarge,
                    )
                    Text(
                        text = entry.tag,
                        style = MaterialTheme.typography.bodySmall,
                        color = PantheonGold,
                    )
                }
                Text(
                    text = entry.formattedTokens,
                    style = MaterialTheme.typography.titleSmall,
                    color = PantheonGold,
                )
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = entry.displaySnippet,
                style = MaterialTheme.typography.bodySmall,
                color = PantheonTextSecondary,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = entry.createdAt ?: "",
                style = MaterialTheme.typography.labelSmall,
                color = PantheonTextSecondary,
            )
        }
    }
}

@Composable
private fun VaultStoreDialog(
    onDismiss: () -> Unit,
    onStore: (source: String, tag: String, content: String, tokens: Int) -> Unit,
) {
    var source by remember { mutableStateOf("") }
    var tag by remember { mutableStateOf("") }
    var content by remember { mutableStateOf("") }
    var tokens by remember { mutableStateOf("0") }

    androidx.compose.material3.AlertDialog(
        onDismissRequest = onDismiss,
        confirmButton = {
            Button(
                onClick = { onStore(source, tag, content, tokens.toIntOrNull() ?: 0) },
                enabled = source.isNotBlank() && content.isNotBlank(),
                colors = ButtonDefaults.buttonColors(containerColor = PantheonGold),
            ) {
                Text(stringResource(R.string.vault_action_store))
            }
        },
        dismissButton = {
            Button(
                onClick = onDismiss,
                colors = ButtonDefaults.buttonColors(containerColor = PantheonSurface),
            ) {
                Text(stringResource(R.string.action_cancel))
            }
        },
        title = { Text(stringResource(R.string.vault_store_title), color = PantheonGold) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    value = source,
                    onValueChange = { source = it },
                    label = { Text(stringResource(R.string.vault_field_source)) },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = PantheonGold,
                        cursorColor = PantheonGold,
                        focusedLabelColor = PantheonGold,
                    ),
                )
                OutlinedTextField(
                    value = tag,
                    onValueChange = { tag = it },
                    label = { Text(stringResource(R.string.vault_field_tag)) },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = PantheonGold,
                        cursorColor = PantheonGold,
                        focusedLabelColor = PantheonGold,
                    ),
                )
                OutlinedTextField(
                    value = content,
                    onValueChange = { content = it },
                    label = { Text(stringResource(R.string.vault_field_content)) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(120.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = PantheonGold,
                        cursorColor = PantheonGold,
                        focusedLabelColor = PantheonGold,
                    ),
                )
                OutlinedTextField(
                    value = tokens,
                    onValueChange = { tokens = it },
                    label = { Text(stringResource(R.string.vault_field_tokens)) },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = PantheonGold,
                        cursorColor = PantheonGold,
                        focusedLabelColor = PantheonGold,
                    ),
                )
            }
        },
        containerColor = PantheonSurface,
    )
}

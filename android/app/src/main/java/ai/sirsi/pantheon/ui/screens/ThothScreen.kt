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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
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
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.TabRowDefaults
import androidx.compose.material3.TabRowDefaults.tabIndicatorOffset
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import ai.sirsi.pantheon.R
import ai.sirsi.pantheon.bridge.PantheonBridge
import ai.sirsi.pantheon.models.FilterResult
import ai.sirsi.pantheon.models.ProjectInfo
import ai.sirsi.pantheon.models.VaultEntry
import ai.sirsi.pantheon.models.VaultSearchResult
import ai.sirsi.pantheon.models.VaultStoreStats
import ai.sirsi.pantheon.ui.components.DeityHeader
import ai.sirsi.pantheon.ui.components.StatPill
import ai.sirsi.pantheon.ui.theme.PantheonBlack
import ai.sirsi.pantheon.ui.theme.PantheonError
import ai.sirsi.pantheon.ui.theme.PantheonGold
import ai.sirsi.pantheon.ui.theme.PantheonSurface
import ai.sirsi.pantheon.ui.theme.PantheonTextSecondary

/**
 * Thoth screen with tabbed interface: Memory, Filter, and Vault.
 * RTK and Vault content are folded in as tools serving Thoth's context optimization mission.
 */
@Composable
fun ThothScreen() {
    var selectedTab by remember { mutableIntStateOf(0) }
    val tabs = listOf("Memory", "Filter", "Vault")

    Column(
        modifier = Modifier.fillMaxSize(),
    ) {
        // Header (always visible above tabs)
        Column(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 16.dp),
        ) {
            DeityHeader(
                glyph = "\uD80C\uDC5F",
                name = stringResource(R.string.thoth_name),
                subtitle = stringResource(R.string.thoth_subtitle),
                description = stringResource(R.string.thoth_description),
            )
        }

        // Tab row
        TabRow(
            selectedTabIndex = selectedTab,
            containerColor = PantheonBlack,
            contentColor = PantheonGold,
            indicator = { tabPositions ->
                if (selectedTab < tabPositions.size) {
                    TabRowDefaults.SecondaryIndicator(
                        modifier = Modifier.tabIndicatorOffset(tabPositions[selectedTab]),
                        color = PantheonGold,
                    )
                }
            },
        ) {
            tabs.forEachIndexed { index, title ->
                Tab(
                    selected = selectedTab == index,
                    onClick = { selectedTab = index },
                    text = {
                        Text(
                            text = title,
                            color = if (selectedTab == index) PantheonGold else PantheonTextSecondary,
                        )
                    },
                )
            }
        }

        // Tab content
        when (selectedTab) {
            0 -> ThothMemoryTab()
            1 -> ThothFilterTab()
            2 -> ThothVaultTab()
        }
    }
}

// MARK: - Memory Tab

@Composable
private fun ThothMemoryTab() {
    var isSyncing by remember { mutableStateOf(false) }
    var projectInfo by remember { mutableStateOf<ProjectInfo?>(null) }
    var syncStatus by remember { mutableStateOf<String?>(null) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()
    val fallbackError = stringResource(R.string.error_generic)

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = 16.dp, vertical = 16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Button(
            onClick = {
                scope.launch {
                    isSyncing = true
                    errorMessage = null
                    try {
                        val dataDir = PantheonBridge.version() // verify bridge
                        val root = android.os.Environment.getExternalStorageDirectory().absolutePath
                        projectInfo = PantheonBridge.thothDetectProject(root)
                        PantheonBridge.thothSync(root)
                        syncStatus = "Synced"
                    } catch (e: Exception) {
                        errorMessage = e.message ?: fallbackError
                    } finally {
                        isSyncing = false
                    }
                }
            },
            enabled = !isSyncing,
            modifier = Modifier.fillMaxWidth(),
            colors = ButtonDefaults.buttonColors(containerColor = PantheonGold),
            shape = RoundedCornerShape(8.dp),
        ) {
            if (isSyncing) {
                CircularProgressIndicator(
                    color = MaterialTheme.colorScheme.onPrimary,
                    strokeWidth = 2.dp,
                )
            } else {
                Text(stringResource(R.string.action_sync))
            }
        }

        // Error
        if (errorMessage != null) {
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

        // Project Info
        if (projectInfo != null) {
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = PantheonSurface),
                shape = RoundedCornerShape(12.dp),
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text(
                        text = "Project Info",
                        style = MaterialTheme.typography.titleMedium,
                        color = PantheonGold,
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    InfoRow("Name", projectInfo?.name ?: "Unknown")
                    InfoRow("Language", projectInfo?.language ?: "Unknown")
                    InfoRow("Version", projectInfo?.version ?: "Unknown")
                    InfoRow("Root", projectInfo?.root ?: "N/A")
                }
            }
        }

        // Sync status
        if (syncStatus != null) {
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = PantheonSurface),
                shape = RoundedCornerShape(12.dp),
            ) {
                Text(
                    text = "Status: $syncStatus",
                    modifier = Modifier.padding(16.dp),
                    style = MaterialTheme.typography.bodyLarge,
                    color = PantheonGold,
                )
            }
        }
    }
}

// MARK: - Filter Tab (formerly RTKScreen)

@Composable
private fun ThothFilterTab() {
    var rawInput by remember { mutableStateOf("") }
    var isFiltering by remember { mutableStateOf(false) }
    var filterResult by remember { mutableStateOf<FilterResult?>(null) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 16.dp, vertical = 16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        // Input area
        item {
            OutlinedTextField(
                value = rawInput,
                onValueChange = { rawInput = it },
                label = { Text(stringResource(R.string.rtk_input_hint)) },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(200.dp),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = PantheonGold,
                    unfocusedBorderColor = PantheonTextSecondary,
                    cursorColor = PantheonGold,
                    focusedLabelColor = PantheonGold,
                ),
                textStyle = MaterialTheme.typography.bodySmall.copy(fontFamily = FontFamily.Monospace),
                maxLines = 20,
            )
        }

        // Filter button
        item {
            Button(
                onClick = {
                    scope.launch {
                        isFiltering = true
                        errorMessage = null
                        try {
                            filterResult = withContext(Dispatchers.IO) {
                                PantheonBridge.rtkFilter(rawInput)
                            }
                        } catch (e: Exception) {
                            errorMessage = e.message ?: "Filter failed"
                        } finally {
                            isFiltering = false
                        }
                    }
                },
                enabled = !isFiltering && rawInput.isNotBlank(),
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.buttonColors(containerColor = PantheonGold),
                shape = RoundedCornerShape(8.dp),
            ) {
                if (isFiltering) {
                    CircularProgressIndicator(
                        color = MaterialTheme.colorScheme.onPrimary,
                        strokeWidth = 2.dp,
                    )
                } else {
                    Text(stringResource(R.string.action_filter))
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

        // Stats
        if (filterResult != null) {
            item {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = PantheonSurface),
                    shape = RoundedCornerShape(12.dp),
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = stringResource(R.string.rtk_stats_title),
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
                            StatPill(
                                label = stringResource(R.string.rtk_original),
                                value = filterResult!!.formattedOriginal,
                            )
                            StatPill(
                                label = stringResource(R.string.rtk_filtered),
                                value = filterResult!!.formattedFiltered,
                            )
                            StatPill(
                                label = stringResource(R.string.rtk_ratio),
                                value = filterResult!!.reductionPercent,
                            )
                            StatPill(
                                label = stringResource(R.string.rtk_lines_removed),
                                value = "${filterResult!!.linesRemoved}",
                            )
                            StatPill(
                                label = stringResource(R.string.rtk_dupes),
                                value = "${filterResult!!.dupsCollapsed}",
                            )
                        }
                    }
                }
            }

            // Filtered output
            item {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = PantheonSurface),
                    shape = RoundedCornerShape(12.dp),
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = stringResource(R.string.rtk_output_title),
                            style = MaterialTheme.typography.titleMedium,
                            color = PantheonGold,
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = filterResult!!.output,
                            style = MaterialTheme.typography.bodySmall.copy(
                                fontFamily = FontFamily.Monospace,
                            ),
                            color = PantheonTextSecondary,
                        )
                    }
                }
            }
        }
    }
}

// MARK: - Vault Tab (formerly VaultScreen)

@Composable
private fun ThothVaultTab() {
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
                .padding(horizontal = 16.dp, vertical = 16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
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

// MARK: - Shared subviews

@Composable
internal fun InfoRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.bodyMedium,
            color = PantheonTextSecondary,
            modifier = Modifier.weight(0.35f),
        )
        Text(
            text = value,
            style = MaterialTheme.typography.bodyMedium,
            modifier = Modifier.weight(0.65f),
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

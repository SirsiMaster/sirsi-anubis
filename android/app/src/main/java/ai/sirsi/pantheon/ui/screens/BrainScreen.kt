package ai.sirsi.pantheon.ui.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
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
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.launch
import ai.sirsi.pantheon.R
import ai.sirsi.pantheon.bridge.PantheonBridge
import ai.sirsi.pantheon.models.BatchResult
import ai.sirsi.pantheon.models.FileClass
import ai.sirsi.pantheon.models.FileClassification
import ai.sirsi.pantheon.models.ModelInfo
import ai.sirsi.pantheon.ui.components.DeityHeader
import ai.sirsi.pantheon.ui.components.StatPill
import ai.sirsi.pantheon.ui.theme.PantheonError
import ai.sirsi.pantheon.ui.theme.PantheonGold
import ai.sirsi.pantheon.ui.theme.PantheonLapis
import ai.sirsi.pantheon.ui.theme.PantheonSuccess
import ai.sirsi.pantheon.ui.theme.PantheonSurface
import ai.sirsi.pantheon.ui.theme.PantheonTextSecondary
import androidx.compose.ui.graphics.Color

/**
 * Brain file classification screen. Classify individual files or
 * batch-scan directories using the Brain model.
 */
@Composable
fun BrainScreen() {
    var filePath by remember { mutableStateOf("") }
    var batchDir by remember { mutableStateOf("") }
    var isClassifying by remember { mutableStateOf(false) }
    var isBatchRunning by remember { mutableStateOf(false) }
    var singleResult by remember { mutableStateOf<FileClassification?>(null) }
    var batchResult by remember { mutableStateOf<BatchResult?>(null) }
    var modelInfo by remember { mutableStateOf<ModelInfo?>(null) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    // Load model info on mount
    LaunchedEffect(Unit) {
        try {
            modelInfo = PantheonBridge.brainModelInfo()
        } catch (_: Exception) {
            // Model info may not be available
        }
    }

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 16.dp, vertical = 24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        item {
            DeityHeader(
                glyph = "\uD83E\uDDE0",
                name = stringResource(R.string.brain_name),
                subtitle = stringResource(R.string.brain_subtitle),
                description = stringResource(R.string.brain_description),
            )
        }

        // Model info card
        if (modelInfo != null) {
            item {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = PantheonSurface),
                    shape = RoundedCornerShape(12.dp),
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = stringResource(R.string.brain_model_title),
                            style = MaterialTheme.typography.titleMedium,
                            color = PantheonGold,
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        InfoRow(
                            stringResource(R.string.brain_backend),
                            modelInfo!!.backendName,
                        )
                        InfoRow(
                            stringResource(R.string.brain_status),
                            if (modelInfo!!.loaded) "Loaded" else "Not loaded",
                        )
                        modelInfo!!.version?.let {
                            InfoRow(stringResource(R.string.label_version), it)
                        }
                    }
                }
            }
        }

        // Single file classification
        item {
            Text(
                text = stringResource(R.string.brain_classify_title),
                style = MaterialTheme.typography.titleMedium,
                color = PantheonGold,
            )
        }

        item {
            OutlinedTextField(
                value = filePath,
                onValueChange = { filePath = it },
                label = { Text(stringResource(R.string.brain_file_hint)) },
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

        item {
            Button(
                onClick = {
                    scope.launch {
                        isClassifying = true
                        errorMessage = null
                        try {
                            singleResult = PantheonBridge.brainClassify(filePath)
                        } catch (e: Exception) {
                            errorMessage = e.message ?: "Classification failed"
                        } finally {
                            isClassifying = false
                        }
                    }
                },
                enabled = !isClassifying && filePath.isNotBlank(),
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.buttonColors(containerColor = PantheonGold),
                shape = RoundedCornerShape(8.dp),
            ) {
                if (isClassifying) {
                    CircularProgressIndicator(
                        color = MaterialTheme.colorScheme.onPrimary,
                        strokeWidth = 2.dp,
                    )
                } else {
                    Text(stringResource(R.string.action_classify))
                }
            }
        }

        // Single result
        if (singleResult != null) {
            item {
                ClassificationRow(singleResult!!)
            }
        }

        // Batch mode
        item {
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = stringResource(R.string.brain_batch_title),
                style = MaterialTheme.typography.titleMedium,
                color = PantheonGold,
            )
        }

        item {
            OutlinedTextField(
                value = batchDir,
                onValueChange = { batchDir = it },
                label = { Text(stringResource(R.string.brain_dir_hint)) },
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

        item {
            Button(
                onClick = {
                    scope.launch {
                        isBatchRunning = true
                        errorMessage = null
                        try {
                            batchResult = PantheonBridge.brainClassifyBatch(listOf(batchDir))
                        } catch (e: Exception) {
                            errorMessage = e.message ?: "Batch classification failed"
                        } finally {
                            isBatchRunning = false
                        }
                    }
                },
                enabled = !isBatchRunning && batchDir.isNotBlank(),
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.buttonColors(containerColor = PantheonGold),
                shape = RoundedCornerShape(8.dp),
            ) {
                if (isBatchRunning) {
                    CircularProgressIndicator(
                        color = MaterialTheme.colorScheme.onPrimary,
                        strokeWidth = 2.dp,
                    )
                } else {
                    Text(stringResource(R.string.action_batch_classify))
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

        // Batch stats
        if (batchResult != null) {
            item {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = PantheonSurface),
                    shape = RoundedCornerShape(12.dp),
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = stringResource(R.string.brain_batch_stats),
                            style = MaterialTheme.typography.titleMedium,
                            color = PantheonGold,
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceEvenly,
                        ) {
                            StatPill(
                                label = stringResource(R.string.brain_total_files),
                                value = "${batchResult!!.totalFiles}",
                            )
                            StatPill(
                                label = stringResource(R.string.brain_elapsed),
                                value = "${batchResult!!.elapsedMs}ms",
                            )
                        }
                    }
                }
            }

            // Batch results list
            items(batchResult!!.results, key = { it.id }) { classification ->
                ClassificationRow(classification)
            }
        }
    }
}

@Composable
private fun ClassificationRow(classification: FileClassification) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = PantheonSurface),
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier.padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = classification.path.substringAfterLast('/'),
                    style = MaterialTheme.typography.bodyLarge,
                )
                Text(
                    text = classification.path,
                    style = MaterialTheme.typography.labelSmall,
                    color = PantheonTextSecondary,
                    maxLines = 1,
                )
                classification.language?.let {
                    Text(
                        text = it,
                        style = MaterialTheme.typography.bodySmall,
                        color = PantheonTextSecondary,
                    )
                }
            }
            Spacer(modifier = Modifier.width(8.dp))
            Column(horizontalAlignment = Alignment.End) {
                ClassBadge(classification.fileClass, classification.displayClass)
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = classification.formattedConfidence,
                    style = MaterialTheme.typography.labelSmall,
                    color = PantheonTextSecondary,
                )
            }
        }
    }
}

@Composable
private fun ClassBadge(fileClass: FileClass, label: String) {
    val badgeColor = when (fileClass) {
        FileClass.SOURCE -> PantheonGold
        FileClass.TEST -> PantheonSuccess
        FileClass.CONFIG -> PantheonLapis
        FileClass.DOCS -> Color(0xFF42A5F5)
        FileClass.BUILD -> Color(0xFFFF9800)
        FileClass.DATA -> Color(0xFF9C27B0)
        FileClass.GENERATED -> PantheonTextSecondary
        FileClass.BINARY -> Color(0xFF795548)
        FileClass.MEDIA -> Color(0xFFE91E63)
        FileClass.UNKNOWN -> PantheonTextSecondary
    }
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(4.dp))
            .background(badgeColor.copy(alpha = 0.2f))
            .padding(horizontal = 8.dp, vertical = 2.dp),
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = badgeColor,
        )
    }
}

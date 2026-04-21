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
import androidx.compose.foundation.rememberScrollState
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
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import ai.sirsi.pantheon.R
import ai.sirsi.pantheon.bridge.PantheonBridge
import ai.sirsi.pantheon.models.FilterResult
import ai.sirsi.pantheon.ui.components.DeityHeader
import ai.sirsi.pantheon.ui.components.StatPill
import ai.sirsi.pantheon.ui.theme.PantheonError
import ai.sirsi.pantheon.ui.theme.PantheonGold
import ai.sirsi.pantheon.ui.theme.PantheonSurface
import ai.sirsi.pantheon.ui.theme.PantheonTextSecondary

/**
 * RTK output filter screen. Paste raw tool output and filter it down
 * to the essential signal, stripping noise, duplicates, and ANSI codes.
 */
@Composable
fun RTKScreen() {
    var rawInput by remember { mutableStateOf("") }
    var isFiltering by remember { mutableStateOf(false) }
    var filterResult by remember { mutableStateOf<FilterResult?>(null) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 16.dp, vertical = 24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        item {
            DeityHeader(
                glyph = "\u26A1",
                name = stringResource(R.string.rtk_name),
                subtitle = stringResource(R.string.rtk_subtitle),
                description = stringResource(R.string.rtk_description),
            )
        }

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

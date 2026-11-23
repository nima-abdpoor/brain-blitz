package com.brainblitz.android

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.tween
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.selection.selectableGroup
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.Icon
import androidx.compose.material3.LocalContentColor
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedIconButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import kotlin.math.PI
import kotlin.math.sin

@Composable
internal fun PlayPage(
    screen: AppScreen.Play,
    onJoin: (MatchSelection) -> Unit,
    onRetry: () -> Unit,
    onCancel: () -> Unit,
    onBack: () -> Unit,
) {
    var categoryChoice by rememberSaveable { mutableStateOf(screen.previousSelection?.category) }
    var timerChoice by rememberSaveable { mutableStateOf(screen.previousSelection?.timeOption?.id) }
    var playerChoice by rememberSaveable { mutableStateOf(screen.previousSelection?.players) }
    val queue = screen.queue
    if (queue != null) {
        QueuePage(queue, screen.notice, onCancel)
        return
    }

    val options = screen.options
    val counts = options?.playerCounts.orEmpty().distinct().sorted()
    val timeOptions = options?.timeOptions.orEmpty()
    val category = categoryChoice?.takeIf { it in screen.categories } ?: screen.categories.firstOrNull()
    val timer = timeOptions.firstOrNull { it.id == timerChoice } ?: timeOptions.firstOrNull()
    val players = playerChoice?.takeIf { it in counts } ?: counts.firstOrNull()
    val selection = if (!screen.busy && category != null && timer != null && players != null) {
        MatchSelection(category, timer, players)
    } else {
        null
    }

    ArenaPage(
        top = { ArenaBrandBar(screen.profile, onBack) },
        bottom = {
            Column {
                if (selection != null) {
                    ArenaActions {
                        ArenaButton(
                            R.string.find_match,
                            onClick = { onJoin(selection) },
                            forward = true,
                            pressMotion = true,
                        )
                    }
                }
                ArenaNavigation(onPlay = {}, onProfile = onBack, playing = true)
            }
        },
    ) {
        ArenaHero(
            title = stringResource(R.string.find_a_match),
            description = stringResource(R.string.multiplayer_trivia),
            symbol = R.drawable.ic_arena_bolt,
        )
        when {
            screen.busy -> ArenaState(stringResource(R.string.loading_match_options), Icons.Filled.Search, loading = true)
            category != null && timer != null && players != null -> {
                CategoryPicker(screen.categories, category) { categoryChoice = it }
                TimerPicker(timeOptions, timer) { timerChoice = it.id }
                PlayerStepper(counts, players) { playerChoice = it }
            }
            else -> {
                if (screen.notice == null) {
                    ArenaBanner(stringResource(R.string.no_match_options), Icons.Filled.Warning)
                }
                ArenaSecondaryButton(R.string.try_again, onRetry)
            }
        }
        Notice(screen.notice)
    }
}

@Composable
private fun QueuePage(queue: QueueEntry, notice: Int?, onCancel: () -> Unit) {
    val status = when {
        queue.cancelling -> R.string.cancelling_search
        queue.acknowledged -> R.string.waiting_for_players
        else -> R.string.joining_queue
    }
    ArenaPage(
        top = {
            ArenaMatchBar(
                stringResource(R.string.matchmaking),
                Icons.AutoMirrored.Filled.ArrowBack,
                stringResource(R.string.cancel_search),
                onCancel,
            )
        },
        bottom = { ArenaActions { ArenaSecondaryButton(R.string.cancel_search, onCancel, enabled = !queue.cancelling) } },
        centered = true,
    ) {
        ArenaState(stringResource(status), Icons.Filled.Search, loading = true)
        SelectionSummary(queue.selection)
        Notice(notice)
    }
}

@Composable
private fun CategoryPicker(categories: List<String>, selected: String, onSelect: (String) -> Unit) {
    val colors = MaterialTheme.colorScheme
    val fontScale = LocalDensity.current.fontScale.coerceAtLeast(1f)
    var animatedCategory by remember { mutableStateOf<String?>(null) }
    var activation by remember { mutableIntStateOf(0) }
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Text(stringResource(R.string.category), style = MaterialTheme.typography.titleSmall)
        BoxWithConstraints(Modifier.fillMaxWidth().selectableGroup()) {
            val columns = (maxWidth.value / (100 * fontScale)).toInt().coerceIn(1, 3)
            Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                categories.chunked(columns).forEach { row ->
                    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        row.forEach { category ->
                            val active = category == selected
                            Surface(
                                modifier = Modifier.weight(1f),
                                shape = MaterialTheme.shapes.medium,
                                border = BorderStroke(1.dp, if (active) colors.primary else colors.outlineVariant),
                                color = if (active) colors.primaryContainer else colors.surface,
                                contentColor = if (active) colors.onPrimaryContainer else colors.onSurface,
                            ) {
                                Column(
                                    Modifier
                                        .fillMaxWidth()
                                        .heightIn(min = 66.dp)
                                        .selectable(active, role = Role.RadioButton) {
                                            onSelect(category)
                                            animatedCategory = category
                                            activation++
                                        }
                                        .padding(horizontal = 6.dp, vertical = 7.dp),
                                    verticalArrangement = Arrangement.spacedBy(4.dp, Alignment.CenterVertically),
                                    horizontalAlignment = Alignment.CenterHorizontally,
                                ) {
                                    CategoryGlyph(
                                        category,
                                        activation = if (animatedCategory == category) activation else 0,
                                        tint = if (active) colors.primary else colors.onSurfaceVariant,
                                    )
                                    Text(categoryName(category), style = MaterialTheme.typography.labelLarge, textAlign = TextAlign.Center)
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun CategoryGlyph(category: String, activation: Int, tint: Color) {
    val motion = remember { Animatable(1f) }
    LaunchedEffect(activation) {
        if (activation == 0) {
            motion.snapTo(1f)
        } else {
            motion.snapTo(0f)
            motion.animateTo(1f, tween(durationMillis = 580, easing = LinearEasing))
        }
    }

    Box(Modifier.size(22.dp), contentAlignment = Alignment.Center) {
        Icon(
            painterResource(categorySymbol(category)),
            contentDescription = null,
            modifier = Modifier.size(22.dp).graphicsLayer {
                val phase = motion.value
                val arc = sin(PI * phase).toFloat()
                translationY = if (category == "SPORT") -6.dp.toPx() * arc else 0f
                rotationZ = if (category == "MUSIC") sin(4 * PI * phase).toFloat() * arc * 18f else 0f
                val pulse = if (category == "TECH") 1f + .16f * arc * arc else 1f
                scaleX = pulse
                scaleY = pulse
            },
            tint = tint,
        )
        if (category == "TECH") {
            Canvas(Modifier.matchParentSize()) {
                val phase = motion.value
                if (phase > 0f && phase < 1f) {
                    val y = size.height * phase
                    drawLine(
                        color = tint,
                        start = Offset(0f, y),
                        end = Offset(size.width, y),
                        strokeWidth = 1.5.dp.toPx(),
                        cap = StrokeCap.Round,
                        alpha = sin(PI * phase).toFloat() * .7f,
                    )
                }
            }
        }
    }
}

@Composable
private fun TimerPicker(options: List<TimeOption>, selected: TimeOption, onSelect: (TimeOption) -> Unit) {
    val colors = MaterialTheme.colorScheme
    val fontScale = LocalDensity.current.fontScale
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Text(stringResource(R.string.question_timer), style = MaterialTheme.typography.titleSmall)
        Surface(color = colors.surfaceVariant, shape = MaterialTheme.shapes.medium) {
            Column(
                Modifier.fillMaxWidth().selectableGroup().padding(4.dp),
                verticalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                options.chunked(if (fontScale >= 1.5f) 1 else 2).forEach { row ->
                    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        row.forEach { option ->
                            val active = option.id == selected.id
                            Surface(
                                Modifier.weight(1f),
                                shape = MaterialTheme.shapes.small,
                                color = if (active) colors.surface else colors.surfaceVariant,
                                border = if (active) BorderStroke(1.dp, colors.primary) else null,
                            ) {
                                Text(
                                    text = option.label,
                                    modifier = Modifier
                                        .selectable(active, role = Role.RadioButton) { onSelect(option) }
                                        .heightIn(min = 48.dp)
                                        .padding(horizontal = 7.dp, vertical = 13.dp),
                                    color = if (active) colors.onSurface else colors.onSurfaceVariant,
                                    style = MaterialTheme.typography.labelLarge,
                                    textAlign = TextAlign.Center,
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun PlayerStepper(counts: List<Int>, selected: Int, onSelect: (Int) -> Unit) {
    val index = counts.indexOf(selected)
    val label = stringResource(R.string.players_selected, selected)
    Row(
        Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(stringResource(R.string.players), Modifier.weight(1f), style = MaterialTheme.typography.titleSmall)
        OutlinedIconButton(
            onClick = { counts.getOrNull(index - 1)?.let(onSelect) },
            enabled = index > 0,
            shape = MaterialTheme.shapes.small,
            modifier = Modifier.size(48.dp),
        ) {
            val fewerLabel = stringResource(R.string.fewer_players)
            val tint = LocalContentColor.current
            Canvas(Modifier.size(19.dp).semantics { contentDescription = fewerLabel }) {
                drawLine(
                    color = tint,
                    start = Offset(size.width * .2f, center.y),
                    end = Offset(size.width * .8f, center.y),
                    strokeWidth = 2.dp.toPx(),
                    cap = StrokeCap.Round,
                )
            }
        }
        Text(
            selected.toString(),
            Modifier.widthIn(min = 24.dp).semantics { contentDescription = label },
            style = MaterialTheme.typography.titleMedium,
            textAlign = TextAlign.Center,
        )
        OutlinedIconButton(
            onClick = { counts.getOrNull(index + 1)?.let(onSelect) },
            enabled = index < counts.lastIndex,
            shape = MaterialTheme.shapes.small,
            modifier = Modifier.size(48.dp),
        ) {
            Icon(Icons.Filled.Add, stringResource(R.string.more_players))
        }
    }
}

@Composable
internal fun SelectionSummary(selection: MatchSelection) {
    val timer = when {
        selection.timeOption.label.isNotEmpty() -> selection.timeOption.label
        selection.timeOption.seconds > 0 -> pluralStringResource(
            R.plurals.timer_duration, selection.timeOption.seconds, selection.timeOption.seconds,
        )
        else -> stringResource(R.string.default_timer)
    }
    Text(
        pluralStringResource(
            R.plurals.selection_summary,
            selection.players,
            categoryName(selection.category),
            timer,
            selection.players,
        ),
        modifier = Modifier.fillMaxWidth(),
        textAlign = TextAlign.Center,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        style = MaterialTheme.typography.bodyMedium,
    )
}

@Composable
internal fun categoryName(category: String): String = when (category) {
    "SPORT" -> stringResource(R.string.category_sports)
    "MUSIC" -> stringResource(R.string.category_music)
    "TECH" -> stringResource(R.string.category_tech)
    else -> category
}

private fun categorySymbol(category: String): Int = when (category) {
    "SPORT" -> R.drawable.ic_arena_trophy
    "MUSIC" -> R.drawable.ic_arena_headphones
    "TECH" -> R.drawable.ic_arena_chip
    else -> R.drawable.ic_arena_bolt
}

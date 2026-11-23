package com.brainblitz.android

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.KeyboardArrowDown
import androidx.compose.material.icons.filled.KeyboardArrowUp
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.produceState
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.role
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import java.time.Duration
import java.time.Instant
import kotlinx.coroutines.delay

@Composable
internal fun MatchPage(
    screen: AppScreen.Match,
    onReady: () -> Unit,
    onAnswer: (Int, String) -> Unit,
    onPlayAgain: () -> Unit,
    onRetry: () -> Unit,
    onLeave: () -> Unit,
) {
    val phase = screen.game.phase
    val secondsToReady = (phase as? GamePhase.Ready)?.deadline?.let { countdown(it) }
    val title = when (phase) {
        is GamePhase.Ready, GamePhase.Cancelled -> stringResource(R.string.lobby)
        is GamePhase.Round -> stringResource(R.string.round_number, phase.number, phase.totalQuestions)
        is GamePhase.Finished -> stringResource(R.string.results)
        GamePhase.BetweenRounds -> stringResource(R.string.round_complete)
    }

    ArenaPage(
        top = {
            ArenaMatchBar(
                title,
                if (screen.game.terminal) Icons.AutoMirrored.Filled.ArrowBack else Icons.Filled.Close,
                stringResource(if (screen.game.terminal) R.string.back_to_profile else R.string.leave_match),
                onLeave,
            )
        },
        bottom = {
            when {
                screen.game.terminal -> ArenaActions {
                    ArenaButton(R.string.play_again, onPlayAgain, forward = true, pressMotion = true)
                    ArenaSecondaryButton(R.string.back_to_profile, onLeave)
                }
                phase is GamePhase.Ready && screen.recovery == null && !phase.waiting -> ArenaActions {
                    ArenaButton(
                        R.string.ready, onReady, enabled = secondsToReady != 0L, forward = true, pressMotion = true,
                    )
                }
            }
        },
        centered = phase is GamePhase.Ready || phase == GamePhase.Cancelled,
    ) {
        when (phase) {
            is GamePhase.Ready -> {
                ReadyLobby(screen, phase, secondsToReady)
                screen.recovery?.let { RecoveryControls(it, onRetry) }
            }
            is GamePhase.Round -> QuestionPage(
                phase, screen.selection, screen.recovery, onRetry,
            ) { onAnswer(phase.number, it) }
            GamePhase.BetweenRounds -> {
                val recovery = screen.recovery
                if (recovery != null) RecoveryControls(recovery, onRetry)
                else ArenaBanner(stringResource(R.string.waiting_for_round), Icons.Filled.Info)
            }
            is GamePhase.Finished -> Standings(phase.scores, screen.profile.id)
            GamePhase.Cancelled -> ArenaState(
                stringResource(R.string.match_expired),
                Icons.Filled.Info,
                about = stringResource(R.string.ready_deadline_expired),
            )
        }
        Notice(screen.notice)
        screen.game.lastRecap?.let {
            RecapCard(it.roundNumber, it.recap, initiallyExpanded = phase == GamePhase.BetweenRounds)
        }
    }
}

@Composable
private fun ReadyLobby(screen: AppScreen.Match, ready: GamePhase.Ready, seconds: Long?) {
    ArenaState(
        stringResource(R.string.match_found),
        Icons.Filled.Check,
        about = if (screen.recovery != null) null
        else stringResource(if (ready.waiting) R.string.waiting_for_ready_players else R.string.everyone_confirms_ready),
    )
    SelectionSummary(screen.selection)
    if (seconds != null) {
        Text(
            pluralStringResource(R.plurals.ready_seconds, seconds.toInt(), seconds),
            Modifier.fillMaxWidth(),
            style = MaterialTheme.typography.titleMedium,
            textAlign = TextAlign.Center,
            color = if (seconds <= 5) MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.primary,
        )
    }
    if (BuildConfig.DEBUG) {
        Text(
            stringResource(R.string.match_reference, screen.game.gameId),
            Modifier.fillMaxWidth(),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
internal fun RecoveryControls(status: RecoveryStatus, onRetry: () -> Unit, duringQuestion: Boolean = false) {
    val colors = MaterialTheme.colorScheme
    val retry = status == RecoveryStatus.Retry
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        ArenaBanner(
            title = stringResource(if (retry) R.string.recovery_failed else R.string.reconnecting_match),
            icon = if (retry) Icons.Filled.Warning else Icons.Filled.Refresh,
            description = if (duringQuestion) stringResource(R.string.round_timer_continues) else null,
            container = if (retry) colors.errorContainer else colors.primaryContainer,
            textColor = if (retry) colors.onErrorContainer else colors.onPrimaryContainer,
        )
        if (retry) ArenaButton(R.string.try_again, onRetry)
    }
}

@Composable
private fun countdown(deadline: Instant): Long {
    val seconds by produceState(secondsUntil(deadline, Instant.now()), deadline) {
        while (true) {
            val now = Instant.now()
            value = secondsUntil(deadline, now)
            if (value == 0L) break
            delay(Duration.between(now, deadline).toMillis() % 1_000 + 1)
        }
    }
    return seconds
}

@Composable
private fun QuestionPage(
    round: GamePhase.Round,
    selection: MatchSelection,
    recovery: RecoveryStatus?,
    onRetry: () -> Unit,
    onAnswer: (String) -> Unit,
) {
    val seconds = countdown(round.question.deadline)
    val submission = round.submission
    QuestionTimer(selection, seconds)
    ArenaHero(title = round.question.content, titleStyle = MaterialTheme.typography.titleLarge)
    if (recovery != null) RecoveryControls(recovery, onRetry, duringQuestion = seconds > 0)
    if (submission != null) {
        AnswerFeedbackBanner(submission)
    } else if (seconds == 0L) {
        val colors = MaterialTheme.colorScheme
        ArenaBanner(
            stringResource(R.string.time_expired),
            Icons.Filled.Info,
            stringResource(R.string.waiting_for_round),
            colors.errorContainer,
            colors.onErrorContainer,
        )
    }

    val enabled = recovery == null && submission == null && seconds > 0
    Column(verticalArrangement = Arrangement.spacedBy(7.dp)) {
        round.question.choices.forEachIndexed { index, choice ->
            val picked = submission?.choice == choice
            AnswerRow(
                choice,
                index,
                picked,
                enabled,
                verdict = if (picked) submission.feedback?.isCorrect else null,
            ) { onAnswer(choice) }
        }
    }
}

@Composable
private fun QuestionTimer(selection: MatchSelection, seconds: Long) {
    val colors = MaterialTheme.colorScheme
    val urgent = seconds <= 3
    val secondsLabel = pluralStringResource(R.plurals.seconds_remaining, seconds.toInt(), seconds)
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Text(
                categoryName(selection.category),
                Modifier.weight(1f),
                color = colors.onSurfaceVariant,
                style = MaterialTheme.typography.titleSmall,
            )
            Text(
                stringResource(R.string.timer_readout, seconds),
                Modifier.clearAndSetSemantics { contentDescription = secondsLabel },
                color = if (urgent) colors.error else colors.onSurface,
                style = MaterialTheme.typography.titleLarge,
            )
        }
        val duration = selection.timeOption.seconds
        if (duration > 0) {
            val fraction = (seconds.toFloat() / duration).coerceIn(0f, 1f)
            Box(Modifier.fillMaxWidth().height(6.dp).clip(CircleShape).background(colors.surfaceVariant)) {
                Box(Modifier.fillMaxWidth(fraction).height(6.dp).background(if (urgent) colors.error else colors.primary, CircleShape))
            }
        }
    }
}

@Composable
private fun AnswerFeedbackBanner(submission: Submission) {
    val colors = MaterialTheme.colorScheme
    val tones = LocalArenaTones.current
    when (submission.feedback?.isCorrect) {
        true -> ArenaBanner(
            stringResource(R.string.answer_correct),
            Icons.Filled.Check,
            stringResource(R.string.answer_locked),
            tones.correctSurface,
            tones.correctText,
        )
        false -> ArenaBanner(
            stringResource(R.string.answer_incorrect),
            Icons.Filled.Close,
            stringResource(R.string.answer_locked),
            colors.errorContainer,
            colors.onErrorContainer,
        )
        null -> ArenaBanner(
            stringResource(R.string.answer_sent),
            Icons.Filled.Check,
            stringResource(R.string.confirming_answer),
        )
    }
}

@Composable
private fun AnswerRow(
    choice: String,
    index: Int,
    picked: Boolean,
    enabled: Boolean,
    verdict: Boolean?,
    onAnswer: () -> Unit,
) {
    val colors = MaterialTheme.colorScheme
    val tones = LocalArenaTones.current
    val fill = when {
        verdict == true -> tones.correctSurface
        verdict == false -> colors.errorContainer
        picked -> colors.primaryContainer
        else -> colors.surface
    }
    val text = when {
        verdict == true -> tones.correctText
        verdict == false -> colors.onErrorContainer
        picked -> colors.onPrimaryContainer
        enabled -> colors.onSurface
        else -> colors.onSurfaceVariant
    }
    val interactions = remember { MutableInteractionSource() }
    Surface(
        onClick = onAnswer,
        enabled = enabled,
        modifier = Modifier
            .fillMaxWidth()
            .arenaPressMotion(interactions, enabled, pressedScale = .99f)
            .semantics {
                role = Role.Button
                selected = picked
            },
        shape = MaterialTheme.shapes.medium,
        color = fill,
        contentColor = text,
        border = BorderStroke(1.dp, if (picked) text else colors.outlineVariant),
        interactionSource = interactions,
    ) {
        Row(
            Modifier.fillMaxWidth().heightIn(min = 54.dp).padding(horizontal = 10.dp, vertical = 8.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Box(
                Modifier.size(31.dp).background(colors.surfaceVariant, MaterialTheme.shapes.small),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    if (index < 26) ('A' + index).toString() else (index + 1).toString(),
                    Modifier.clearAndSetSemantics {},
                    color = colors.onSurfaceVariant,
                    style = MaterialTheme.typography.labelLarge,
                )
            }
            Text(choice, Modifier.weight(1f), style = MaterialTheme.typography.titleSmall)
            if (verdict != null) Icon(if (verdict) Icons.Filled.Check else Icons.Filled.Close, null, Modifier.size(19.dp))
        }
    }
}

@Composable
private fun RecapCard(round: Int, recap: RoundRecap, initiallyExpanded: Boolean) {
    var expanded by rememberSaveable(round) { mutableStateOf(initiallyExpanded) }
    val colors = MaterialTheme.colorScheme
    Surface(
        color = colors.surface,
        shape = MaterialTheme.shapes.medium,
        border = BorderStroke(1.dp, colors.outlineVariant),
    ) {
        Column(Modifier.fillMaxWidth()) {
            Surface(onClick = { expanded = !expanded }, color = colors.surface) {
                Row(
                    Modifier.fillMaxWidth().heightIn(min = 48.dp).padding(horizontal = 12.dp, vertical = 8.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(stringResource(R.string.round_recap, round), Modifier.weight(1f), style = MaterialTheme.typography.titleSmall)
                    Icon(
                        if (expanded) Icons.Filled.KeyboardArrowUp else Icons.Filled.KeyboardArrowDown,
                        stringResource(if (expanded) R.string.hide_recap else R.string.show_recap),
                    )
                }
            }
            if (expanded) {
                Column(Modifier.padding(start = 12.dp, end = 12.dp, bottom = 12.dp), verticalArrangement = Arrangement.spacedBy(5.dp)) {
                    Text(recap.question, style = MaterialTheme.typography.bodyMedium)
                    Text(
                        stringResource(R.string.correct_answer, recap.correctAnswer),
                        style = MaterialTheme.typography.titleSmall,
                        color = LocalArenaTones.current.correctText,
                    )
                    Text(
                        if (recap.answered) stringResource(R.string.your_answer, recap.playerAnswer)
                        else stringResource(R.string.no_answer_submitted),
                        style = MaterialTheme.typography.bodySmall,
                        color = colors.onSurfaceVariant,
                    )
                }
            }
        }
    }
}

@Composable
private fun Standings(scores: List<PlayerScore>, ownId: String?) {
    val colors = MaterialTheme.colorScheme
    val winningScore = scores.firstOrNull()?.point
    val self = scores.firstOrNull { it.playerId == ownId }
    val outcome = when {
        self == null || self.point != winningScore || scores.size < 2 -> null
        scores.count { it.point == winningScore } > 1 -> stringResource(R.string.tied_for_first)
        else -> stringResource(R.string.you_won)
    }
    ArenaHero(
        title = outcome ?: stringResource(R.string.final_standings),
        description = if (outcome != null) stringResource(R.string.final_standings) else null,
        symbol = R.drawable.ic_arena_trophy,
        titleStyle = MaterialTheme.typography.headlineMedium,
    )
    Surface(
        color = colors.surface,
        shape = MaterialTheme.shapes.large,
        border = BorderStroke(1.dp, colors.outlineVariant),
    ) {
        Column(Modifier.fillMaxWidth().padding(6.dp)) {
            var rank = 0
            scores.forEachIndexed { index, player ->
                if (index == 0 || scores[index - 1].point != player.point) rank = index + 1
                val name = player.displayName.ifBlank { stringResource(R.string.player_id, player.playerId) }
                val isYou = player.playerId == ownId
                if (index > 0) HorizontalDivider(color = colors.outlineVariant)
                Row(
                    Modifier
                        .fillMaxWidth()
                        .clip(MaterialTheme.shapes.small)
                        .background(if (isYou) colors.primaryContainer else colors.surface)
                        .padding(horizontal = 8.dp, vertical = 10.dp),
                    horizontalArrangement = Arrangement.spacedBy(7.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        stringResource(R.string.standing_rank, rank),
                        Modifier.widthIn(min = 22.dp),
                        color = colors.onSurfaceVariant,
                        style = MaterialTheme.typography.titleSmall,
                    )
                    Box(
                        Modifier.size(30.dp).background(colors.surfaceVariant, CircleShape),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(playerInitials(name), Modifier.clearAndSetSemantics {}, style = MaterialTheme.typography.labelSmall)
                    }
                    Text(
                        if (isYou) stringResource(R.string.player_you, name) else name,
                        Modifier.weight(1f),
                        style = MaterialTheme.typography.titleSmall,
                    )
                    Text(
                        pluralStringResource(R.plurals.player_points, player.point, player.point),
                        Modifier.weight(.52f, fill = false),
                        textAlign = TextAlign.End,
                        style = MaterialTheme.typography.labelLarge,
                    )
                }
            }
        }
    }
}

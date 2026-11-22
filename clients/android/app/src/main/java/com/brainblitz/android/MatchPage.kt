package com.brainblitz.android

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedCard
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.produceState
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
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
    onLeave: () -> Unit,
) {
    val phase = screen.game.phase
    if (!screen.disconnected) {
        when (phase) {
            is GamePhase.Ready -> {
                Heading(stringResource(R.string.match_found))
                SelectionSummary(screen.selection)
                if (BuildConfig.DEBUG) {
                    Text(stringResource(R.string.match_reference, screen.game.gameId), style = MaterialTheme.typography.bodySmall)
                }
                if (phase.waiting) {
                    Status(R.string.waiting_for_ready_players)
                } else {
                    Button(onClick = onReady) { Text(stringResource(R.string.ready)) }
                }
            }
            is GamePhase.Round -> QuestionPage(phase) { onAnswer(phase.number, it) }
            GamePhase.BetweenRounds -> Status(R.string.waiting_for_round)
            is GamePhase.Finished -> {
                Heading(stringResource(R.string.final_standings))
                Standings(phase.scores)
                Button(onClick = onPlayAgain) { Text(stringResource(R.string.play_again)) }
            }
        }
        screen.game.lastRecap?.let { RecapCard(it.roundNumber, it.recap) }
    }
    Notice(screen.notice)
    val backLabel = if (screen.disconnected || phase is GamePhase.Finished) R.string.back_to_profile else R.string.leave_match
    OutlinedButton(onClick = onLeave) { Text(stringResource(backLabel)) }
}

@Composable
private fun QuestionPage(round: GamePhase.Round, onAnswer: (String) -> Unit) {
    val deadline = round.question.deadline
    val seconds by produceState(round.secondsRemaining(Instant.now()), deadline) {
        while (true) {
            val now = Instant.now()
            value = round.secondsRemaining(now)
            if (value == 0L) break
            delay(Duration.between(now, deadline).toMillis() % 1_000 + 1)
        }
    }
    val submission = round.submission

    Heading(stringResource(R.string.round_number, round.number, round.totalQuestions))
    Text(
        pluralStringResource(R.plurals.seconds_remaining, seconds.toInt(), seconds),
        style = MaterialTheme.typography.titleMedium,
    )
    Text(round.question.content, style = MaterialTheme.typography.titleLarge)
    if (submission != null) {
        Text(stringResource(R.string.your_answer, submission.choice))
        Status(when (submission.feedback?.isCorrect) {
            true -> R.string.answer_correct
            false -> R.string.answer_incorrect
            null -> R.string.answer_sent
        })
    } else if (seconds == 0L) {
        Status(R.string.time_expired)
    }

    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        round.question.choices.forEach { choice ->
            OutlinedButton(
                onClick = { onAnswer(choice) },
                enabled = submission == null && seconds > 0,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(choice)
            }
        }
    }
}

@Composable
private fun RecapCard(round: Int, recap: RoundRecap) {
    OutlinedCard(Modifier.fillMaxWidth()) {
        Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(stringResource(R.string.round_recap, round), style = MaterialTheme.typography.titleMedium)
            Text(recap.question)
            Status(when {
                !recap.answered -> R.string.no_answer_submitted
                recap.isCorrect -> R.string.answer_correct
                else -> R.string.answer_incorrect
            })
            if (recap.answered) Text(stringResource(R.string.your_answer, recap.playerAnswer))
            Text(stringResource(R.string.correct_answer, recap.correctAnswer))
        }
    }
}

@Composable
private fun Standings(scores: List<PlayerScore>) {
    var rank = 0
    scores.forEachIndexed { index, player ->
        if (index == 0 || scores[index - 1].point != player.point) rank = index + 1
        val name = player.displayName.ifBlank { stringResource(R.string.player_id, player.playerId) }
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            Text(stringResource(R.string.player_placement, rank, name), Modifier.weight(1f))
            Text(pluralStringResource(R.plurals.player_points, player.point, player.point))
        }
    }
}

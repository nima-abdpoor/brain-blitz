package com.brainblitz.android

import java.time.Duration
import java.time.Instant

internal sealed interface GamePhase {
    data class Ready(val waiting: Boolean = false) : GamePhase
    data class Round(
        val number: Int,
        val totalQuestions: Int,
        val question: GameQuestion,
        val submission: Submission? = null,
    ) : GamePhase {
        fun submit(choice: String, now: Instant): Round? {
            if (submission != null || now >= question.deadline || choice !in question.choices) return null
            return copy(submission = Submission(choice))
        }

        fun secondsRemaining(now: Instant): Long {
            val remaining = Duration.between(now, question.deadline)
            if (remaining.isNegative) return 0
            return remaining.seconds + if (remaining.nano > 0) 1 else 0
        }
    }

    data object BetweenRounds : GamePhase
    data class Finished(val scores: List<PlayerScore>) : GamePhase
}

internal data class Submission(val choice: String, val feedback: AnswerFeedback? = null)

internal data class GameProgress(
    val gameId: String,
    val phase: GamePhase = GamePhase.Ready(),
    val lastRecap: MatchEvent.RoundClosed? = null,
) {
    val roundNumber: Int?
        get() = (phase as? GamePhase.Round)?.number ?: lastRecap?.roundNumber

    fun apply(event: MatchEvent.Gameplay): GameProgress {
        if (event.gameId != gameId || phase is GamePhase.Finished) return this
        val round = phase as? GamePhase.Round
        return when (event) {
            is MatchEvent.QuestionPublished -> {
                if (event.roundNumber <= (roundNumber ?: 0)) return this
                copy(phase = GamePhase.Round(event.roundNumber, event.totalQuestions, event.question))
            }
            is MatchEvent.AnswerAccepted -> {
                val submission = round?.submission ?: return this
                if (event.feedback.questionId != round.question.id || submission.feedback != null) return this
                copy(phase = round.copy(submission = submission.copy(feedback = event.feedback)))
            }
            is MatchEvent.RoundClosed -> {
                if (round == null || event.roundNumber != round.number || event.recap.questionId != round.question.id) {
                    return this
                }
                copy(phase = GamePhase.BetweenRounds, lastRecap = event)
            }
            is MatchEvent.Completed -> copy(phase = GamePhase.Finished(
                event.scores.sortedWith(compareByDescending<PlayerScore> { it.point }.thenBy { it.playerId }),
            ))
        }
    }
}

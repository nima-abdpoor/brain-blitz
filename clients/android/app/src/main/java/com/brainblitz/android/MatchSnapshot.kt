package com.brainblitz.android

import java.time.Instant
import java.time.format.DateTimeParseException
import kotlinx.serialization.Serializable
import kotlinx.serialization.SerializationException

@Serializable
internal data class MatchSnapshot(
    val gameId: String,
    val category: String,
    val timeOptionId: String,
    val questionSeconds: Int,
    val players: Int,
    val phase: String,
    val ready: Boolean = false,
    val readyUntil: String? = null,
    val roundNumber: Int = 0,
    val totalQuestions: Int = 0,
    val question: WireQuestion? = null,
    val submission: RecoveredSubmission? = null,
    val recap: RoundRecap? = null,
    val recapRound: Int = 0,
    val scores: List<PlayerScore> = emptyList(),
) {
    fun selection() = MatchSelection(category, TimeOption(timeOptionId, "", questionSeconds), players)

    fun progress(): GameProgress {
        require(gameId.isNotBlank())
        val phase = when (phase) {
            "ready" -> GamePhase.Ready(waiting = ready, deadline = readyUntil?.let(::parseDeadline))
            "round" -> {
                require(roundNumber in 1..totalQuestions)
                val active = requireNotNull(question).toQuestion()
                val answer = submission?.let { Submission(it.choice, AnswerFeedback(active.id, it.isCorrect)) }
                GamePhase.Round(roundNumber, totalQuestions, active, answer)
            }
            "finished" -> GamePhase.Finished(scores.inStandingOrder())
            else -> throw SerializationException("Unknown match phase")
        }
        return GameProgress(gameId, phase, recap?.let { MatchEvent.RoundClosed(gameId, recapRound, it) })
    }
}

@Serializable
internal data class RecoveredSubmission(val choice: String, val isCorrect: Boolean)

@Serializable
internal data class WireQuestion(val id: String, val content: String, val choices: List<String>, val ttl: String) {
    fun toQuestion(): GameQuestion {
        require(id.isNotBlank() && choices.isNotEmpty())
        return GameQuestion(id, content, choices, parseDeadline(ttl))
    }
}

internal fun parseDeadline(value: String): Instant = try {
    Instant.parse(value)
} catch (error: DateTimeParseException) {
    throw SerializationException("Invalid game deadline", error)
}

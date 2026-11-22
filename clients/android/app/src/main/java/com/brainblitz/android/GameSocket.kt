package com.brainblitz.android

import java.io.IOException
import java.time.Instant
import java.time.format.DateTimeParseException
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.channels.ReceiveChannel
import kotlinx.serialization.Serializable
import kotlinx.serialization.SerializationException
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.decodeFromJsonElement
import kotlinx.serialization.json.jsonObject
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener

internal data class MatchSelection(val category: String, val timeOption: TimeOption, val players: Int)

internal sealed interface MatchEvent {
    data class Categories(val names: List<String>) : MatchEvent
    data object Queued : MatchEvent
    data object Rejected : MatchEvent
    data class Created(val gameId: String) : MatchEvent

    sealed interface Gameplay : MatchEvent {
        val gameId: String
    }

    data class QuestionPublished(
        override val gameId: String,
        val roundNumber: Int,
        val totalQuestions: Int,
        val question: GameQuestion,
    ) : Gameplay

    data class AnswerAccepted(override val gameId: String, val feedback: AnswerFeedback) : Gameplay
    data class RoundClosed(override val gameId: String, val roundNumber: Int, val recap: RoundRecap) : Gameplay
    data class Completed(override val gameId: String, val scores: List<PlayerScore>) : Gameplay
}

internal data class GameQuestion(val id: String, val content: String, val choices: List<String>, val deadline: Instant)

@Serializable
internal data class AnswerFeedback(val questionId: String, val isCorrect: Boolean)

@Serializable
internal data class RoundRecap(
    val questionId: String,
    val question: String,
    val playerAnswer: String,
    val correctAnswer: String,
    val answered: Boolean,
    val isCorrect: Boolean,
)

@Serializable
internal data class PlayerScore(val playerId: String, val point: Int, val displayName: String = "")

internal class GameSocket(client: OkHttpClient, request: Request, private val json: Json) {
    private val opened = CompletableDeferred<Unit>()
    private val messages = Channel<MatchEvent>(capacity = 16)
    val events: ReceiveChannel<MatchEvent> = messages

    private val socket = client.newWebSocket(request, object : WebSocketListener() {
        override fun onOpen(webSocket: WebSocket, response: Response) {
            opened.complete(Unit)
        }

        override fun onMessage(webSocket: WebSocket, text: String) {
            val event = try {
                decodeEvent(text)
            } catch (_: IllegalArgumentException) {
                fail(IOException("Unexpected game frame"))
                return
            }
            if (event != null && messages.trySend(event).isFailure) {
                fail(IOException("Game event buffer unavailable"))
            }
        }

        override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
            webSocket.close(1000, null)
            fail(IOException("Game connection closed"))
        }

        override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
            response?.body?.close()
            val failure = if (response == null) {
                IOException("Game connection failed", t)
            } else {
                GatewayHttpException(Gateway.GAME_SOCKET_PATH, response.code)
            }
            fail(failure)
        }
    })

    suspend fun join(selection: MatchSelection, displayName: String) {
        opened.await()
        send(JoinCommand(
            command = "ADD_TO_WAITING_LIST",
            category = selection.category,
            timeOptionId = selection.timeOption.id,
            players = selection.players,
            displayName = displayName,
        ))
    }

    fun ready(gameId: String) = send(ReadyCommand("READY", gameId))

    fun answer(gameId: String, questionId: String, choice: String) =
        send(AnswerCommand("ANSWER", Answer(gameId, questionId, choice)))

    private inline fun <reified T> send(command: T) {
        if (!socket.send(json.encodeToString(command))) {
            throw IOException("Game connection closed")
        }
    }

    fun close() {
        messages.cancel()
        socket.cancel()
    }

    private fun fail(error: IOException) {
        opened.completeExceptionally(error)
        messages.close(error)
    }

    private fun decodeEvent(text: String): MatchEvent? {
        val frame = json.parseToJsonElement(text).jsonObject
        if ("event" !in frame && "categories" in frame) {
            return MatchEvent.Categories(json.decodeFromJsonElement<InitialOptions>(frame).categories)
        }
        val event = json.decodeFromJsonElement<Envelope>(frame)
        return when (event.event) {
            "ADDED_TO_WAITING_LIST" -> if (event.success) MatchEvent.Queued else MatchEvent.Rejected
            "ERROR" -> MatchEvent.Rejected
            "MATCH_CREATED" -> MatchEvent.Created(metadata<GameMetadata>(event).gameId)
            "QUESTIONS_PUBLISHED" -> {
                val data = metadata<QuestionMetadata>(event)
                val question = data.questions.singleOrNull()
                if (question == null || question.id.isBlank() || question.choices.isEmpty() ||
                    data.roundNumber !in 1..data.totalQuestions
                ) throw SerializationException("Invalid active round")
                val deadline = try {
                    Instant.parse(question.ttl)
                } catch (error: DateTimeParseException) {
                    throw SerializationException("Invalid question deadline", error)
                }
                MatchEvent.QuestionPublished(
                    data.gameId, data.roundNumber, data.totalQuestions,
                    GameQuestion(question.id, question.content, question.choices, deadline),
                )
            }
            "ANSWER_ACCEPTED" -> metadata<AnswerMetadata>(event).let {
                // Older game servers acknowledge the answer without a personal verdict.
                it.answerFeedback?.let { feedback -> MatchEvent.AnswerAccepted(it.gameId, feedback) }
            }
            "ROUND_CLOSED" -> metadata<RecapMetadata>(event).let {
                MatchEvent.RoundClosed(it.gameId, it.roundNumber, it.recap)
            }
            "COMPLETED" -> metadata<CompletionMetadata>(event).let {
                MatchEvent.Completed(it.gameId, it.leaderBoard.playerPoint)
            }
            else -> null
        }
    }

    private inline fun <reified T> metadata(event: Envelope): T {
        val payload = event.metaData
        if (!event.success || payload == null) throw SerializationException("Missing game event metadata")
        return json.decodeFromJsonElement(payload)
    }

    @Serializable
    private data class JoinCommand(
        val command: String,
        val category: String,
        val timeOptionId: String,
        val players: Int,
        val displayName: String,
    )

    @Serializable
    private data class ReadyCommand(val command: String, val gameId: String)

    @Serializable
    private data class AnswerCommand(val command: String, val answer: Answer)

    @Serializable
    private data class Answer(val gameId: String, val questionId: String, val choice: String)

    @Serializable
    private data class InitialOptions(val categories: List<String>)

    @Serializable
    private data class Envelope(val event: String, val success: Boolean, val metaData: JsonElement? = null)

    @Serializable
    private data class GameMetadata(val gameId: String) {
        init {
            require(gameId.isNotBlank())
        }
    }

    @Serializable
    private data class QuestionMetadata(
        val gameId: String,
        val roundNumber: Int,
        val totalQuestions: Int,
        val questions: List<WireQuestion>,
    )

    @Serializable
    private data class WireQuestion(val id: String, val content: String, val choices: List<String>, val ttl: String)

    @Serializable
    private data class AnswerMetadata(val gameId: String, val answerFeedback: AnswerFeedback? = null)

    @Serializable
    private data class RecapMetadata(val gameId: String, val roundNumber: Int, val recap: RoundRecap)

    @Serializable
    private data class CompletionMetadata(val gameId: String, val leaderBoard: LeaderBoard)

    @Serializable
    private data class LeaderBoard(val playerPoint: List<PlayerScore>)
}

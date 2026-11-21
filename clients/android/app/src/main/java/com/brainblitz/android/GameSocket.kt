package com.brainblitz.android

import java.io.IOException
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.channels.ReceiveChannel
import kotlinx.serialization.Serializable
import kotlinx.serialization.SerializationException
import kotlinx.serialization.json.Json
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
}

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
        val command = JoinCommand(
            command = "ADD_TO_WAITING_LIST",
            category = selection.category,
            timeOptionId = selection.timeOption.id,
            players = selection.players,
            displayName = displayName,
        )
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
            "MATCH_CREATED" -> {
                val gameId = event.metaData?.gameId
                if (!event.success || gameId.isNullOrBlank()) {
                    throw SerializationException("Missing game ID")
                }
                MatchEvent.Created(gameId)
            }
            else -> null
        }
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
    private data class InitialOptions(val categories: List<String>)

    @Serializable
    private data class Envelope(val event: String, val success: Boolean, val metaData: GameMetadata? = null)

    @Serializable
    private data class GameMetadata(val gameId: String? = null)
}

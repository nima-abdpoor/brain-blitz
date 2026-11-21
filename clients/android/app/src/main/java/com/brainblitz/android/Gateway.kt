package com.brainblitz.android

import java.io.IOException
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean
import kotlin.coroutines.resumeWithException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeoutOrNull
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.HttpUrl
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener

internal data class GatewayCheck(
    val displayName: String,
    val playerCounts: List<Int>,
    val timeOptions: List<TimeOption>,
    val categories: List<String>,
)

@Serializable
internal data class TimeOption(val id: String, val label: String, val seconds: Int)

internal class GatewayHttpException(val endpoint: String, val status: Int) : IOException()

internal class Gateway(
    private val origin: HttpUrl,
    private val client: OkHttpClient = OkHttpClient.Builder()
        .callTimeout(8, TimeUnit.SECONDS)
        .followRedirects(false)
        .build(),
) {
    private val json = Json { ignoreUnknownKeys = true }

    suspend fun check(email: String, password: String): GatewayCheck {
        val login = decode<Login>(
            request(LOGIN_PATH).post(
                json.encodeToString(Credentials(email, password)).toRequestBody(JSON_MEDIA_TYPE),
            ).build(),
        )
        val profile = decode<Profile>(request(PROFILE_PATH, login.accessToken).build())
        val counts = decode<GameOptions>(request(GAME_OPTIONS_PATH, login.accessToken).build())
        val timers = decode<List<TimeOption>>(request(TIME_OPTIONS_PATH, login.accessToken).build())
        val initial = readInitialFrame(login.accessToken)
        return GatewayCheck(profile.displayName, counts.numberOfPlayers, timers, initial.categories)
    }

    private fun request(path: String, accessToken: String? = null): Request.Builder {
        val builder = Request.Builder().url(origin.newBuilder().encodedPath(path).build())
        if (accessToken != null) {
            builder.header("Authorization", "Bearer $accessToken")
        }
        return builder
    }

    private suspend inline fun <reified T> decode(request: Request): T = withContext(Dispatchers.IO) {
        client.newCall(request).execute().use { response ->
            if (!response.isSuccessful) {
                throw GatewayHttpException(request.url.encodedPath, response.code)
            }
            val body = response.body?.string() ?: throw IOException("Missing gateway response")
            json.decodeFromString<T>(body)
        }
    }

    private suspend fun readInitialFrame(accessToken: String): InitialFrame =
        withTimeoutOrNull(8_000) {
            suspendCancellableCoroutine { continuation ->
                val finished = AtomicBoolean()
                val listener = object : WebSocketListener() {
                    override fun onMessage(webSocket: WebSocket, text: String) {
                        if (finished.compareAndSet(false, true)) {
                            webSocket.close(1000, null)
                            continuation.resumeWith(runCatching { json.decodeFromString<InitialFrame>(text) })
                        }
                    }

                    override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                        webSocket.close(1000, null)
                        if (finished.compareAndSet(false, true)) {
                            continuation.resumeWithException(IOException("Socket closed before the initial frame"))
                        }
                    }

                    override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                        response?.body?.close()
                        if (finished.compareAndSet(false, true)) {
                            val failure = response?.let { GatewayHttpException(GAME_SOCKET_PATH, it.code) } ?: t
                            continuation.resumeWithException(failure)
                        }
                    }
                }
                val socket = client.newWebSocket(request(GAME_SOCKET_PATH, accessToken).build(), listener)
                continuation.invokeOnCancellation { socket.cancel() }
            }
        } ?: throw IOException("Timed out waiting for the initial game frame")

    @Serializable
    private class Credentials(val email: String, val password: String)

    @Serializable
    private class Login(val accessToken: String)

    @Serializable
    private class Profile(val displayName: String)

    @Serializable
    private class GameOptions(val numberOfPlayers: List<Int>)

    @Serializable
    private class InitialFrame(val categories: List<String>)

    companion object {
        const val LOGIN_PATH = "/user-service/public/api/v1/login"
        private const val PROFILE_PATH = "/user-service/api/v1/profile"
        private const val GAME_OPTIONS_PATH = "/game-service/api/v1/game-options"
        private const val TIME_OPTIONS_PATH = "/question-service/api/v1/question/time-options"
        private const val GAME_SOCKET_PATH = "/game-service/api/v1/process-game"
        private val JSON_MEDIA_TYPE = "application/json".toMediaType()
    }
}

package com.brainblitz.android

import java.io.IOException
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.SerializationException
import kotlinx.serialization.json.Json
import okhttp3.HttpUrl
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody

@Serializable
internal data class Profile(
    val displayName: String,
    val username: String,
    val statistics: PlayerStatistics? = null,
    val id: String? = null,
)

@Serializable
internal data class PlayerStatistics(
    val gamesPlayed: Long,
    val wins: Long,
    val losses: Long,
    val draws: Long,
    val totalPoints: Long,
    val bestScore: Long,
    val answersSubmitted: Long,
    val correctAnswers: Long,
    val winRate: Double,
    val averagePoints: Double,
    val accuracy: Double,
    val rank: Long,
    val rankedPlayers: Long,
)

@Serializable
internal data class TimeOption(val id: String, val label: String, val seconds: Int)

internal data class MatchOptions(val playerCounts: List<Int>, val timeOptions: List<TimeOption>)

internal class GatewayHttpException(
    val endpoint: String,
    val status: Int,
    val duplicateAccount: Boolean = false,
) : IOException() {
    val sessionRejected: Boolean
        get() = when (endpoint) {
            Gateway.LOGIN_PATH, Gateway.SIGNUP_PATH -> false
            Gateway.PROFILE_PATH -> status == 401 || status == 404
            else -> status == 401
        }
}

internal class Gateway(
    private val origin: HttpUrl,
    private val client: OkHttpClient = OkHttpClient.Builder()
        .callTimeout(8, TimeUnit.SECONDS)
        .followRedirects(false)
        .build(),
) {
    private val json = Json { ignoreUnknownKeys = true }
    private val heartbeatClient = client.newBuilder().pingInterval(10, TimeUnit.SECONDS).build()

    suspend fun signUp(email: String, password: String) {
        decode<SignUp>(accountRequest(SIGNUP_PATH, email, password))
    }

    suspend fun login(email: String, password: String): String =
        decode<Login>(accountRequest(LOGIN_PATH, email, password)).accessToken

    suspend fun profile(accessToken: String): Profile =
        decode(request(PROFILE_PATH, accessToken).build())

    suspend fun matchOptions(accessToken: String): MatchOptions {
        val gameOptions = decode<GameOptions>(request(GAME_OPTIONS_PATH, accessToken).build())
        val timeOptions = decode<List<TimeOption>>(request(TIME_OPTIONS_PATH, accessToken).build())
        return MatchOptions(gameOptions.numberOfPlayers, timeOptions)
    }

    fun openGame(accessToken: String, heartbeat: Boolean = false): GameSocket =
        GameSocket(if (heartbeat) heartbeatClient else client, request(GAME_SOCKET_PATH, accessToken).build(), json)

    private fun accountRequest(path: String, email: String, password: String): Request =
        request(path).post(
            json.encodeToString(Credentials(email, password)).toRequestBody(JSON_MEDIA_TYPE),
        ).build()

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
                val path = request.url.encodedPath
                val duplicate = path == SIGNUP_PATH && response.code == 400 &&
                    isDuplicateAccount(response.peekBody(4096).string())
                throw GatewayHttpException(path, response.code, duplicate)
            }
            val body = response.body?.string() ?: throw IOException("Missing gateway response")
            json.decodeFromString<T>(body)
        }
    }

    private fun isDuplicateAccount(body: String): Boolean = try {
        json.decodeFromString<AccountError>(body).message == "username already exists"
    } catch (_: SerializationException) {
        false
    }

    @Serializable
    private class Credentials(val email: String, val password: String)

    @Serializable
    private class SignUp(val displayName: String)

    @Serializable
    private class Login(val accessToken: String)

    @Serializable
    private class AccountError(val message: String)

    @Serializable
    private class GameOptions(val numberOfPlayers: List<Int>)

    companion object {
        const val SIGNUP_PATH = "/user-service/public/api/v1/signup"
        const val LOGIN_PATH = "/user-service/public/api/v1/login"
        const val PROFILE_PATH = "/user-service/api/v1/profile"
        const val GAME_SOCKET_PATH = "/game-service/api/v1/process-game"
        private const val GAME_OPTIONS_PATH = "/game-service/api/v1/game-options"
        private const val TIME_OPTIONS_PATH = "/question-service/api/v1/question/time-options"
        private val JSON_MEDIA_TYPE = "application/json".toMediaType()
    }
}

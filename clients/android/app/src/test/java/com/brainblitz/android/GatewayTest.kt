package com.brainblitz.android

import java.time.Instant
import java.util.Base64
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withTimeout
import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertThrows
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest

class GatewayTest {
    @get:Rule
    val server = MockWebServer()

    @Test
    fun createsAnAccountAndLoadsProfileAndMatchOptions() {
        reply("""{"displayName":"Player"}""")
        reply("""{"id":"5","accessToken":"access","refreshToken":"refresh"}""")
        reply("""{"id":"5","displayName":"Player","username":"player@example.com","statistics":null}""")
        reply("""{"minPlayers":2,"maxPlayers":3,"numberOfPlayers":[2,3]}""")
        reply("""[{"id":"quick","label":"Quick","seconds":10,"enabled":true}]""")

        runBlocking {
            val gateway = Gateway(server.url("/"))
            gateway.signUp("player@example.com", "example-password")
            val token = gateway.login("player@example.com", "example-password")
            assertEquals(Profile("Player", "player@example.com"), gateway.profile(token))
            assertEquals(MatchOptions(listOf(2, 3), listOf(TimeOption("quick", "Quick", 10))), gateway.matchOptions(token))
        }

        listOf(Gateway.SIGNUP_PATH, Gateway.LOGIN_PATH).forEach { path ->
            val request = nextRequest()
            assertEquals(path, request.path)
            assertEquals("POST", request.method)
            assertNull(request.getHeader("Authorization"))
            assertEquals("""{"email":"player@example.com","password":"example-password"}""", request.body.readUtf8())
        }
        listOf(
            Gateway.PROFILE_PATH,
            "/game-service/api/v1/game-options",
            "/question-service/api/v1/question/time-options",
        ).forEach { path ->
            val request = nextRequest()
            assertEquals(path, request.path)
            assertEquals("GET", request.method)
            assertEquals("Bearer access", request.getHeader("Authorization"))
        }
    }

    @Test
    fun distinguishesDuplicateAccountAndExpiredSessionFromBadPassword() {
        server.enqueue(MockResponse().setResponseCode(400).setBody("""{"message":"username already exists"}"""))
        server.enqueue(MockResponse().setResponseCode(403))
        server.enqueue(MockResponse().setResponseCode(401))
        val gateway = Gateway(server.url("/"))

        val duplicate = assertThrows(GatewayHttpException::class.java) {
            runBlocking { gateway.signUp("player@example.com", "example-password") }
        }
        assertTrue(duplicate.duplicateAccount)
        assertFalse(duplicate.sessionRejected)

        val login = assertThrows(GatewayHttpException::class.java) {
            runBlocking { gateway.login("player@example.com", "wrong-password") }
        }
        assertEquals(403, login.status)
        assertFalse(login.sessionRejected)

        val profile = assertThrows(GatewayHttpException::class.java) {
            runBlocking { gateway.profile("expired") }
        }
        assertTrue(profile.sessionRejected)
    }

    @Test
    fun joinsAndReadsMatchCreatedOnTheSameSocket() {
        val command = CompletableFuture<String>()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                webSocket.send("""{"categories":["MUSIC","SPORT"],"numberOfPlayers":[2,3]}""")
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                command.complete(text)
                webSocket.send("""{"success":true,"event":"ADDED_TO_WAITING_LIST","metaData":{"gameId":""}}""")
                webSocket.send("""{"success":true,"event":"MATCH_CREATED","metaData":{"gameId":"game-37"}}""")
            }
        }))

        runBlocking {
            val socket = Gateway(server.url("/")).openGame("access")
            try {
                withTimeout(5_000) {
                    socket.join(MatchSelection("MUSIC", TimeOption("quick", "Quick", 10), 3), "Player")
                    assertEquals(MatchEvent.Categories(listOf("MUSIC", "SPORT")), socket.events.receive())
                    assertEquals(MatchEvent.Queued, socket.events.receive())
                    assertEquals(MatchEvent.Created("game-37"), socket.events.receive())
                }
            } finally {
                socket.close()
            }
        }

        val request = nextRequest()
        assertEquals(Gateway.GAME_SOCKET_PATH, request.path)
        assertEquals("Bearer access", request.getHeader("Authorization"))
        assertEquals("websocket", request.getHeader("Upgrade"))
        assertEquals(
            Json.parseToJsonElement("""{"command":"ADD_TO_WAITING_LIST","category":"MUSIC","timeOptionId":"quick","players":3,"displayName":"Player"}"""),
            Json.parseToJsonElement(command.get(1, TimeUnit.SECONDS)),
        )
    }

    @Test
    fun closingAnAttemptDiscardsAMatchFrameRacingWithCancellation() {
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                webSocket.send("""{"success":true,"event":"ADDED_TO_WAITING_LIST"}""")
                webSocket.send("""{"success":true,"event":"MATCH_CREATED","metaData":{"gameId":"late-match"}}""")
            }
        }))

        runBlocking {
            val socket = Gateway(server.url("/")).openGame("access")
            try {
                withTimeout(5_000) {
                    socket.join(MatchSelection("SPORT", TimeOption("quick", "Quick", 10), 2), "Player")
                    assertEquals(MatchEvent.Queued, socket.events.receive())
                    socket.close()
                    assertTrue(socket.events.receiveCatching().isClosed)
                }
            } finally {
                socket.close()
            }
        }
    }

    @Test
    fun receivesQueueRejectionFirstEvenWhenTheServerAlsoSendsAnAcknowledgement() {
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                webSocket.send("""{"success":false,"event":"ERROR","message":"internal server error"}""")
                webSocket.send("""{"success":true,"event":"ADDED_TO_WAITING_LIST"}""")
            }
        }))

        runBlocking {
            val socket = Gateway(server.url("/")).openGame("access")
            try {
                withTimeout(5_000) {
                    socket.join(MatchSelection("SPORT", TimeOption("quick", "Quick", 10), 2), "Player")
                    assertEquals(MatchEvent.Rejected, socket.events.receive())
                }
            } finally {
                socket.close()
            }
        }
    }

    @Test
    fun readsExpirationWithoutTreatingAnUnreadableTokenAsATimestamp() {
        val payload = Base64.getUrlEncoder().withoutPadding().encodeToString("""{"exp":1234567890,"id":"5"}""".toByteArray())
        assertEquals(Instant.ofEpochSecond(1_234_567_890), tokenExpiration("header.$payload.signature"))
        assertNull(tokenExpiration("invalid"))
        assertNull(tokenExpiration("header.%.signature"))
    }

    private fun reply(body: String) {
        server.enqueue(MockResponse().setBody(body))
    }

    private fun nextRequest(): RecordedRequest {
        val request = server.takeRequest(3, TimeUnit.SECONDS)
        assertNotNull(request)
        return requireNotNull(request)
    }
}

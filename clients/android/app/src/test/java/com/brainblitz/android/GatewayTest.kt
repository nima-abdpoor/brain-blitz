package com.brainblitz.android

import kotlinx.coroutines.runBlocking
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertThrows
import org.junit.Rule
import org.junit.Test

class GatewayTest {
    @get:Rule
    val server = MockWebServer()

    @Test
    fun usesGatewayRoutesAndReadsTheUnenvelopedInitialFrame() {
        server.enqueue(MockResponse().setBody("""{"id":"5","accessToken":"access","refreshToken":"refresh"}"""))
        server.enqueue(MockResponse().setBody("""{"id":"5","displayName":"Player","statistics":null}"""))
        server.enqueue(MockResponse().setBody("""{"minPlayers":2,"maxPlayers":3,"numberOfPlayers":[2,3]}"""))
        server.enqueue(MockResponse().setBody("""[{"id":"quick","label":"Quick","seconds":10,"enabled":true}]"""))
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                webSocket.send("""{"categories":["MUSIC","SPORT"],"numberOfPlayers":[2,3]}""")
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                webSocket.close(1000, null)
            }
        }))

        val result = runBlocking { Gateway(server.url("/")).check("player@example.com", "example-password") }

        assertEquals(
            GatewayCheck("Player", listOf(2, 3), listOf(TimeOption("quick", "Quick", 10)), listOf("MUSIC", "SPORT")),
            result,
        )
        val login = server.takeRequest()
        assertEquals("/user-service/public/api/v1/login", login.path)
        assertEquals("POST", login.method)
        assertNull(login.getHeader("Authorization"))
        assertEquals("""{"email":"player@example.com","password":"example-password"}""", login.body.readUtf8())

        val protectedPaths = listOf(
            "/user-service/api/v1/profile",
            "/game-service/api/v1/game-options",
            "/question-service/api/v1/question/time-options",
            "/game-service/api/v1/process-game",
        )
        for (path in protectedPaths) {
            val request = server.takeRequest()
            assertEquals(path, request.path)
            assertEquals("GET", request.method)
            assertEquals("Bearer access", request.getHeader("Authorization"))
            if (path.endsWith("/process-game")) assertEquals("websocket", request.getHeader("Upgrade"))
        }
    }

    @Test
    fun rejectedLoginStopsTheCheck() {
        server.enqueue(MockResponse().setResponseCode(403))

        val failure = assertThrows(GatewayHttpException::class.java) {
            runBlocking { Gateway(server.url("/")).check("player@example.com", "wrong-password") }
        }

        assertEquals(403, failure.status)
        assertEquals("/user-service/public/api/v1/login", failure.endpoint)
        assertEquals(1, server.requestCount)
    }
}

package com.brainblitz.android

import java.time.Instant
import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withTimeout
import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Rule
import org.junit.Test
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer

class MatchRecoveryTest {
    @get:Rule
    val server = MockWebServer()

    @Test
    fun snapshotReplacesEarlierEventsAndKeepsTheAcceptedAnswerLocked() {
        val commands = LinkedBlockingQueue<String>()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                commands.add(text)
                webSocket.send("""
                    {"success":true,"event":"MATCH_CREATED","metaData":{"gameId":"g1"}}
                """)
                webSocket.send("""
                    {"success":true,"event":"MATCH_SNAPSHOT","metaData":{"snapshot":{
                        "gameId":"g1","players":2,"category":"TECH","timeOptionId":"quick","questionSeconds":10,
                        "phase":"round","roundNumber":2,"totalQuestions":3,
                        "question":{"id":"q2","content":"Which protocol?","choices":["HTTP","SMTP"],"ttl":"2026-11-22T08:31:00Z"},
                        "submission":{"choice":"HTTP","isCorrect":true}
                    }}}
                """)
                webSocket.send("""
                    {"success":true,"event":"QUESTIONS_PUBLISHED","metaData":{
                        "gameId":"g1","roundNumber":3,"totalQuestions":3,
                        "questions":[{"id":"q3","content":"New question","choices":["A","B"],"ttl":"2026-11-22T08:31:07Z"}]
                    }}
                """)
            }
        }))

        runBlocking {
            val socket = Gateway(server.url("/")).openGame("access")
            try {
                withTimeout(5_000) {
                    val snapshot = requireNotNull(socket.recover())
                    val progress = snapshot.progress()
                    val round = progress.phase as GamePhase.Round
                    assertEquals(2, round.number)
                    assertEquals(Instant.parse("2026-11-22T08:31:00Z"), round.question.deadline)
                    assertEquals(Submission("HTTP", AnswerFeedback("q2", true)), round.submission)
                    assertNull(round.submit("SMTP", round.question.deadline.minusSeconds(1)))
                    val next = progress.apply(socket.events.receive() as MatchEvent.Gameplay)
                    assertEquals(3, next.roundNumber)
                    assertNull((next.phase as GamePhase.Round).submission)
                }
            } finally {
                socket.close()
            }
        }

        assertEquals("""{"command":"SYNC_MATCH"}""", commands.poll(1, TimeUnit.SECONDS))
        assertNull(commands.poll())
        assertEquals("Bearer access", server.takeRequest(1, TimeUnit.SECONDS)?.getHeader("Authorization"))
    }

    @Test
    fun restoresReadinessAndFinalStandingsWithoutAnActiveQuestion() {
        val ready = Json.decodeFromString<MatchSnapshot>(
            """
                {"gameId":"g1","players":2,"category":"SPORT","timeOptionId":"quick","questionSeconds":10,
                "phase":"ready","ready":true,"readyUntil":"2026-11-22T08:31:00Z"}
            """,
        )
        assertEquals(GamePhase.Ready(true, Instant.parse("2026-11-22T08:31:00Z")), ready.progress().phase)

        val final = Json.decodeFromString<MatchSnapshot>(
            """
                {"gameId":"g1","players":2,"category":"SPORT","timeOptionId":"quick","questionSeconds":10,
                "phase":"finished","scores":[{"playerId":"2","point":0},{"playerId":"1","point":15}]}
            """,
        )
        assertEquals(GamePhase.Finished(listOf(PlayerScore("1", 15), PlayerScore("2", 0))), final.progress().phase)
    }

    @Test
    fun waitsForTheServerToAcknowledgeLeavingTheQueue() {
        val commands = LinkedBlockingQueue<String>()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                commands.add(text)
                if (text == """{"command":"LEAVE_QUEUE"}""") {
                    webSocket.send("""{"success":true,"event":"LEFT_WAITING_LIST"}""")
                } else {
                    webSocket.send("""{"success":true,"event":"ADDED_TO_WAITING_LIST"}""")
                }
            }
        }))

        runBlocking {
            val socket = Gateway(server.url("/")).openGame("access")
            try {
                withTimeout(5_000) {
                    socket.join(MatchSelection("SPORT", TimeOption("quick", "Quick", 10), 2), "Player")
                    assertEquals(MatchEvent.Queued, socket.events.receive())
                    socket.leaveQueue()
                    assertEquals(MatchEvent.LeftQueue, socket.events.receive())
                }
            } finally {
                socket.close()
            }
        }

        commands.poll(1, TimeUnit.SECONDS)
        assertEquals("""{"command":"LEAVE_QUEUE"}""", commands.poll(1, TimeUnit.SECONDS))
    }
}

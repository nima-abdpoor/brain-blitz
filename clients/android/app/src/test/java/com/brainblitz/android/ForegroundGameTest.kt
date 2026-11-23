package com.brainblitz.android

import java.time.Instant
import java.util.concurrent.CompletableFuture
import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withTimeout
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertSame
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test

class ForegroundGameTest {
    @get:Rule
    val server = MockWebServer()

    @Test
    fun handlesAnAnsweredRoundAndAnExpiredRoundBeforeFinalStandings() {
        val connection = CompletableFuture<WebSocket>()
        val commands = LinkedBlockingQueue<String>()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                connection.complete(webSocket)
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                commands.add(text)
                val command = Json.parseToJsonElement(text).jsonObject.getValue("command").jsonPrimitive.content
                when (command) {
                    "ADD_TO_WAITING_LIST" -> webSocket.send(MATCH_CREATED)
                    "READY" -> webSocket.send(QUESTION_ONE)
                    "ANSWER" -> {
                        webSocket.send(ACCEPTED_WITHOUT_FEEDBACK)
                        webSocket.send(ANSWER_ACCEPTED)
                    }
                }
            }
        }))

        runBlocking {
            val socket = Gateway(server.url("/")).openGame("access")
            try {
                withTimeout(5_000) {
                    socket.join(MatchSelection("TECH", TimeOption("quick", "Quick", 10), 2), "Player")
                    val created = socket.events.receive() as MatchEvent.Created
                    var game = GameProgress(created.gameId)
                    socket.ready(game.gameId)
                    val question = socket.events.receive() as MatchEvent.QuestionPublished
                    game = game.apply(question)
                    val active = game.phase as GamePhase.Round
                    assertEquals(2, active.totalQuestions)
                    assertEquals(deadline, active.question.deadline)

                    val submitted = requireNotNull(active.submit("Kotlin", deadline.minusMillis(1)))
                    game = game.copy(phase = submitted)
                    socket.answer(game.gameId, active.question.id, submitted.submission!!.choice)
                    val accepted = socket.events.receive() as MatchEvent.AnswerAccepted
                    game = game.apply(accepted)
                    assertEquals(AnswerFeedback("q1", true), (game.phase as GamePhase.Round).submission?.feedback)
                    assertNull(game.lastRecap)

                    val peer = connection.get(1, TimeUnit.SECONDS)
                    peer.send(FIRST_RECAP)
                    peer.send(QUESTION_TWO)
                    game = game.apply(socket.events.receive() as MatchEvent.RoundClosed)
                    assertEquals(GamePhase.BetweenRounds, game.phase)
                    assertEquals("Kotlin", game.lastRecap?.recap?.correctAnswer)
                    game = game.apply(socket.events.receive() as MatchEvent.QuestionPublished)
                    val expired = game.phase as GamePhase.Round
                    assertEquals(0L, expired.secondsRemaining(expired.question.deadline))
                    assertNull(expired.submit("HTTP", expired.question.deadline))

                    peer.send(EXPIRED_RECAP)
                    peer.send(COMPLETED)
                    peer.close(1000, null)
                    game = game.apply(socket.events.receive() as MatchEvent.RoundClosed)
                    assertFalse(requireNotNull(game.lastRecap).recap.answered)
                    game = game.apply(socket.events.receive() as MatchEvent.Completed)
                    assertEquals(
                        listOf(PlayerScore("2", 12, "Player"), PlayerScore("8", 0)),
                        (game.phase as GamePhase.Finished).scores,
                    )
                }
            } finally {
                socket.close()
            }
        }

        assertTrue(requireNotNull(commands.poll(1, TimeUnit.SECONDS)).contains("ADD_TO_WAITING_LIST"))
        assertCommand(commands, """{"command":"READY","gameId":"g1"}""")
        assertCommand(commands, """{"command":"ANSWER","answer":{"gameId":"g1","questionId":"q1","choice":"Kotlin"}}""")
        assertTrue(commands.isEmpty())
    }

    @Test
    fun ignoresStaleEventsAndAllowsOneValidChoiceBeforeTheAbsoluteDeadline() {
        val question = GameQuestion("q1", "Question", listOf("A", "B"), deadline)
        val published = MatchEvent.QuestionPublished("g1", 1, 2, question)
        val game = GameProgress("g1").apply(published)
        val round = game.phase as GamePhase.Round
        assertEquals(2L, round.secondsRemaining(deadline.minusMillis(1_001)))
        assertEquals(1L, round.secondsRemaining(deadline.minusNanos(1)))
        assertEquals(0L, round.secondsRemaining(deadline.plusMillis(1)))
        assertNull(round.submit("C", deadline.minusSeconds(1)))
        assertNull(round.submit("A", deadline))

        val submitted = requireNotNull(round.submit("B", deadline.minusNanos(1)))
        assertNull(submitted.submit("A", deadline.minusNanos(1)))
        val waiting = game.copy(phase = submitted)
        assertSame(waiting, waiting.apply(published))
        assertSame(waiting, waiting.apply(MatchEvent.AnswerAccepted("other", AnswerFeedback("q1", true))))
        assertSame(waiting, waiting.apply(MatchEvent.AnswerAccepted("g1", AnswerFeedback("old-question", true))))
        val confirmed = waiting.apply(MatchEvent.AnswerAccepted("g1", AnswerFeedback("q1", false)))
        assertEquals(false, (confirmed.phase as GamePhase.Round).submission?.feedback?.isCorrect)
        val tie = listOf(PlayerScore("2", 0), PlayerScore("1", 0))
        val finished = confirmed.apply(MatchEvent.Completed("g1", tie))
        assertEquals(tie.reversed(), (finished.phase as GamePhase.Finished).scores)
        assertSame(finished, finished.apply(published.copy(roundNumber = 2)))
    }

    private fun assertCommand(commands: LinkedBlockingQueue<String>, expected: String) {
        val actual = requireNotNull(commands.poll(1, TimeUnit.SECONDS))
        assertEquals(Json.parseToJsonElement(expected), Json.parseToJsonElement(actual))
    }

    private companion object {
        val deadline: Instant = Instant.parse("2026-11-22T07:30:05.123456789Z")
        const val MATCH_CREATED = """
            {"success":true,"event":"MATCH_CREATED","metaData":{
                "gameId":"g1","questions":null,"leaderBoard":{"gameId":"","playerPoint":null}
            }}
        """
        const val QUESTION_ONE = """
            {"success":true,"event":"QUESTIONS_PUBLISHED","metaData":{
                "gameId":"g1","roundNumber":1,"totalQuestions":2,"questions":[{
                    "id":"q1","content":"Which language?","choices":["Kotlin","Rust"],
                    "difficulty":"EASY","ttl":"2026-11-22T07:30:05.123456789Z"
                }]
            }}
        """
        const val ANSWER_ACCEPTED = """
            {"success":true,"event":"ANSWER_ACCEPTED","metaData":{
                "gameId":"g1","questions":null,"answerFeedback":{"questionId":"q1","isCorrect":true},
                "leaderBoard":{"gameId":"g1","playerPoint":[{"playerId":"2","point":12,"answers":null}]}
            }}
        """
        const val ACCEPTED_WITHOUT_FEEDBACK = """
            {"success":true,"event":"ANSWER_ACCEPTED","metaData":{"gameId":"g1","questions":null}}
        """
        const val FIRST_RECAP = """
            {"success":true,"event":"ROUND_CLOSED","metaData":{
                "gameId":"g1","roundNumber":1,"recap":{
                    "questionId":"q1","question":"Which language?","playerAnswer":"Kotlin",
                    "correctAnswer":"Kotlin","answered":true,"isCorrect":true
                }
            }}
        """
        const val QUESTION_TWO = """
            {"success":true,"event":"QUESTIONS_PUBLISHED","metaData":{
                "gameId":"g1","roundNumber":2,"totalQuestions":2,"questions":[{
                    "id":"q2","content":"Which protocol?","choices":["SSH","HTTP"],"ttl":"2026-11-22T07:30:15Z"
                }]
            }}
        """
        const val EXPIRED_RECAP = """
            {"success":true,"event":"ROUND_CLOSED","metaData":{
                "gameId":"g1","roundNumber":2,"recap":{
                    "questionId":"q2","question":"Which protocol?","playerAnswer":"",
                    "correctAnswer":"HTTP","answered":false,"isCorrect":false
                }
            }}
        """
        const val COMPLETED = """
            {"success":true,"event":"COMPLETED","metaData":{"gameId":"g1","leaderBoard":{
                "gameId":"g1","playerPoint":[
                    {"playerId":"8","point":0,"answers":[]},
                    {"playerId":"2","displayName":"Player","point":12,"answers":[{
                        "questionId":"q1","playerAnswer":"Kotlin","correctAnswer":"Kotlin","isCorrect":true
                    }]}
                ]
            }}}
        """
    }
}

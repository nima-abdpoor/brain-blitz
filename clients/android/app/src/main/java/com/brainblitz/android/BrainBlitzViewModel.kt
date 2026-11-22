package com.brainblitz.android

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import java.io.IOException
import java.security.GeneralSecurityException
import java.security.ProviderException
import java.time.Duration
import java.time.Instant
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull
import kotlinx.serialization.SerializationException
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull

internal enum class AccountMode { SignIn, SignUp }

internal data class QueueEntry(val selection: MatchSelection, val acknowledged: Boolean = false)

internal sealed interface AppScreen {
    data class Account(
        val mode: AccountMode = AccountMode.SignIn,
        val busy: Boolean = false,
        val notice: Int? = null,
    ) : AppScreen

    data class ProfilePage(
        val profile: Profile? = null,
        val busy: Boolean = true,
        val notice: Int? = null,
    ) : AppScreen

    data class Play(
        val profile: Profile,
        val categories: List<String>,
        val options: MatchOptions? = null,
        val busy: Boolean = true,
        val queue: QueueEntry? = null,
        val notice: Int? = null,
        val previousSelection: MatchSelection? = null,
    ) : AppScreen

    data class Match(
        val profile: Profile,
        val selection: MatchSelection,
        val game: GameProgress,
        val notice: Int? = null,
        val disconnected: Boolean = false,
    ) : AppScreen
}

internal class BrainBlitzViewModel(application: Application) : AndroidViewModel(application) {
    private val origin = BuildConfig.GATEWAY_URL.toHttpUrlOrNull()
    private val gateway = origin?.let { Gateway(it) }
    private val sessionStore = origin?.let { SessionStore(application, it.toString()) }
    private val mutableScreen = MutableStateFlow<AppScreen>(AppScreen.Account(busy = true))
    val screen = mutableScreen.asStateFlow()
    val configured = gateway != null

    private var accessToken: String? = null
    private var requestJob: Job? = null
    private var expiryJob: Job? = null
    private var socket: GameSocket? = null
    private var gameJob: Job? = null

    // Rejoining accounts may receive no initial frame. Use the server's known categories until it sends a list.
    private var categories = listOf("SPORT", "MUSIC", "TECH")

    init {
        if (sessionStore == null) {
            mutableScreen.value = AppScreen.Account(notice = R.string.missing_gateway)
        } else {
            runRequest(onError = { showAccountError(AccountMode.SignIn, it) }) {
                val stored = sessionStore.read()
                if (stored == null) {
                    mutableScreen.value = AppScreen.Account()
                } else {
                    openSession(stored)
                }
            }
        }
    }

    fun changeAccountMode() {
        val account = screen.value as? AppScreen.Account ?: return
        if (account.busy) return
        val mode = if (account.mode == AccountMode.SignIn) AccountMode.SignUp else AccountMode.SignIn
        mutableScreen.value = AppScreen.Account(mode)
    }

    fun submitAccount(email: String, password: String) {
        val gateway = gateway ?: return
        val store = sessionStore ?: return
        val account = screen.value as? AppScreen.Account ?: return
        if (account.busy || email.isBlank() || password.isEmpty()) return

        var mode = account.mode
        if (mode == AccountMode.SignUp && password.toByteArray(Charsets.UTF_8).size > 72) {
            mutableScreen.value = account.copy(notice = R.string.password_too_long)
            return
        }
        mutableScreen.value = account.copy(busy = true, notice = null)
        runRequest(onError = { showAccountError(mode, it) }) {
            if (mode == AccountMode.SignUp) {
                gateway.signUp(email, password)
                mode = AccountMode.SignIn
            }
            val token = gateway.login(email, password)
            store.write(token)
            openSession(token)
        }
    }

    private fun showAccountError(mode: AccountMode, message: Int) {
        mutableScreen.value = if (accessToken == null) {
            AppScreen.Account(mode, notice = message)
        } else {
            AppScreen.ProfilePage(busy = false, notice = message)
        }
    }

    private suspend fun openSession(token: String) {
        val expiration = tokenExpiration(token)
        if (expiration != null && expiration <= Instant.now()) {
            endSession(R.string.session_expired)
            return
        }
        accessToken = token
        expiryJob?.cancel()
        if (expiration != null) {
            expiryJob = viewModelScope.launch {
                while (Instant.now() < expiration) {
                    val remaining = Duration.between(Instant.now(), expiration)
                    // Recheck periodically so device clock changes do not leave a stale session open.
                    delay(remaining.coerceAtMost(Duration.ofMinutes(1)).toMillis().coerceAtLeast(1))
                }
                endSession(R.string.session_expired)
            }
        }
        mutableScreen.value = AppScreen.ProfilePage()
        fetchProfile(token)
    }

    fun retryProfile() {
        val page = screen.value as? AppScreen.ProfilePage ?: return
        if (page.busy) return
        val token = activeToken() ?: return
        mutableScreen.value = page.copy(busy = true, notice = null)
        runRequest(onError = { mutableScreen.value = page.copy(busy = false, notice = it) }) {
            fetchProfile(token)
        }
    }

    private suspend fun fetchProfile(token: String) {
        val profile = gateway?.profile(token) ?: return
        mutableScreen.value = AppScreen.ProfilePage(profile, busy = false)
    }

    fun openPlay() {
        val profile = (screen.value as? AppScreen.ProfilePage)?.profile ?: return
        loadOptions(AppScreen.Play(profile, categories))
    }

    fun retryOptions() {
        val play = screen.value as? AppScreen.Play ?: return
        if (!play.busy && play.queue == null) loadOptions(play)
    }

    private fun loadOptions(play: AppScreen.Play) {
        val gateway = gateway ?: return
        val token = activeToken() ?: return
        mutableScreen.value = play.copy(busy = true, notice = null)
        runRequest(onError = { mutableScreen.value = play.copy(busy = false, notice = it) }) {
            mutableScreen.value = play.copy(options = gateway.matchOptions(token), busy = false, notice = null)
        }
    }

    fun joinQueue(selection: MatchSelection) {
        val gateway = gateway ?: return
        val play = screen.value as? AppScreen.Play ?: return
        val options = play.options ?: return
        if (play.queue != null || play.busy || selection.category !in play.categories ||
            selection.players !in options.playerCounts || selection.timeOption !in options.timeOptions
        ) return

        val token = activeToken() ?: return
        mutableScreen.value = play.copy(queue = QueueEntry(selection), notice = null)
        val game = gateway.openGame(token)
        socket = game
        gameJob = viewModelScope.launch {
            try {
                val acknowledged = withTimeoutOrNull(12_000) {
                    game.join(selection, play.profile.displayName)
                    while ((screen.value as? AppScreen.Play)?.queue?.acknowledged == false) {
                        handleEvent(game, game.events.receive())
                    }
                    true
                }
                if (acknowledged == null) {
                    showSocketError(R.string.queue_timeout)
                } else {
                    for (event in game.events) {
                        handleEvent(game, event)
                        if ((screen.value as? AppScreen.Match)?.game?.phase is GamePhase.Finished) break
                    }
                }
            } catch (error: GatewayHttpException) {
                if (socket === game && error.sessionRejected) endSession(R.string.session_expired)
                else if (socket === game) showSocketError(R.string.connection_error)
            } catch (_: QueueRejected) {
                if (socket === game) showSocketError(R.string.queue_rejected)
            } catch (_: IOException) {
                if (socket === game) showSocketError(R.string.game_connection_lost)
            } finally {
                if (socket === game) {
                    socket = null
                    gameJob = null
                    game.close()
                }
            }
        }
    }

    private fun handleEvent(game: GameSocket, event: MatchEvent) {
        if (socket !== game) return
        when (val current = screen.value) {
            is AppScreen.Play -> handleQueueEvent(current, event)
            is AppScreen.Match -> handleMatchEvent(current, event)
            else -> Unit
        }
    }

    private fun handleQueueEvent(play: AppScreen.Play, event: MatchEvent) {
        val queue = play.queue ?: return
        mutableScreen.value = when (event) {
            is MatchEvent.Categories -> {
                categories = event.names
                play.copy(categories = categories)
            }
            MatchEvent.Queued -> play.copy(queue = queue.copy(acknowledged = true))
            MatchEvent.Rejected -> throw QueueRejected()
            is MatchEvent.Created -> AppScreen.Match(play.profile, queue.selection, GameProgress(event.gameId))
            is MatchEvent.Gameplay -> play
        }
    }

    private fun handleMatchEvent(match: AppScreen.Match, event: MatchEvent) {
        when (event) {
            is MatchEvent.Gameplay -> {
                val game = match.game.apply(event)
                if (game === match.game) return
                val disconnected = match.disconnected && game.phase !is GamePhase.Finished
                mutableScreen.value = match.copy(
                    game = game,
                    notice = if (disconnected) match.notice else null,
                    disconnected = disconnected,
                )
            }
            MatchEvent.Rejected -> {
                if (match.disconnected) return
                val ready = match.game.phase as? GamePhase.Ready
                mutableScreen.value = if (ready != null) {
                    match.copy(game = match.game.copy(phase = GamePhase.Ready()), notice = R.string.ready_rejected)
                } else {
                    // ERROR has no round or command ID. Keep the answer locked and wait for the recap.
                    match.copy(notice = R.string.game_command_rejected)
                }
            }
            else -> Unit
        }
    }

    private fun showSocketError(message: Int) {
        when (val current = screen.value) {
            is AppScreen.Play -> mutableScreen.value = current.copy(queue = null, notice = message)
            is AppScreen.Match -> mutableScreen.value = current.copy(notice = message, disconnected = true)
            else -> Unit
        }
    }

    fun ready() {
        val match = screen.value as? AppScreen.Match ?: return
        val ready = match.game.phase as? GamePhase.Ready ?: return
        if (ready.waiting) return
        sendGameCommand(match, ready.copy(waiting = true)) { it.ready(match.game.gameId) }
    }

    fun answer(roundNumber: Int, choice: String) {
        val match = screen.value as? AppScreen.Match ?: return
        val round = match.game.phase as? GamePhase.Round ?: return
        if (round.number != roundNumber) return
        val submitted = round.submit(choice, Instant.now()) ?: return
        sendGameCommand(match, submitted) { it.answer(match.game.gameId, round.question.id, choice) }
    }

    private fun sendGameCommand(match: AppScreen.Match, phase: GamePhase, send: (GameSocket) -> Unit) {
        val game = socket ?: return
        if (match.disconnected || activeToken() == null) return
        mutableScreen.value = match.copy(game = match.game.copy(phase = phase), notice = null)
        try {
            send(game)
        } catch (_: IOException) {
            showSocketError(R.string.game_connection_lost)
        }
    }

    fun playAgain() {
        val match = screen.value as? AppScreen.Match ?: return
        if (match.game.phase !is GamePhase.Finished) return
        closeGame()
        loadOptions(AppScreen.Play(match.profile, categories, previousSelection = match.selection))
    }

    fun cancelQueue() {
        val play = screen.value as? AppScreen.Play ?: return
        if (play.queue == null) return
        closeGame()
        mutableScreen.value = play.copy(queue = null)
    }

    fun returnToProfile() {
        val profile = when (val current = screen.value) {
            is AppScreen.Play -> current.profile
            is AppScreen.Match -> current.profile
            else -> return
        }
        requestJob?.cancel()
        closeGame()
        mutableScreen.value = AppScreen.ProfilePage(profile, busy = false)
    }

    fun signOut() {
        endSession()
    }

    private fun activeToken(): String? {
        val token = accessToken ?: return null
        val expiration = tokenExpiration(token)
        if (expiration == null || expiration > Instant.now()) return token
        endSession(R.string.session_expired)
        return null
    }

    private fun endSession(notice: Int? = null) {
        requestJob?.cancel()
        expiryJob?.cancel()
        closeGame()
        accessToken = null
        mutableScreen.value = AppScreen.Account(busy = true)
        runRequest(onError = { mutableScreen.value = AppScreen.Account(notice = R.string.storage_error) }) {
            sessionStore?.clear()
            mutableScreen.value = AppScreen.Account(notice = notice)
        }
    }

    private fun closeGame() {
        val game = socket
        socket = null
        gameJob?.cancel()
        gameJob = null
        game?.close()
    }

    private fun runRequest(onError: (Int) -> Unit, request: suspend () -> Unit) {
        requestJob = viewModelScope.launch {
            try {
                request()
            } catch (error: GatewayHttpException) {
                if (error.sessionRejected) endSession(R.string.session_expired)
                else onError(httpMessage(error))
            } catch (_: SerializationException) {
                onError(R.string.invalid_response)
            } catch (_: GeneralSecurityException) {
                onError(R.string.storage_error)
            } catch (_: ProviderException) {
                onError(R.string.storage_error)
            } catch (_: IOException) {
                onError(R.string.connection_error)
            }
        }
    }

    private fun httpMessage(error: GatewayHttpException): Int = when {
        error.duplicateAccount -> R.string.duplicate_account
        error.endpoint == Gateway.LOGIN_PATH && error.status in 400..499 -> R.string.invalid_login
        error.endpoint == Gateway.SIGNUP_PATH && error.status in 400..499 -> R.string.invalid_signup
        else -> R.string.gateway_error
    }

    override fun onCleared() {
        closeGame()
    }

    private class QueueRejected : IOException()
}

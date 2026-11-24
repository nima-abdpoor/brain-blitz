package com.brainblitz.android

import android.app.Application
import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import java.io.IOException
import java.security.GeneralSecurityException
import java.security.MessageDigest
import java.security.ProviderException
import java.time.Duration
import java.time.Instant
import java.util.Base64
import kotlin.time.TimeSource
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.TimeoutCancellationException
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull
import kotlinx.coroutines.withTimeout
import kotlinx.coroutines.withContext
import kotlinx.serialization.SerializationException
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull

internal enum class AccountMode { SignIn, SignUp }

internal data class QueueEntry(
    val selection: MatchSelection,
    val acknowledged: Boolean = false,
    val cancelling: Boolean = false,
)

internal enum class RecoveryStatus { Connecting, Retry }

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
        val recovery: RecoveryStatus? = null,
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
        val recovery: RecoveryStatus? = null,
    ) : AppScreen
}

internal class BrainBlitzViewModel(application: Application) : AndroidViewModel(application) {
    private val origin = BuildConfig.GATEWAY_URL.toHttpUrlOrNull()
    private val gateway = origin?.let { Gateway(it) }
    private val sessionStore = origin?.let { SessionStore(application, it.toString()) }
    private val dismissedResults = application.getSharedPreferences("match-results", Context.MODE_PRIVATE)
    private val connectivity = application.getSystemService(ConnectivityManager::class.java)
    private val mutableScreen = MutableStateFlow<AppScreen>(AppScreen.Account(busy = true))
    val screen = mutableScreen.asStateFlow()
    val configured = gateway != null

    private var accessToken: String? = null
    private var requestJob: Job? = null
    private var expiryJob: Job? = null
    private var socket: GameSocket? = null
    private var gameJob: Job? = null
    private var cancelQueueJob: Job? = null
    private var dismissResultJob: Job? = null
    private var recoveryRequest: RecoveryRequest? = null
    private var foreground = false
    private var network = connectivity?.activeNetwork
    private var supportsRecovery = false

    private val networkCallback = object : ConnectivityManager.NetworkCallback() {
        override fun onAvailable(network: Network) {
            viewModelScope.launch { changeNetwork(network) }
        }

        override fun onLost(network: Network) {
            viewModelScope.launch {
                if (this@BrainBlitzViewModel.network == network) changeNetwork(null)
            }
        }
    }

    // Rejoining accounts may receive no initial frame. Use the server's known categories until it sends a list.
    private var categories = listOf("SPORT", "MUSIC", "TECH", "GEOGRAPHY", "SCIENCE", "HISTORY")

    init {
        connectivity?.registerDefaultNetworkCallback(networkCallback)
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
        startRecovery(RecoveryRequest(profile))
    }

    fun openPlay() {
        val page = screen.value as? AppScreen.ProfilePage ?: return
        val profile = page.profile ?: return
        if (page.recovery == RecoveryStatus.Connecting) return
        startRecovery(RecoveryRequest(profile, openPlay = true))
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
        val game = gateway.openGame(token, heartbeat = supportsRecovery)
        socket = game
        gameJob = viewModelScope.launch {
            var recovery: RecoveryRequest? = null
            try {
                val acknowledged = withTimeoutOrNull(12_000) {
                    game.join(selection, play.profile.displayName)
                    while ((screen.value as? AppScreen.Play)?.queue?.acknowledged == false) {
                        handleEvent(game, game.events.receive())
                    }
                    true
                }
                if (acknowledged == null) {
                    recovery = recoveryForCurrentScreen(R.string.queue_timeout)
                } else {
                    receiveGameEvents(game)
                }
            } catch (error: GatewayHttpException) {
                if (socket === game && error.sessionRejected) endSession(R.string.session_expired)
                else recovery = recoveryForCurrentScreen(R.string.connection_error)
            } catch (_: QueueRejected) {
                val current = screen.value as? AppScreen.Play
                if (current?.queue?.cancelling == true) {
                    recovery = recoveryForCurrentScreen(null)
                } else if (current != null && socket === game) {
                    mutableScreen.value = current.copy(queue = null, notice = R.string.queue_rejected)
                }
            } catch (_: IOException) {
                recovery = recoveryForCurrentScreen(R.string.game_connection_lost)
            } finally {
                if (socket === game) {
                    socket = null
                    gameJob = null
                    cancelQueueJob?.cancel()
                    game.close()
                    recovery?.let(::startRecovery)
                }
            }
        }
    }

    private suspend fun receiveGameEvents(game: GameSocket) {
        for (event in game.events) {
            handleEvent(game, event)
            when (val current = screen.value) {
                is AppScreen.Match -> if (current.game.terminal) return
                is AppScreen.Play -> if (current.queue == null) return
                else -> return
            }
        }
        throw IOException("Game connection closed")
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
            MatchEvent.Queued -> {
                if (queue.cancelling) requestQueueExit()
                play.copy(queue = queue.copy(acknowledged = true))
            }
            MatchEvent.LeftQueue -> play.copy(queue = null, previousSelection = queue.selection)
            MatchEvent.Rejected -> throw QueueRejected()
            is MatchEvent.Created -> AppScreen.Match(
                play.profile, queue.selection, GameProgress(event.gameId, GamePhase.Ready(deadline = event.readyUntil)),
            )
            else -> play
        }
    }

    private fun handleMatchEvent(match: AppScreen.Match, event: MatchEvent) {
        when (event) {
            is MatchEvent.Gameplay -> {
                val game = match.game.apply(event)
                if (game === match.game) return
                mutableScreen.value = match.copy(game = game, notice = null, recovery = null)
            }
            MatchEvent.Rejected -> {
                val ready = match.game.phase as? GamePhase.Ready
                mutableScreen.value = if (ready != null) {
                    match.copy(game = match.game.copy(phase = ready.copy(waiting = false)), notice = R.string.ready_rejected)
                } else {
                    // ERROR has no round or command ID. Keep the answer locked and wait for the recap.
                    match.copy(notice = R.string.game_command_rejected)
                }
            }
            else -> Unit
        }
    }

    private fun recoveryForCurrentScreen(notice: Int?): RecoveryRequest? = when (val current = screen.value) {
        is AppScreen.Play -> current.queue?.let { queue ->
            RecoveryRequest(current.profile, fallback = current.copy(
                queue = null, notice = if (queue.cancelling) null else notice, previousSelection = queue.selection,
            ))
        }
        is AppScreen.Match -> if (current.game.terminal) null else RecoveryRequest(current.profile, matchId = current.game.gameId)
        else -> null
    }

    fun setForeground(visible: Boolean) {
        if (foreground == visible) return
        foreground = visible
        if (visible) reconnectIfNeeded()
    }

    private fun changeNetwork(current: Network?) {
        if (network == current) return
        network = current
        val request = recoveryForCurrentScreen(R.string.game_connection_lost) ?: recoveryRequest
        if (request != null) startRecovery(request) else reconnectIfNeeded()
    }

    fun retryRecovery() {
        recoveryRequest?.let(::startRecovery)
    }

    private val canConnect: Boolean
        get() = foreground && (connectivity == null || network != null)

    private fun reconnectIfNeeded() {
        if (!canConnect || gameJob?.isActive == true) return
        val request = recoveryRequest
        if (request != null) {
            resumeRecovery(request)
        } else {
            val page = screen.value as? AppScreen.ProfilePage
            if (page?.profile == null && page?.notice == R.string.connection_error) retryProfile()
        }
    }

    private fun startRecovery(request: RecoveryRequest) {
        closeGame()
        recoveryRequest = request
        showRecovery(request, RecoveryStatus.Connecting)
        reconnectIfNeeded()
    }

    private fun resumeRecovery(request: RecoveryRequest) {
        val gateway = gateway ?: return
        val token = activeToken() ?: return
        showRecovery(request, RecoveryStatus.Connecting)
        gameJob = viewModelScope.launch {
            var attempts = 0
            val backoff = listOf(0L, 500L, 1_000L, 2_000L)
            while (attempts < backoff.size && canConnect) {
                delay(backoff[attempts++])
                if (!canConnect) break
                val game = gateway.openGame(token, heartbeat = true)
                socket = game
                var connectedAt: TimeSource.Monotonic.ValueTimeMark? = null
                try {
                    val snapshot = withTimeout(5_000) { game.recover { categories = it } }
                    supportsRecovery = true
                    if (snapshot == null) {
                        finishDiscovery(request)
                        return@launch
                    }
                    val progress = snapshot.progress()
                    if (progress.phase is GamePhase.Finished && progress.gameId == dismissedResult(request.profile)) {
                        finishDiscovery(request)
                        return@launch
                    }
                    connectedAt = TimeSource.Monotonic.markNow()
                    val current = screen.value as? AppScreen.Match
                    val selection = if (current?.game?.gameId == progress.gameId) current.selection else snapshot.selection()
                    mutableScreen.value = AppScreen.Match(request.profile, selection, progress)
                    if (!progress.terminal) receiveGameEvents(game)
                    recoveryRequest = null
                    return@launch
                } catch (error: GatewayHttpException) {
                    if (error.sessionRejected) {
                        endSession(R.string.session_expired)
                        return@launch
                    }
                } catch (_: SnapshotUnsupported) {
                    supportsRecovery = false
                    if (request.matchId == null && screen.value !is AppScreen.Match) finishDiscovery(request)
                    else showRecovery(request, RecoveryStatus.Retry)
                    return@launch
                } catch (_: TimeoutCancellationException) {
                    // A failed handshake is retried with the same bounds as an I/O error.
                } catch (_: IOException) {
                    // A replacement socket must establish a fresh snapshot.
                } catch (_: IllegalArgumentException) {
                    // An invalid snapshot must never reopen an answer.
                } finally {
                    if (socket === game) socket = null
                    game.close()
                }
                if (connectedAt?.elapsedNow()?.inWholeSeconds?.let { it >= 10 } == true) attempts = 0
                showRecovery(request, RecoveryStatus.Connecting)
            }
            if (canConnect) showRecovery(request, RecoveryStatus.Retry)
        }
    }

    private fun showRecovery(request: RecoveryRequest, status: RecoveryStatus) {
        val current = screen.value as? AppScreen.Match
        mutableScreen.value = current?.copy(recovery = status, notice = null)
            ?: AppScreen.ProfilePage(request.profile, busy = false, recovery = status)
    }

    private fun finishDiscovery(request: RecoveryRequest) {
        recoveryRequest = null
        val current = screen.value as? AppScreen.Match
        val ready = current?.game?.phase as? GamePhase.Ready
        when {
            current != null && ready?.deadline?.let { Instant.now() >= it } == true -> {
                mutableScreen.value = current.copy(game = current.game.copy(phase = GamePhase.Cancelled), recovery = null)
            }
            request.fallback != null && current == null -> mutableScreen.value = request.fallback.copy(categories = categories)
            request.openPlay && current == null -> loadOptions(AppScreen.Play(request.profile, categories))
            else -> mutableScreen.value = AppScreen.ProfilePage(
                request.profile, busy = false, notice = if (current == null) null else R.string.match_unavailable,
            )
        }
    }

    private fun resultKey(profile: Profile): String {
        val account = "$origin\n${profile.username}".toByteArray(Charsets.UTF_8)
        return Base64.getUrlEncoder().encodeToString(MessageDigest.getInstance("SHA-256").digest(account))
    }

    private fun dismissedResult(profile: Profile): String? = dismissedResults.getString(resultKey(profile), null)

    private fun leaveMatch(match: AppScreen.Match, navigate: () -> Unit) {
        if (match.game.phase !is GamePhase.Finished) {
            navigate()
            return
        }
        if (dismissResultJob?.isActive == true) return
        dismissResultJob = viewModelScope.launch {
            val saved = withContext(Dispatchers.IO) {
                dismissedResults.edit().putString(resultKey(match.profile), match.game.gameId).commit()
            }
            if (screen.value !== match) return@launch
            if (saved) navigate()
            else mutableScreen.value = match.copy(notice = R.string.result_dismissal_error)
        }
    }

    fun ready() {
        val match = screen.value as? AppScreen.Match ?: return
        val ready = match.game.phase as? GamePhase.Ready ?: return
        if (ready.waiting || ready.deadline?.let { Instant.now() >= it } == true) return
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
        if (match.recovery != null || activeToken() == null) return
        mutableScreen.value = match.copy(game = match.game.copy(phase = phase), notice = null)
        try {
            send(game)
        } catch (_: IOException) {
            startRecovery(RecoveryRequest(match.profile, matchId = match.game.gameId))
        }
    }

    fun playAgain() {
        val match = screen.value as? AppScreen.Match ?: return
        if (!match.game.terminal) return
        leaveMatch(match) {
            closeGame()
            loadOptions(AppScreen.Play(match.profile, categories, previousSelection = match.selection))
        }
    }

    fun cancelQueue() {
        val play = screen.value as? AppScreen.Play ?: return
        val queue = play.queue ?: return
        if (queue.cancelling) return
        mutableScreen.value = play.copy(queue = queue.copy(cancelling = true))
        if (queue.acknowledged) {
            try {
                requestQueueExit()
            } catch (_: IOException) {
                recoveryForCurrentScreen(null)?.let(::startRecovery)
            }
        }
    }

    private fun requestQueueExit() {
        val game = socket ?: throw IOException("Queue connection closed")
        game.leaveQueue()
        cancelQueueJob = viewModelScope.launch {
            delay(4_000)
            val queue = (screen.value as? AppScreen.Play)?.queue
            if (socket === game && queue?.cancelling == true) {
                // Leave this watchdog before replacing the connection it manages.
                cancelQueueJob = null
                recoveryForCurrentScreen(null)?.let(::startRecovery)
            }
        }
    }

    fun returnToProfile() {
        when (val current = screen.value) {
            is AppScreen.Play -> showProfile(current.profile)
            is AppScreen.Match -> leaveMatch(current) { showProfile(current.profile) }
            else -> Unit
        }
    }

    private fun showProfile(profile: Profile) {
        requestJob?.cancel()
        recoveryRequest = null
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
        dismissResultJob?.cancel()
        recoveryRequest = null
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
        cancelQueueJob?.cancel()
        cancelQueueJob = null
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
        connectivity?.unregisterNetworkCallback(networkCallback)
        closeGame()
    }

    private class QueueRejected : IOException()

    private data class RecoveryRequest(
        val profile: Profile,
        val fallback: AppScreen.Play? = null,
        val openPlay: Boolean = false,
        val matchId: String? = null,
    )
}

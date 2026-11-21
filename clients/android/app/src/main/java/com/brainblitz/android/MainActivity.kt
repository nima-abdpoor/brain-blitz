package com.brainblitz.android

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.BackHandler
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.key
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.LiveRegionMode
import androidx.compose.ui.semantics.liveRegion
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.compose.collectAsStateWithLifecycle

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        val model = ViewModelProvider(this)[BrainBlitzViewModel::class.java]
        setContent {
            val colors = if (isSystemInDarkTheme()) darkColorScheme() else lightColorScheme()
            val screen by model.screen.collectAsStateWithLifecycle()
            MaterialTheme(colorScheme = colors) {
                BrainBlitzApp(screen, model)
            }
        }
    }
}

@Composable
private fun BrainBlitzApp(screen: AppScreen, model: BrainBlitzViewModel) {
    BackHandler(enabled = screen is AppScreen.Play || screen is AppScreen.Match) {
        if (screen is AppScreen.Play && screen.queue != null) model.cancelQueue()
        else model.returnToProfile()
    }

    Surface(Modifier.fillMaxSize()) {
        key(screen::class) {
            Column(
                modifier = Modifier
                    .safeDrawingPadding()
                    .imePadding()
                    .verticalScroll(rememberScrollState())
                    .padding(24.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                when (screen) {
                    is AppScreen.Account -> AccountForm(screen, model.configured, model::submitAccount, model::changeAccountMode)
                    is AppScreen.ProfilePage -> ProfilePage(screen, model::openPlay, model::retryProfile, model::signOut)
                    is AppScreen.Play -> PlayPage(screen, model::joinQueue, model::retryOptions, model::cancelQueue, model::returnToProfile)
                    is AppScreen.Match -> MatchPage(screen, model::returnToProfile)
                }
            }
        }
    }
}

@Composable
private fun AccountForm(
    screen: AppScreen.Account,
    configured: Boolean,
    onSubmit: (String, String) -> Unit,
    onSwitchMode: () -> Unit,
) {
    var email by rememberSaveable { mutableStateOf("") }
    var password by remember(screen.mode) { mutableStateOf("") }
    val focusManager = LocalFocusManager.current
    val enabled = configured && !screen.busy
    val signingUp = screen.mode == AccountMode.SignUp

    Heading(stringResource(R.string.app_name))
    OutlinedTextField(
        value = email,
        onValueChange = { email = it },
        label = { Text(stringResource(R.string.email)) },
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email),
        singleLine = true,
        enabled = enabled,
        modifier = Modifier.fillMaxWidth(),
    )
    OutlinedTextField(
        value = password,
        onValueChange = { password = it },
        label = { Text(stringResource(R.string.password)) },
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
        visualTransformation = PasswordVisualTransformation(),
        singleLine = true,
        enabled = enabled,
        modifier = Modifier.fillMaxWidth(),
    )
    Button(
        enabled = enabled && email.isNotBlank() && password.isNotEmpty(),
        onClick = {
            focusManager.clearFocus()
            onSubmit(email.trim(), password)
            password = ""
        },
    ) {
        Text(stringResource(if (signingUp) R.string.create_account else R.string.sign_in))
    }
    val alternative = if (signingUp) R.string.use_existing_account else R.string.create_account
    TextButton(enabled = enabled, onClick = onSwitchMode) {
        Text(stringResource(alternative))
    }
    if (screen.busy) Status(R.string.working)
    Notice(screen.notice)
}

@Composable
private fun ProfilePage(
    screen: AppScreen.ProfilePage,
    onPlay: () -> Unit,
    onRetry: () -> Unit,
    onSignOut: () -> Unit,
) {
    Heading(stringResource(R.string.profile))
    val profile = screen.profile
    if (profile != null) {
        Text(profile.displayName, style = MaterialTheme.typography.titleLarge)
        Text(profile.username, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Button(onClick = onPlay) { Text(stringResource(R.string.play)) }
        ProfileStatistics(profile.statistics)
    } else if (screen.busy) {
        Status(R.string.loading_profile)
    } else {
        Button(onClick = onRetry) { Text(stringResource(R.string.try_again)) }
    }
    Notice(screen.notice)
    OutlinedButton(onClick = onSignOut) { Text(stringResource(R.string.sign_out)) }
}

@Composable
private fun ProfileStatistics(stats: PlayerStatistics?) {
    if (stats == null) {
        Text(stringResource(R.string.statistics_unavailable))
        return
    }
    val rank = if (stats.rank == 0L) {
        stringResource(R.string.unranked)
    } else {
        stringResource(R.string.count_of_total, stats.rank, stats.rankedPlayers)
    }
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Stat(R.string.rank, rank)
        Stat(R.string.games_played, stats.gamesPlayed.toString())
        Stat(R.string.wins, stats.wins.toString())
        Stat(R.string.losses, stats.losses.toString())
        Stat(R.string.draws, stats.draws.toString())
        Stat(R.string.win_rate, stringResource(R.string.percentage, stats.winRate))
        Stat(R.string.total_points, stats.totalPoints.toString())
        Stat(R.string.average_score, stringResource(R.string.decimal_statistic, stats.averagePoints))
        Stat(R.string.best_score, stats.bestScore.toString())
        Stat(R.string.correct_answers, stringResource(R.string.count_of_total, stats.correctAnswers, stats.answersSubmitted))
        Stat(R.string.accuracy, stringResource(R.string.percentage, stats.accuracy))
    }
}

@Composable
private fun Stat(label: Int, value: String) {
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(16.dp)) {
        Text(stringResource(label), Modifier.weight(1f))
        Text(value)
    }
}

@Composable
private fun PlayPage(
    screen: AppScreen.Play,
    onJoin: (MatchSelection) -> Unit,
    onRetry: () -> Unit,
    onCancel: () -> Unit,
    onBack: () -> Unit,
) {
    var categoryChoice by rememberSaveable { mutableStateOf<String?>(null) }
    var timerChoice by rememberSaveable { mutableStateOf<String?>(null) }
    var playerChoice by rememberSaveable { mutableStateOf<Int?>(null) }
    val options = screen.options
    val queue = screen.queue

    Heading(stringResource(R.string.play))
    when {
        queue != null -> {
            Status(if (queue.acknowledged) R.string.waiting_for_players else R.string.joining_queue)
            SelectionSummary(queue.selection)
            OutlinedButton(onClick = onCancel) { Text(stringResource(R.string.cancel_search)) }
        }
        screen.busy -> Status(R.string.loading_match_options)
        options == null -> Button(onClick = onRetry) { Text(stringResource(R.string.try_again)) }
        else -> {
            val category = screen.categories.firstOrNull { it == categoryChoice } ?: screen.categories.firstOrNull()
            val timer = options.timeOptions.firstOrNull { it.id == timerChoice } ?: options.timeOptions.firstOrNull()
            val players = options.playerCounts.firstOrNull { it == playerChoice } ?: options.playerCounts.firstOrNull()
            if (category == null || timer == null || players == null) {
                Text(stringResource(R.string.no_match_options))
                Button(onClick = onRetry) { Text(stringResource(R.string.try_again)) }
            } else {
                ChoiceMenu(R.string.category, screen.categories, category, { categoryName(it) }) { categoryChoice = it }
                ChoiceMenu(R.string.question_timer, options.timeOptions, timer, { it.label }) { timerChoice = it.id }
                ChoiceMenu(R.string.players, options.playerCounts, players, { it.toString() }) { playerChoice = it }
                Button(onClick = { onJoin(MatchSelection(category, timer, players)) }) {
                    Text(stringResource(R.string.find_match))
                }
            }
        }
    }
    Notice(screen.notice)
    if (queue == null) TextButton(onClick = onBack) { Text(stringResource(R.string.back_to_profile)) }
}

@Composable
private fun <T> ChoiceMenu(
    label: Int,
    options: List<T>,
    selected: T,
    display: @Composable (T) -> String,
    onSelect: (T) -> Unit,
) {
    var expanded by remember { mutableStateOf(false) }
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(stringResource(label), style = MaterialTheme.typography.titleSmall)
        Box {
            OutlinedButton(onClick = { expanded = true }) { Text(display(selected)) }
            DropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
                options.forEach { option ->
                    DropdownMenuItem(
                        text = { Text(display(option)) },
                        onClick = {
                            onSelect(option)
                            expanded = false
                        },
                    )
                }
            }
        }
    }
}

@Composable
private fun MatchPage(screen: AppScreen.Match, onLeave: () -> Unit) {
    Heading(stringResource(R.string.match_found))
    SelectionSummary(screen.selection)
    if (BuildConfig.DEBUG) {
        Text(stringResource(R.string.match_reference, screen.gameId), style = MaterialTheme.typography.bodySmall)
    }
    Notice(screen.notice)
    OutlinedButton(onClick = onLeave) { Text(stringResource(R.string.leave_match)) }
}

@Composable
private fun SelectionSummary(selection: MatchSelection) {
    Text(
        pluralStringResource(
            R.plurals.selection_summary,
            selection.players,
            categoryName(selection.category),
            selection.timeOption.label,
            selection.players,
        ),
    )
}

@Composable
private fun categoryName(category: String): String = when (category) {
    "SPORT" -> stringResource(R.string.category_sports)
    "MUSIC" -> stringResource(R.string.category_music)
    "TECH" -> stringResource(R.string.category_tech)
    else -> category
}

@Composable
private fun Heading(title: String) {
    Text(title, style = MaterialTheme.typography.headlineMedium)
}

@Composable
private fun Status(message: Int) {
    Text(stringResource(message), Modifier.semantics { liveRegion = LiveRegionMode.Polite })
}

@Composable
private fun Notice(message: Int?) {
    if (message != null) {
        Text(
            stringResource(message),
            Modifier.semantics { liveRegion = LiveRegionMode.Polite },
            color = MaterialTheme.colorScheme.error,
        )
    }
}

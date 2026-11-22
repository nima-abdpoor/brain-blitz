package com.brainblitz.android

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.KeyboardArrowDown
import androidx.compose.material.icons.filled.KeyboardArrowUp
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusDirection
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.unit.dp

@Composable
internal fun AccountPage(
    screen: AppScreen.Account,
    configured: Boolean,
    onSubmit: (String, String) -> Unit,
    onSwitchMode: () -> Unit,
) {
    var email by rememberSaveable { mutableStateOf("") }
    var password by remember(screen.mode) { mutableStateOf("") }
    var passwordVisible by remember(screen.mode) { mutableStateOf(false) }
    val focusManager = LocalFocusManager.current
    val enabled = configured && !screen.busy
    val signingUp = screen.mode == AccountMode.SignUp
    val canSubmit = enabled && email.isNotBlank() && password.isNotEmpty()

    fun submit() {
        if (!canSubmit) return
        focusManager.clearFocus()
        onSubmit(email.trim(), password)
        password = ""
        passwordVisible = false
    }

    ArenaPage(top = { ArenaBrandBar() }) {
        ArenaHero(
            title = stringResource(if (signingUp) R.string.join_brainblitz else R.string.welcome_back),
            description = stringResource(R.string.multiplayer_trivia),
            symbol = R.drawable.ic_arena_bolt,
        )
        OutlinedTextField(
            value = email,
            onValueChange = { email = it },
            label = { Text(stringResource(R.string.email)) },
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email, imeAction = ImeAction.Next),
            keyboardActions = KeyboardActions(onNext = { focusManager.moveFocus(FocusDirection.Down) }),
            singleLine = true,
            enabled = enabled,
            modifier = Modifier.fillMaxWidth(),
        )
        OutlinedTextField(
            value = password,
            onValueChange = { password = it },
            label = { Text(stringResource(R.string.password)) },
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password, imeAction = ImeAction.Done),
            keyboardActions = KeyboardActions(onDone = { submit() }),
            visualTransformation = if (passwordVisible) VisualTransformation.None else PasswordVisualTransformation(),
            trailingIcon = {
                TextButton(onClick = { passwordVisible = !passwordVisible }, enabled = enabled) {
                    Text(stringResource(if (passwordVisible) R.string.hide_password else R.string.show_password))
                }
            },
            singleLine = true,
            enabled = enabled,
            modifier = Modifier.fillMaxWidth(),
        )
        Notice(screen.notice)
        ArenaButton(
            label = if (signingUp) R.string.create_account else R.string.sign_in,
            onClick = ::submit,
            enabled = canSubmit,
            forward = true,
        )
        TextButton(onClick = onSwitchMode, enabled = enabled, modifier = Modifier.align(Alignment.CenterHorizontally)) {
            Text(stringResource(if (signingUp) R.string.use_existing_account else R.string.create_account))
        }
        if (screen.busy) Status(R.string.working)
    }
}

@Composable
internal fun ProfilePage(
    screen: AppScreen.ProfilePage,
    onPlay: () -> Unit,
    onRetry: () -> Unit,
    onRetryRecovery: () -> Unit,
    onSignOut: () -> Unit,
) {
    val profile = screen.profile
    val canPlay = profile != null && screen.recovery != RecoveryStatus.Connecting
    ArenaPage(
        top = { ArenaBrandBar() },
        bottom = { ArenaNavigation(onPlay, {}, playing = false, enabled = canPlay) },
    ) {
        Heading(stringResource(R.string.profile))
        if (profile != null) {
            Surface(color = MaterialTheme.colorScheme.primaryContainer, shape = MaterialTheme.shapes.large) {
                Row(
                    Modifier.fillMaxWidth().padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    Box(
                        Modifier.size(48.dp).background(MaterialTheme.colorScheme.surface, CircleShape),
                        contentAlignment = Alignment.Center,
                    ) {
                        val initials = playerInitials(profile.displayName)
                        if (initials.isEmpty()) Icon(Icons.Filled.Person, null)
                        else Text(initials, Modifier.clearAndSetSemantics {}, style = MaterialTheme.typography.titleMedium)
                    }
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Text(profile.displayName, style = MaterialTheme.typography.titleLarge)
                        Text(profile.username, style = MaterialTheme.typography.bodySmall)
                    }
                }
            }
            screen.recovery?.let { RecoveryControls(it, onRetryRecovery) }
            ProfileStatistics(profile.statistics)
        } else if (screen.busy) {
            ArenaState(stringResource(R.string.loading_profile), Icons.Filled.Person, loading = true)
        } else {
            ArenaButton(R.string.try_again, onRetry)
        }
        Notice(screen.notice)
        ArenaSecondaryButton(R.string.sign_out, onSignOut)
    }
}

@Composable
private fun ProfileStatistics(stats: PlayerStatistics?) {
    if (stats == null) {
        Status(R.string.statistics_unavailable)
        return
    }
    var expanded by rememberSaveable { mutableStateOf(false) }
    val rank = if (stats.rank == 0L) {
        stringResource(R.string.unranked)
    } else {
        stringResource(R.string.count_of_total, stats.rank, stats.rankedPlayers)
    }
    Text(
        stringResource(R.string.game_statistics),
        Modifier.semantics { heading() },
        style = MaterialTheme.typography.titleMedium,
    )
    val summary = listOf(
        R.string.rank to rank,
        R.string.games_played to stats.gamesPlayed.toString(),
        R.string.wins to stats.wins.toString(),
        R.string.accuracy to stringResource(R.string.percentage, stats.accuracy),
    )
    summary.chunked(2).forEach { row ->
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            row.forEach { (label, value) ->
                Surface(
                    modifier = Modifier.weight(1f),
                    color = MaterialTheme.colorScheme.surface,
                    shape = MaterialTheme.shapes.medium,
                ) {
                    Column(Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Text(stringResource(label), color = MaterialTheme.colorScheme.onSurfaceVariant, style = MaterialTheme.typography.bodySmall)
                        Text(value, style = MaterialTheme.typography.titleLarge)
                    }
                }
            }
        }
    }
    TextButton(onClick = { expanded = !expanded }) {
        Text(stringResource(if (expanded) R.string.fewer_stats else R.string.more_stats))
        Icon(if (expanded) Icons.Filled.KeyboardArrowUp else Icons.Filled.KeyboardArrowDown, null)
    }
    if (expanded) {
        Stat(R.string.losses, stats.losses.toString())
        Stat(R.string.draws, stats.draws.toString())
        Stat(R.string.win_rate, stringResource(R.string.percentage, stats.winRate))
        Stat(R.string.total_points, stats.totalPoints.toString())
        Stat(R.string.average_score, stringResource(R.string.decimal_statistic, stats.averagePoints))
        Stat(R.string.best_score, stats.bestScore.toString())
        Stat(R.string.correct_answers, stringResource(R.string.count_of_total, stats.correctAnswers, stats.answersSubmitted))
    }
}

@Composable
private fun Stat(label: Int, value: String) {
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Text(stringResource(label), Modifier.weight(1f), color = MaterialTheme.colorScheme.onSurfaceVariant)
            Text(value, Modifier.weight(1f), textAlign = androidx.compose.ui.text.style.TextAlign.End)
        }
        HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
    }
}

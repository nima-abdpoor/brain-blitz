package com.brainblitz.android

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.BackHandler
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.key
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.compose.collectAsStateWithLifecycle

class MainActivity : ComponentActivity() {
    private val model by lazy { ViewModelProvider(this)[BrainBlitzViewModel::class.java] }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            val screen by model.screen.collectAsStateWithLifecycle()
            ArenaTheme {
                BrainBlitzApp(screen, model)
            }
        }
    }

    override fun onStart() {
        super.onStart()
        model.setForeground(true)
    }

    override fun onStop() {
        model.setForeground(false)
        super.onStop()
    }
}

@Composable
private fun BrainBlitzApp(screen: AppScreen, model: BrainBlitzViewModel) {
    val match = screen as? AppScreen.Match
    val activeMatch = match != null && !match.game.terminal
    var confirmLeave by rememberSaveable(match?.game?.gameId) { mutableStateOf(false) }
    val onLeave: () -> Unit = {
        if (activeMatch) confirmLeave = true else model.returnToProfile()
    }

    BackHandler(enabled = screen is AppScreen.Play || match != null) {
        if (screen is AppScreen.Play && screen.queue != null) model.cancelQueue()
        else onLeave()
    }

    Surface(color = MaterialTheme.colorScheme.background, modifier = Modifier.fillMaxSize()) {
        Box(Modifier.fillMaxSize().safeDrawingPadding().imePadding()) {
            key(screen::class, match?.game?.gameId, match?.game?.roundNumber, match?.game?.terminal) {
                when (screen) {
                    is AppScreen.Account -> AccountPage(
                        screen, model.configured, model::submitAccount, model::changeAccountMode,
                    )
                    is AppScreen.ProfilePage -> ProfilePage(
                        screen, model::openPlay, model::retryProfile, model::retryRecovery, model::signOut,
                    )
                    is AppScreen.Play -> PlayPage(
                        screen, model::joinQueue, model::retryOptions, model::cancelQueue, model::returnToProfile,
                    )
                    is AppScreen.Match -> MatchPage(
                        screen, model::ready, model::answer, model::playAgain, model::retryRecovery, onLeave,
                    )
                }
            }
            if (confirmLeave && activeMatch) {
                AlertDialog(
                    onDismissRequest = { confirmLeave = false },
                    title = { Text(stringResource(R.string.confirm_leave_match)) },
                    text = { Text(stringResource(R.string.match_continues)) },
                    confirmButton = {
                        TextButton(onClick = {
                            confirmLeave = false
                            model.returnToProfile()
                        }) { Text(stringResource(R.string.leave_match)) }
                    },
                    dismissButton = {
                        TextButton(onClick = { confirmLeave = false }) { Text(stringResource(R.string.stay_in_match)) }
                    },
                )
            }
        }
    }
}

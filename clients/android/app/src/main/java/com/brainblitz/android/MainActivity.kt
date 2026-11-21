package com.brainblitz.android

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.compose.collectAsStateWithLifecycle

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        val model = ViewModelProvider(this)[ConnectionViewModel::class.java]
        setContent {
            val colors = if (isSystemInDarkTheme()) darkColorScheme() else lightColorScheme()
            val state by model.state.collectAsStateWithLifecycle()
            MaterialTheme(colorScheme = colors) {
                Surface(Modifier.fillMaxSize()) {
                    Column(
                        modifier = Modifier
                            .safeDrawingPadding()
                            .imePadding()
                            .verticalScroll(rememberScrollState())
                            .padding(24.dp),
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                    ) {
                        val check = (state as? ConnectionState.Connected)?.check
                        if (check == null) {
                            CheckForm(state, model.configured, model::check)
                        } else {
                            CheckResult(check, model::reset)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun CheckForm(state: ConnectionState, configured: Boolean, onCheck: (String, String) -> Unit) {
    var email by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    val focusManager = LocalFocusManager.current
    val checking = state == ConnectionState.Checking

    Text(stringResource(R.string.connection_title), style = MaterialTheme.typography.headlineMedium)
    Text(stringResource(R.string.connection_description))
    OutlinedTextField(
        value = email,
        onValueChange = { email = it },
        label = { Text(stringResource(R.string.email)) },
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email),
        singleLine = true,
        enabled = configured && !checking,
        modifier = Modifier.fillMaxWidth(),
    )
    OutlinedTextField(
        value = password,
        onValueChange = { password = it },
        label = { Text(stringResource(R.string.password)) },
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
        visualTransformation = PasswordVisualTransformation(),
        singleLine = true,
        enabled = configured && !checking,
        modifier = Modifier.fillMaxWidth(),
    )
    Button(
        enabled = configured && !checking && email.isNotBlank() && password.isNotEmpty(),
        onClick = {
            focusManager.clearFocus()
            onCheck(email.trim(), password)
            password = ""
        },
    ) {
        val label = if (checking) R.string.checking_connection else R.string.check_connection
        Text(stringResource(label))
    }
    val messageId = if (configured) (state as? ConnectionState.Failed)?.messageId else R.string.missing_gateway
    if (messageId != null) {
        Text(stringResource(messageId), color = MaterialTheme.colorScheme.error)
    }
}

@Composable
private fun CheckResult(check: GatewayCheck, onReset: () -> Unit) {
    Text(stringResource(R.string.connected_title), style = MaterialTheme.typography.headlineMedium)
    Text(stringResource(R.string.signed_in_as, check.displayName))
    OptionSummary(stringResource(R.string.categories), check.categories.joinToString(", "))
    OptionSummary(stringResource(R.string.player_counts), check.playerCounts.joinToString(", "))
    val timers = check.timeOptions.map { stringResource(R.string.question_timer, it.label, it.seconds) }
    OptionSummary(stringResource(R.string.question_timers), timers.joinToString("\n"))
    Button(onClick = onReset) { Text(stringResource(R.string.check_again)) }
}

@Composable
private fun OptionSummary(label: String, items: String) {
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(label, style = MaterialTheme.typography.titleMedium)
        Text(items.ifEmpty { stringResource(R.string.empty_options) })
    }
}

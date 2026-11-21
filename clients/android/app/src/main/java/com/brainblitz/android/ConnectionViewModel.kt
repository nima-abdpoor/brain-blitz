package com.brainblitz.android

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import java.io.IOException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.serialization.SerializationException
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull

internal sealed interface ConnectionState {
    data object Idle : ConnectionState
    data object Checking : ConnectionState
    data class Connected(val check: GatewayCheck) : ConnectionState
    data class Failed(val messageId: Int) : ConnectionState
}

internal class ConnectionViewModel : ViewModel() {
    private val gateway = BuildConfig.GATEWAY_URL.toHttpUrlOrNull()?.let { Gateway(it) }
    private val mutableState = MutableStateFlow<ConnectionState>(ConnectionState.Idle)
    val state = mutableState.asStateFlow()
    val configured = gateway != null

    fun check(email: String, password: String) {
        val gateway = gateway ?: return
        if (mutableState.value == ConnectionState.Checking) return
        mutableState.value = ConnectionState.Checking
        viewModelScope.launch {
            mutableState.value = try {
                ConnectionState.Connected(gateway.check(email, password))
            } catch (error: GatewayHttpException) {
                ConnectionState.Failed(
                    when {
                        error.endpoint == Gateway.LOGIN_PATH && (error.status == 401 || error.status == 403) ->
                            R.string.invalid_login
                        error.status == 401 -> R.string.sign_in_rejected
                        else -> R.string.gateway_error
                    },
                )
            } catch (_: SerializationException) {
                ConnectionState.Failed(R.string.invalid_response)
            } catch (_: IOException) {
                ConnectionState.Failed(R.string.connection_error)
            }
        }
    }

    fun reset() {
        mutableState.value = ConnectionState.Idle
    }
}

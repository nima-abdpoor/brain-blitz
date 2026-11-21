package com.brainblitz.android

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.AtomicFile
import java.io.File
import java.io.FileNotFoundException
import java.io.IOException
import java.security.GeneralSecurityException
import java.security.KeyStore
import java.time.Instant
import java.util.Base64
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

internal class SessionStore(context: Context, gatewayOrigin: String) {
    private val file = AtomicFile(File(context.noBackupFilesDir, "session-v1"))
    private val origin = gatewayOrigin.toByteArray(Charsets.UTF_8)

    suspend fun read(): String? = withContext(Dispatchers.IO) {
        val bytes = try {
            file.readFully()
        } catch (_: FileNotFoundException) {
            return@withContext null
        }
        try {
            if (bytes.size <= IV_BYTES + TAG_BYTES) throw GeneralSecurityException("Invalid session")
            val key = loadKey() ?: throw GeneralSecurityException("Missing session key")
            val cipher = Cipher.getInstance(TRANSFORMATION)
            cipher.init(Cipher.DECRYPT_MODE, key, GCMParameterSpec(TAG_BYTES * 8, bytes, 0, IV_BYTES))
            cipher.updateAAD(origin)
            cipher.doFinal(bytes, IV_BYTES, bytes.size - IV_BYTES).toString(Charsets.UTF_8)
        } catch (_: GeneralSecurityException) {
            deleteSession()
            null
        }
    }

    suspend fun write(token: String) = withContext(Dispatchers.IO) {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.ENCRYPT_MODE, loadKey() ?: createKey())
        val iv = cipher.iv
        if (iv.size != IV_BYTES) throw GeneralSecurityException("Unexpected session nonce length")
        cipher.updateAAD(origin)
        val encrypted = iv + cipher.doFinal(token.toByteArray(Charsets.UTF_8))

        val stream = file.startWrite()
        try {
            stream.write(encrypted)
            file.finishWrite(stream)
        } catch (error: IOException) {
            file.failWrite(stream)
            throw error
        }
    }

    suspend fun clear() = withContext(Dispatchers.IO) { deleteSession() }

    private fun deleteSession() {
        file.delete()
        if (file.baseFile.exists()) throw IOException("Could not remove the stored session")
    }

    private fun loadKey(): SecretKey? =
        KeyStore.getInstance(KEYSTORE).apply { load(null) }.getKey(KEY_ALIAS, null) as? SecretKey

    private fun createKey(): SecretKey {
        val generator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, KEYSTORE)
        generator.init(
            KeyGenParameterSpec.Builder(
                KEY_ALIAS,
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT,
            ).setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .build(),
        )
        return generator.generateKey()
    }

    private companion object {
        const val KEYSTORE = "AndroidKeyStore"
        const val KEY_ALIAS = "brainblitz.access-token.v1"
        const val TRANSFORMATION = "AES/GCM/NoPadding"
        const val IV_BYTES = 12
        const val TAG_BYTES = 16
    }
}

private val claimsJson = Json { ignoreUnknownKeys = true }

@Serializable
private data class TokenClaims(val exp: Long? = null)

internal fun tokenExpiration(token: String): Instant? {
    val claims = token.split('.').getOrNull(1) ?: return null
    return try {
        val json = Base64.getUrlDecoder().decode(claims).toString(Charsets.UTF_8)
        claimsJson.decodeFromString<TokenClaims>(json).exp?.let(Instant::ofEpochSecond)
    } catch (_: IllegalArgumentException) {
        null
    } catch (_: java.time.DateTimeException) {
        null
    }
}

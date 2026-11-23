package com.brainblitz.android

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Shapes
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp

private val lightArena = lightColorScheme(
    primary = Color(0xFF6137C8),
    onPrimary = Color.White,
    primaryContainer = Color(0xFFECE3FF),
    onPrimaryContainer = Color(0xFF32166F),
    background = Color(0xFFFAF7FF),
    onBackground = Color(0xFF201335),
    surface = Color.White,
    onSurface = Color(0xFF201335),
    surfaceVariant = Color(0xFFF0E9FC),
    onSurfaceVariant = Color(0xFF5F526F),
    outline = Color(0xFF81718F),
    outlineVariant = Color(0xFFDFD5EF),
    error = Color(0xFF8B2D39),
    errorContainer = Color(0xFFFCE7EB),
    onErrorContainer = Color(0xFF762331),
)

private val darkArena = darkColorScheme(
    primary = Color(0xFFCBB3FF),
    onPrimary = Color(0xFF261342),
    primaryContainer = Color(0xFF3A2860),
    onPrimaryContainer = Color(0xFFEEE3FF),
    background = Color(0xFF151123),
    onBackground = Color(0xFFF7F0FF),
    surface = Color(0xFF241B36),
    onSurface = Color(0xFFF7F0FF),
    surfaceVariant = Color(0xFF2F2443),
    onSurfaceVariant = Color(0xFFC6B8D7),
    outline = Color(0xFF9381A9),
    outlineVariant = Color(0xFF4A3A5E),
    error = Color(0xFFFFC9CD),
    errorContainer = Color(0xFF4A2535),
    onErrorContainer = Color(0xFFFFD8DC),
)

internal data class ArenaTones(
    val heroStart: Color,
    val heroEnd: Color,
    val correctSurface: Color,
    val correctText: Color,
) {
    val lime = Color(0xFFE3FA8B)
    val onLime = Color(0xFF282039)
}

private val lightTones = ArenaTones(
    heroStart = Color(0xFF5021B6),
    heroEnd = Color(0xFF7241CF),
    correctSurface = Color(0xFFE2F6E9),
    correctText = Color(0xFF20583E),
)

private val darkTones = ArenaTones(
    heroStart = Color(0xFF3A2187),
    heroEnd = Color(0xFF5E39AD),
    correctSurface = Color(0xFF203C2C),
    correctText = Color(0xFFB6F1C8),
)

internal val LocalArenaTones = staticCompositionLocalOf { lightTones }

private val baseTypography = Typography()
private val arenaTypography = baseTypography.copy(
    displayMedium = baseTypography.displayMedium.copy(fontWeight = FontWeight.Black),
    headlineMedium = baseTypography.headlineMedium.copy(fontWeight = FontWeight.ExtraBold),
    headlineSmall = baseTypography.headlineSmall.copy(fontWeight = FontWeight.ExtraBold),
    titleLarge = baseTypography.titleLarge.copy(fontWeight = FontWeight.Bold),
    titleMedium = baseTypography.titleMedium.copy(fontWeight = FontWeight.Bold),
    titleSmall = baseTypography.titleSmall.copy(fontWeight = FontWeight.Bold),
    labelLarge = baseTypography.labelLarge.copy(fontWeight = FontWeight.Bold),
)

@Composable
internal fun ArenaTheme(content: @Composable () -> Unit) {
    val dark = isSystemInDarkTheme()
    CompositionLocalProvider(LocalArenaTones provides if (dark) darkTones else lightTones) {
        MaterialTheme(
            colorScheme = if (dark) darkArena else lightArena,
            typography = arenaTypography,
            shapes = Shapes(
                small = RoundedCornerShape(10.dp),
                medium = RoundedCornerShape(14.dp),
                large = RoundedCornerShape(22.dp),
                extraLarge = RoundedCornerShape(28.dp),
            ),
            content = content,
        )
    }
}

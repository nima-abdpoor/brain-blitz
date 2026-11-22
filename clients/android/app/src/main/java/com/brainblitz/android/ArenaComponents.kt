package com.brainblitz.android

import androidx.annotation.DrawableRes
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.interaction.InteractionSource
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.interaction.collectIsPressedAsState
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.selection.selectableGroup
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.MotionDurationScale
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.drawscope.scale
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.LiveRegionMode
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.liveRegion
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import kotlin.coroutines.coroutineContext

@Composable
internal fun ArenaPage(
    top: @Composable () -> Unit = {},
    bottom: @Composable () -> Unit = {},
    centered: Boolean = false,
    content: @Composable ColumnScope.() -> Unit,
) {
    Column(Modifier.fillMaxSize(), horizontalAlignment = Alignment.CenterHorizontally) {
        Box(Modifier.widthIn(max = 600.dp).fillMaxWidth()) { top() }
        Column(
            modifier = Modifier
                .weight(1f)
                .widthIn(max = 600.dp)
                .fillMaxWidth()
                .verticalScroll(rememberScrollState())
                .padding(start = 18.dp, top = 8.dp, end = 18.dp, bottom = 18.dp),
            verticalArrangement = Arrangement.spacedBy(
                space = 14.dp,
                alignment = if (centered) Alignment.CenterVertically else Alignment.Top,
            ),
            content = content,
        )
        Box(Modifier.widthIn(max = 600.dp).fillMaxWidth()) { bottom() }
    }
}

@Composable
internal fun ArenaHero(
    title: String,
    description: String? = null,
    @DrawableRes symbol: Int? = null,
    titleStyle: TextStyle = MaterialTheme.typography.displayMedium,
) {
    val tones = LocalArenaTones.current
    val shape = MaterialTheme.shapes.large
    Box(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(Brush.horizontalGradient(listOf(tones.heroStart, tones.heroEnd))),
    ) {
        if (symbol != null) {
            Canvas(Modifier.matchParentSize()) {
                val center = Offset(size.width * .86f, size.height * .52f)
                for (radius in 1..5) {
                    drawCircle(Color.White.copy(alpha = .19f), radius * 21.dp.toPx(), center, style = Stroke(1.dp.toPx()))
                }
            }
        }
        Row(
            Modifier.fillMaxWidth().padding(17.dp),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                if (description != null) {
                    Text(description, style = MaterialTheme.typography.labelMedium, color = Color(0xFFF0E6FF))
                }
                Text(title, Modifier.semantics { heading() }, color = Color.White, style = titleStyle)
            }
            if (symbol != null) {
                Box(
                    Modifier.size(49.dp).rotate(-12f).background(tones.lime, MaterialTheme.shapes.medium),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(painterResource(symbol), contentDescription = null, modifier = Modifier.size(26.dp), tint = tones.onLime)
                }
            }
        }
    }
}

@Composable
internal fun ArenaBrandBar(profile: Profile? = null, onProfile: (() -> Unit)? = null) {
    Row(
        Modifier.fillMaxWidth().heightIn(min = 54.dp).padding(horizontal = 18.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        Box(
            Modifier.size(35.dp).background(MaterialTheme.colorScheme.primary, MaterialTheme.shapes.small),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                painterResource(R.drawable.ic_arena_bolt),
                contentDescription = null,
                modifier = Modifier.size(20.dp),
                tint = MaterialTheme.colorScheme.onPrimary,
            )
        }
        Text(stringResource(R.string.app_name), Modifier.weight(1f), style = MaterialTheme.typography.titleMedium)
        if (profile != null && onProfile != null) {
            val label = stringResource(R.string.profile)
            Surface(
                onClick = onProfile,
                shape = CircleShape,
                color = MaterialTheme.colorScheme.surfaceVariant,
                modifier = Modifier.size(48.dp).semantics { contentDescription = label },
            ) {
                Box(contentAlignment = Alignment.Center) {
                    val initials = playerInitials(profile.displayName)
                    if (initials.isEmpty()) Icon(Icons.Filled.Person, contentDescription = null)
                    else Text(initials, Modifier.clearAndSetSemantics {}, style = MaterialTheme.typography.labelLarge)
                }
            }
        }
    }
}

internal fun playerInitials(name: String): String = name.trim().split(Regex("\\s+")).take(2)
    .joinToString("") { part ->
        if (part.isEmpty()) "" else part.substring(0, part.offsetByCodePoints(0, 1)).uppercase()
    }

@Composable
internal fun ArenaMatchBar(title: String, action: ImageVector, actionLabel: String, onAction: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().heightIn(min = 54.dp).padding(start = 8.dp, end = 18.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        IconButton(onClick = onAction) { Icon(action, actionLabel) }
        Text(title, Modifier.weight(1f).semantics { heading() }, style = MaterialTheme.typography.titleMedium)
    }
}

@Composable
internal fun ArenaNavigation(onPlay: () -> Unit, onProfile: () -> Unit, playing: Boolean, enabled: Boolean = true) {
    val colors = MaterialTheme.colorScheme
    Surface(color = colors.surface) {
        Column {
            HorizontalDivider(color = colors.outlineVariant)
            Row(
                Modifier.fillMaxWidth().selectableGroup().padding(horizontal = 18.dp, vertical = 5.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                for (playTab in listOf(true, false)) {
                    val selected = playTab == playing
                    Row(
                        Modifier
                            .weight(1f)
                            .heightIn(min = 48.dp)
                            .clip(MaterialTheme.shapes.medium)
                            .background(if (selected) colors.primaryContainer else colors.surface)
                            .selectable(selected, enabled = enabled, role = Role.Tab, onClick = if (playTab) onPlay else onProfile)
                            .padding(6.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(6.dp, Alignment.CenterHorizontally),
                    ) {
                        val tint = if (selected) colors.onPrimaryContainer else colors.onSurfaceVariant
                        Icon(if (playTab) Icons.Filled.PlayArrow else Icons.Filled.Person, null, Modifier.size(18.dp), tint)
                        Text(
                            stringResource(if (playTab) R.string.play else R.string.profile),
                            style = MaterialTheme.typography.labelLarge,
                            color = tint,
                        )
                    }
                }
            }
        }
    }
}

@Composable
internal fun ArenaActions(content: @Composable ColumnScope.() -> Unit) {
    Column(
        Modifier.fillMaxWidth().padding(start = 18.dp, end = 18.dp, top = 4.dp, bottom = 10.dp),
        verticalArrangement = Arrangement.spacedBy(6.dp),
        content = content,
    )
}

@Composable
internal fun ArenaButton(
    label: Int,
    onClick: () -> Unit,
    enabled: Boolean = true,
    forward: Boolean = false,
    pressMotion: Boolean = false,
) {
    val interactions = if (pressMotion) remember { MutableInteractionSource() } else null
    val bounds = Modifier.fillMaxWidth().heightIn(min = 50.dp)
    Button(
        onClick = onClick,
        enabled = enabled,
        modifier = if (interactions == null) bounds else bounds.arenaPressMotion(interactions, enabled, pressedScale = .982f),
        shape = MaterialTheme.shapes.medium,
        interactionSource = interactions,
    ) {
        Text(stringResource(label), textAlign = TextAlign.Center, modifier = if (forward) Modifier.weight(1f, fill = false) else Modifier)
        if (forward) {
            Icon(Icons.AutoMirrored.Filled.ArrowForward, null, Modifier.padding(start = 8.dp).size(18.dp))
        }
    }
}

@Composable
internal fun Modifier.arenaPressMotion(
    interactions: InteractionSource,
    enabled: Boolean,
    pressedScale: Float,
): Modifier {
    val pressed by interactions.collectIsPressedAsState()
    val compression = remember { Animatable(1f) }
    val pressing = enabled && pressed
    LaunchedEffect(pressing) {
        if (coroutineContext[MotionDurationScale]?.scaleFactor == 0f) {
            compression.snapTo(1f)
        } else {
            val target = if (pressing) pressedScale else 1f
            if (compression.value != target) {
                compression.animateTo(target, tween(if (pressing) 85 else 125))
            }
        }
    }
    return drawWithContent {
        // Change the paint only so the press keeps its original touch bounds.
        val factor = compression.value
        if (factor == 1f) drawContent() else scale(factor, factor) { this@drawWithContent.drawContent() }
    }
}

@Composable
internal fun ArenaSecondaryButton(label: Int, onClick: () -> Unit, enabled: Boolean = true) {
    OutlinedButton(
        onClick = onClick,
        enabled = enabled,
        modifier = Modifier.fillMaxWidth().heightIn(min = 48.dp),
        shape = MaterialTheme.shapes.medium,
    ) {
        Text(stringResource(label), textAlign = TextAlign.Center)
    }
}

@Composable
internal fun ArenaState(title: String, icon: ImageVector, loading: Boolean = false, about: String? = null) {
    val colors = MaterialTheme.colorScheme
    Column(
        Modifier.fillMaxWidth().padding(horizontal = 8.dp, vertical = 12.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        Box(
            Modifier.size(86.dp).border(1.dp, colors.outlineVariant, CircleShape),
            contentAlignment = Alignment.Center,
        ) {
            Box(
                Modifier.size(66.dp).background(colors.primaryContainer, CircleShape),
                contentAlignment = Alignment.Center,
            ) {
                Icon(icon, null, Modifier.size(29.dp), colors.onPrimaryContainer)
            }
            if (loading) {
                CircularProgressIndicator(Modifier.size(84.dp), strokeWidth = 2.dp)
            }
        }
        Text(
            title,
            Modifier.semantics { heading(); liveRegion = LiveRegionMode.Polite },
            style = MaterialTheme.typography.headlineSmall,
            textAlign = TextAlign.Center,
        )
        if (about != null) {
            Text(about, color = colors.onSurfaceVariant, textAlign = TextAlign.Center)
        }
    }
}

@Composable
internal fun ArenaBanner(
    title: String,
    icon: ImageVector,
    description: String? = null,
    container: Color = MaterialTheme.colorScheme.primaryContainer,
    textColor: Color = MaterialTheme.colorScheme.onPrimaryContainer,
) {
    Surface(color = container, contentColor = textColor, shape = MaterialTheme.shapes.medium) {
        Row(
            Modifier.fillMaxWidth().semantics { liveRegion = LiveRegionMode.Polite }.padding(12.dp),
            horizontalArrangement = Arrangement.spacedBy(9.dp),
        ) {
            Icon(icon, null, Modifier.size(19.dp))
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                Text(title, style = MaterialTheme.typography.titleSmall)
                if (description != null) Text(description, style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}

@Composable
internal fun Heading(title: String) {
    Text(title, Modifier.semantics { heading() }, style = MaterialTheme.typography.headlineSmall)
}

@Composable
internal fun Status(message: Int) {
    Text(
        stringResource(message),
        Modifier.semantics { liveRegion = LiveRegionMode.Polite },
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
}

@Composable
internal fun Notice(message: Int?) {
    if (message != null) {
        ArenaBanner(
            stringResource(message), Icons.Filled.Warning,
            container = MaterialTheme.colorScheme.errorContainer,
            textColor = MaterialTheme.colorScheme.onErrorContainer,
        )
    }
}

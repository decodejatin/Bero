package com.example.bero.ui.theme

import android.app.Activity
import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.dynamicDarkColorScheme
import androidx.compose.material3.dynamicLightColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.SideEffect
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowCompat

// ═══════════════════════════════════════════════════════
//  Luxury Royal Dark Color Scheme
// ═══════════════════════════════════════════════════════
private val DarkColorScheme = darkColorScheme(
    primary = BeroSecondary,                    // Gold as primary in dark mode
    onPrimary = BeroOnSecondary,                // Charcoal on Gold
    primaryContainer = BeroPrimaryLight,        // Muted Plum container
    onPrimaryContainer = BeroOnDarkSurface,     // Warm Cream
    secondary = BeroSecondaryLight,             // Light Gold
    onSecondary = BeroOnSecondary,
    secondaryContainer = BeroDarkSurfaceVariant,
    onSecondaryContainer = BeroOnDarkSurface,
    tertiary = LuxuryGoldLight,
    onTertiary = LuxuryCharcoal,
    background = BeroDarkBackground,            // Rich Onyx
    onBackground = BeroOnDarkBackground,        // Warm Cream
    surface = BeroDarkSurface,                  // Deep Charcoal
    onSurface = BeroOnDarkSurface,              // Warm Cream
    surfaceVariant = BeroDarkSurfaceVariant,    // Muted Plum
    onSurfaceVariant = Color(0xFFB8B0A0),       // Muted Gold-Gray
    error = BeroError,
    onError = Color(0xFFF5F0E8),
    outline = Color(0xFF3A3A52),                // Subtle border
    outlineVariant = Color(0xFF2A2A42)
)

// ═══════════════════════════════════════════════════════
//  Luxury Royal Light Color Scheme
// ═══════════════════════════════════════════════════════
private val LightColorScheme = lightColorScheme(
    primary = BeroPrimary,                      // Deep Charcoal
    onPrimary = BeroOnPrimary,                  // Warm Cream
    primaryContainer = Color(0xFFF0EBE0),       // Warm Sand
    onPrimaryContainer = BeroPrimary,           // Charcoal
    secondary = BeroSecondary,                  // Champagne Gold
    onSecondary = Color(0xFFFFFFFF),            // White on Gold
    secondaryContainer = Color(0xFFF8F0D8),     // Light Gold Wash
    onSecondaryContainer = BeroTertiary,        // Warm Bronze
    tertiary = BeroTertiary,                    // Warm Bronze
    onTertiary = Color(0xFFFFFFFF),
    background = BeroBackground,                // Warm Ivory
    onBackground = BeroOnBackground,            // Charcoal
    surface = BeroSurface,                      // Warm White
    onSurface = BeroOnSurface,                  // Charcoal
    surfaceVariant = BeroSurfaceVariant,        // Soft Linen
    onSurfaceVariant = Color(0xFF5A5A6E),       // Muted Charcoal
    error = BeroError,
    onError = Color(0xFFFFFFFF),
    outline = Color(0xFFD4CFCA),                // Warm border
    outlineVariant = Color(0xFFE8E3DE)          // Subtle border
)

@Composable
fun BeroTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    dynamicColor: Boolean = false,
    content: @Composable () -> Unit
) {
    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> DarkColorScheme
        else -> LightColorScheme
    }

    val view = LocalView.current
    if (!view.isInEditMode) {
        SideEffect {
            val window = (view.context as Activity).window
            // Status bar: deep charcoal in light mode, onyx in dark mode
            window.statusBarColor = if (darkTheme) {
                BeroDarkBackground.toArgb()
            } else {
                BeroPrimary.toArgb()
            }
            WindowCompat.getInsetsController(window, view).isAppearanceLightStatusBars = false
        }
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = BeroTypography,
        content = content
    )
}

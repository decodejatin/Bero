package com.example.bero.ui.splash

import androidx.compose.animation.core.*
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.R
import com.example.bero.ui.theme.LuxuryGold
import com.example.bero.ui.theme.LuxuryOnyx
import com.example.bero.ui.theme.LuxuryCharcoal
import kotlinx.coroutines.delay

/**
 * Luxury splash screen with animated Bero logo.
 * Rich onyx background with champagne gold accents.
 */
@Composable
fun SplashScreen(
    onSplashComplete: () -> Unit
) {
    // --- Animation States ---
    var startAnimation by remember { mutableStateOf(false) }

    val logoScale by animateFloatAsState(
        targetValue = if (startAnimation) 1f else 0.3f,
        animationSpec = tween(durationMillis = 700, easing = FastOutSlowInEasing),
        label = "logoScale"
    )

    val logoAlpha by animateFloatAsState(
        targetValue = if (startAnimation) 1f else 0f,
        animationSpec = tween(durationMillis = 800, easing = LinearEasing),
        label = "logoAlpha"
    )

    var showTitle by remember { mutableStateOf(false) }
    val titleAlpha by animateFloatAsState(
        targetValue = if (showTitle) 1f else 0f,
        animationSpec = tween(durationMillis = 600, easing = FastOutSlowInEasing),
        label = "titleAlpha"
    )

    var showTagline by remember { mutableStateOf(false) }
    val taglineAlpha by animateFloatAsState(
        targetValue = if (showTagline) 1f else 0f,
        animationSpec = tween(durationMillis = 500, easing = FastOutSlowInEasing),
        label = "taglineAlpha"
    )

    // Subtle pulse
    val infiniteTransition = rememberInfiniteTransition(label = "pulse")
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.03f,
        animationSpec = infiniteRepeatable(
            animation = tween(1400, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulseScale"
    )

    // Gold glow ring
    val glowAlpha by infiniteTransition.animateFloat(
        initialValue = 0.1f,
        targetValue = 0.35f,
        animationSpec = infiniteRepeatable(
            animation = tween(1600, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "glowAlpha"
    )

    LaunchedEffect(Unit) {
        startAnimation = true
        delay(400)
        showTitle = true
        delay(300)
        showTagline = true
        delay(1800)
        onSplashComplete()
    }

    // --- UI ---
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                brush = Brush.verticalGradient(
                    colors = listOf(
                        LuxuryOnyx,
                        LuxuryCharcoal,
                        LuxuryOnyx
                    )
                )
            ),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center,
            modifier = Modifier.fillMaxSize()
        ) {
            Spacer(modifier = Modifier.weight(1f))

            // Logo with gold glow ring
            Box(contentAlignment = Alignment.Center) {
                // Gold glow ring
                Box(
                    modifier = Modifier
                        .size(200.dp)
                        .scale(logoScale * pulseScale * 1.15f)
                        .alpha(logoAlpha * glowAlpha)
                        .clip(RoundedCornerShape(28.dp))
                        .background(
                            brush = Brush.radialGradient(
                                colors = listOf(
                                    LuxuryGold.copy(alpha = 0.5f),
                                    LuxuryGold.copy(alpha = 0.0f)
                                )
                            )
                        )
                )

                // Logo
                Image(
                    painter = painterResource(id = R.drawable.bero_logo),
                    contentDescription = "Bero Logo",
                    modifier = Modifier
                        .size(160.dp)
                        .scale(logoScale * pulseScale)
                        .alpha(logoAlpha)
                        .clip(RoundedCornerShape(24.dp)),
                    contentScale = ContentScale.Crop
                )
            }

            Spacer(modifier = Modifier.height(32.dp))

            // App name — Champagne Gold
            Text(
                text = "BERO",
                fontSize = 36.sp,
                fontWeight = FontWeight.ExtraBold,
                color = LuxuryGold,
                letterSpacing = 8.sp,
                modifier = Modifier.alpha(titleAlpha)
            )

            Spacer(modifier = Modifier.height(8.dp))

            // Tagline — Muted cream
            Text(
                text = "Your Trusted Local Services",
                fontSize = 14.sp,
                fontWeight = FontWeight.Normal,
                color = Color(0xFFF5F0E8).copy(alpha = 0.5f),
                letterSpacing = 2.sp,
                textAlign = TextAlign.Center,
                modifier = Modifier.alpha(taglineAlpha)
            )

            Spacer(modifier = Modifier.weight(1f))

            // Version
            Text(
                text = "v1.0",
                fontSize = 12.sp,
                fontWeight = FontWeight.Light,
                color = LuxuryGold.copy(alpha = 0.2f),
                modifier = Modifier
                    .padding(bottom = 32.dp)
                    .alpha(taglineAlpha)
            )
        }
    }
}

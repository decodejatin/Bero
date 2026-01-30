@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.onboarding

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

enum class AppLanguage(
    val code: String,
    val displayName: String,
    val nativeName: String,
    val greeting: String
) {
    ENGLISH("en", "English", "English", "Hello!"),
    HINDI("hi", "Hindi", "हिन्दी", "नमस्ते!")
}

@Composable
fun LanguageSelectionScreen(
    onLanguageSelected: (AppLanguage) -> Unit
) {
    var selectedLanguage by remember { mutableStateOf<AppLanguage?>(null) }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                brush = Brush.verticalGradient(
                    colors = listOf(
                        MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                        MaterialTheme.colorScheme.background,
                        MaterialTheme.colorScheme.secondary.copy(alpha = 0.05f)
                    )
                )
            )
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(60.dp))

            // App Logo/Icon placeholder
            Box(
                modifier = Modifier
                    .size(100.dp)
                    .clip(CircleShape)
                    .background(
                        brush = Brush.linearGradient(
                            colors = listOf(
                                MaterialTheme.colorScheme.primary,
                                MaterialTheme.colorScheme.secondary
                            )
                        )
                    ),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "B",
                    fontSize = 48.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Welcome to Bero",
                fontSize = 28.sp,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onBackground
            )

            Text(
                text = "Bero में आपका स्वागत है",
                fontSize = 20.sp,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
                modifier = Modifier.padding(top = 4.dp)
            )

            Spacer(modifier = Modifier.height(48.dp))

            Text(
                text = "Choose your language",
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                color = MaterialTheme.colorScheme.onBackground
            )

            Text(
                text = "अपनी भाषा चुनें",
                fontSize = 16.sp,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
                modifier = Modifier.padding(top = 4.dp)
            )

            Spacer(modifier = Modifier.height(32.dp))

            // Language Cards
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                AppLanguage.entries.forEach { language ->
                    LanguageCard(
                        language = language,
                        isSelected = selectedLanguage == language,
                        onClick = { selectedLanguage = language },
                        modifier = Modifier.weight(1f)
                    )
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            // Continue Button
            Button(
                onClick = { selectedLanguage?.let { onLanguageSelected(it) } },
                enabled = selectedLanguage != null,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                shape = RoundedCornerShape(16.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.primary
                )
            ) {
                Text(
                    text = if (selectedLanguage == AppLanguage.HINDI) "जारी रखें" else "Continue",
                    fontSize = 18.sp,
                    fontWeight = FontWeight.SemiBold
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            Text(
                text = "You can change this later in Settings",
                fontSize = 14.sp,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.5f),
                textAlign = TextAlign.Center
            )

            Text(
                text = "आप इसे बाद में सेटिंग्स में बदल सकते हैं",
                fontSize = 12.sp,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.4f),
                textAlign = TextAlign.Center,
                modifier = Modifier.padding(top = 4.dp)
            )

            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@Composable
private fun LanguageCard(
    language: AppLanguage,
    isSelected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val scale by animateFloatAsState(
        targetValue = if (isSelected) 1.02f else 1f,
        label = "scale"
    )

    val borderColor by animateColorAsState(
        targetValue = if (isSelected) MaterialTheme.colorScheme.primary else Color.Transparent,
        label = "borderColor"
    )

    val backgroundColor by animateColorAsState(
        targetValue = if (isSelected)
            MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.3f)
        else
            MaterialTheme.colorScheme.surface,
        label = "backgroundColor"
    )

    Card(
        modifier = modifier
            .scale(scale)
            .aspectRatio(0.85f)
            .clickable { onClick() },
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(containerColor = backgroundColor),
        border = BorderStroke(
            width = if (isSelected) 2.dp else 1.dp,
            color = if (isSelected) borderColor else MaterialTheme.colorScheme.outlineVariant
        ),
        elevation = CardDefaults.cardElevation(
            defaultElevation = if (isSelected) 8.dp else 2.dp
        )
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(20.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center
            ) {
                // Flag/Icon placeholder
                Box(
                    modifier = Modifier
                        .size(60.dp)
                        .clip(CircleShape)
                        .background(
                            if (language == AppLanguage.ENGLISH)
                                Color(0xFF1976D2)
                            else
                                Color(0xFFFF9800)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = if (language == AppLanguage.ENGLISH) "A" else "अ",
                        fontSize = 28.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )
                }

                Spacer(modifier = Modifier.height(16.dp))

                Text(
                    text = language.nativeName,
                    fontSize = 22.sp,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.onSurface
                )

                if (language != AppLanguage.ENGLISH) {
                    Text(
                        text = language.displayName,
                        fontSize = 14.sp,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                        modifier = Modifier.padding(top = 4.dp)
                    )
                }

                Spacer(modifier = Modifier.height(12.dp))

                Text(
                    text = language.greeting,
                    fontSize = 18.sp,
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.Medium
                )
            }

            // Selected checkmark
            if (isSelected) {
                Box(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(12.dp)
                        .size(28.dp)
                        .clip(CircleShape)
                        .background(MaterialTheme.colorScheme.primary),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        imageVector = Icons.Default.Check,
                        contentDescription = "Selected",
                        tint = Color.White,
                        modifier = Modifier.size(18.dp)
                    )
                }
            }
        }
    }
}

@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.onboarding

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
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
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.R
import com.example.bero.ui.theme.*

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
                        LuxuryCharcoal,
                        LuxuryOnyx,
                        LuxuryCharcoal
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
            Spacer(modifier = Modifier.height(80.dp))

            // Bero Logo
            Image(
                painter = painterResource(id = R.drawable.bero_logo),
                contentDescription = "Bero Logo",
                modifier = Modifier
                    .size(90.dp)
                    .clip(RoundedCornerShape(18.dp))
                    .border(
                        width = 1.5.dp,
                        color = LuxuryGold.copy(alpha = 0.4f),
                        shape = RoundedCornerShape(18.dp)
                    ),
                contentScale = ContentScale.Crop
            )

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Welcome to Bero",
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold,
                color = Color(0xFFF5F0E8)
            )

            Text(
                text = "Bero में आपका स्वागत है",
                style = MaterialTheme.typography.bodyLarge,
                color = Color(0xFFF5F0E8).copy(alpha = 0.4f),
                modifier = Modifier.padding(top = 4.dp)
            )

            Spacer(modifier = Modifier.height(48.dp))

            Text(
                text = "Choose your language",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Medium,
                color = Color(0xFFF5F0E8).copy(alpha = 0.8f)
            )

            Text(
                text = "अपनी भाषा चुनें",
                style = MaterialTheme.typography.bodyMedium,
                color = Color(0xFFF5F0E8).copy(alpha = 0.4f),
                modifier = Modifier.padding(top = 4.dp)
            )

            Spacer(modifier = Modifier.height(32.dp))

            // Language Cards
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                AppLanguage.entries.forEach { language ->
                    LuxuryLanguageCard(
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
                    containerColor = LuxuryGold,
                    contentColor = LuxuryCharcoal,
                    disabledContainerColor = LuxuryGold.copy(alpha = 0.2f),
                    disabledContentColor = LuxuryCharcoal.copy(alpha = 0.4f)
                )
            ) {
                Text(
                    text = if (selectedLanguage == AppLanguage.HINDI) "जारी रखें" else "CONTINUE",
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold,
                    letterSpacing = if (selectedLanguage != AppLanguage.HINDI) 2.sp else 0.sp
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            Text(
                text = "You can change this later in Settings",
                fontSize = 13.sp,
                color = Color(0xFFF5F0E8).copy(alpha = 0.3f),
                textAlign = TextAlign.Center
            )

            Text(
                text = "आप इसे बाद में सेटिंग्स में बदल सकते हैं",
                fontSize = 12.sp,
                color = Color(0xFFF5F0E8).copy(alpha = 0.2f),
                textAlign = TextAlign.Center,
                modifier = Modifier.padding(top = 4.dp)
            )

            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@Composable
private fun LuxuryLanguageCard(
    language: AppLanguage,
    isSelected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val scale by animateFloatAsState(
        targetValue = if (isSelected) 1.02f else 1f,
        animationSpec = tween(200),
        label = "scale"
    )

    val borderColor by animateColorAsState(
        targetValue = if (isSelected) LuxuryGold else Color(0xFFF5F0E8).copy(alpha = 0.1f),
        animationSpec = tween(200),
        label = "borderColor"
    )

    Card(
        modifier = modifier
            .scale(scale)
            .aspectRatio(0.85f)
            .clickable { onClick() },
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(
            containerColor = if (isSelected)
                LuxuryGold.copy(alpha = 0.08f)
            else
                Color(0xFF1E1E32).copy(alpha = 0.7f)
        ),
        border = BorderStroke(
            width = if (isSelected) 2.dp else 1.dp,
            color = borderColor
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
                // Language icon
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .background(
                            color = if (isSelected) LuxuryGold.copy(alpha = 0.15f)
                            else Color(0xFFF5F0E8).copy(alpha = 0.05f),
                            shape = RoundedCornerShape(16.dp)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = if (language == AppLanguage.ENGLISH) "A" else "अ",
                        fontSize = 26.sp,
                        fontWeight = FontWeight.Bold,
                        color = if (isSelected) LuxuryGold else Color(0xFFF5F0E8).copy(alpha = 0.6f)
                    )
                }

                Spacer(modifier = Modifier.height(16.dp))

                Text(
                    text = language.nativeName,
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFFF5F0E8)
                )

                if (language != AppLanguage.ENGLISH) {
                    Text(
                        text = language.displayName,
                        fontSize = 13.sp,
                        color = Color(0xFFF5F0E8).copy(alpha = 0.4f),
                        modifier = Modifier.padding(top = 2.dp)
                    )
                }

                Spacer(modifier = Modifier.height(12.dp))

                Text(
                    text = language.greeting,
                    fontSize = 16.sp,
                    color = LuxuryGold,
                    fontWeight = FontWeight.Medium
                )
            }

            // Checkmark
            if (isSelected) {
                Box(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(12.dp)
                        .size(26.dp)
                        .clip(CircleShape)
                        .background(LuxuryGold),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        imageVector = Icons.Default.Check,
                        contentDescription = "Selected",
                        tint = LuxuryCharcoal,
                        modifier = Modifier.size(16.dp)
                    )
                }
            }
        }
    }
}

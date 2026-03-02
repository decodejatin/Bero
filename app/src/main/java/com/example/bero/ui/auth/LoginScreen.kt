package com.example.bero.ui.auth

import androidx.compose.animation.*
import androidx.compose.animation.core.*
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.R
import com.example.bero.ui.theme.*

/**
 * Premium Login Screen — dark luxury background, frosted card, gold accents
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LoginScreen(
    onPhoneSubmit: (String) -> Unit,
    onTruecallerClick: () -> Unit,
    isLoading: Boolean = false,
    error: String? = null
) {
    var phoneNumber by remember { mutableStateOf("") }
    val focusManager = LocalFocusManager.current
    val focusRequester = remember { FocusRequester() }

    // Subtle shimmer animation
    val infiniteTransition = rememberInfiniteTransition(label = "shimmer")
    val shimmerAlpha by infiniteTransition.animateFloat(
        initialValue = 0.03f,
        targetValue = 0.08f,
        animationSpec = infiniteRepeatable(
            animation = tween(3000, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "shimmerAlpha"
    )

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
        // Subtle radial gold shimmer overlay
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    brush = Brush.radialGradient(
                        colors = listOf(
                            LuxuryGold.copy(alpha = shimmerAlpha),
                            Color.Transparent
                        )
                    )
                )
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Spacer(modifier = Modifier.weight(0.3f))

            // Logo with gold ring
            Box(contentAlignment = Alignment.Center) {
                // Gold border ring
                Box(
                    modifier = Modifier
                        .size(124.dp)
                        .clip(RoundedCornerShape(24.dp))
                        .border(
                            width = 2.dp,
                            brush = Brush.linearGradient(
                                colors = listOf(
                                    LuxuryGold.copy(alpha = 0.8f),
                                    LuxuryGoldLight.copy(alpha = 0.4f),
                                    LuxuryGold.copy(alpha = 0.8f)
                                )
                            ),
                            shape = RoundedCornerShape(24.dp)
                        )
                )
                Image(
                    painter = painterResource(id = R.drawable.bero_logo),
                    contentDescription = "Bero Logo",
                    modifier = Modifier
                        .size(110.dp)
                        .clip(RoundedCornerShape(20.dp)),
                    contentScale = ContentScale.Crop
                )
            }

            Spacer(modifier = Modifier.height(20.dp))

            Text(
                text = "BERO",
                style = MaterialTheme.typography.headlineLarge,
                fontWeight = FontWeight.ExtraBold,
                color = LuxuryGold,
                letterSpacing = 6.sp
            )

            Spacer(modifier = Modifier.height(4.dp))

            Text(
                text = "Your trusted local services",
                style = MaterialTheme.typography.bodyMedium,
                color = Color(0xFFF5F0E8).copy(alpha = 0.5f),
                letterSpacing = 1.sp
            )

            Spacer(modifier = Modifier.height(40.dp))

            // Frosted Login Card
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(24.dp),
                colors = CardDefaults.cardColors(
                    containerColor = Color(0xFF1E1E32).copy(alpha = 0.85f)
                ),
                border = androidx.compose.foundation.BorderStroke(
                    1.dp,
                    LuxuryGold.copy(alpha = 0.15f)
                )
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text(
                        text = "Enter your phone number",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = Color(0xFFF5F0E8)
                    )

                    Spacer(modifier = Modifier.height(4.dp))

                    Text(
                        text = "We'll send you an OTP to verify",
                        style = MaterialTheme.typography.bodySmall,
                        color = Color(0xFFF5F0E8).copy(alpha = 0.5f)
                    )

                    Spacer(modifier = Modifier.height(24.dp))

                    // Phone Number Input with gold focus
                    OutlinedTextField(
                        value = phoneNumber,
                        onValueChange = {
                            if (it.length <= 10 && it.all { char -> char.isDigit() }) {
                                phoneNumber = it
                            }
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .focusRequester(focusRequester),
                        label = { Text("Phone Number") },
                        placeholder = { Text("10-digit mobile number") },
                        prefix = { Text("+91 ", color = LuxuryGold.copy(alpha = 0.7f)) },
                        singleLine = true,
                        keyboardOptions = KeyboardOptions(
                            keyboardType = KeyboardType.Phone,
                            imeAction = ImeAction.Done
                        ),
                        keyboardActions = KeyboardActions(
                            onDone = {
                                focusManager.clearFocus()
                                if (phoneNumber.length == 10) {
                                    onPhoneSubmit("+91$phoneNumber")
                                }
                            }
                        ),
                        shape = RoundedCornerShape(14.dp),
                        enabled = !isLoading,
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = LuxuryGold,
                            unfocusedBorderColor = Color(0xFFF5F0E8).copy(alpha = 0.15f),
                            focusedLabelColor = LuxuryGold,
                            unfocusedLabelColor = Color(0xFFF5F0E8).copy(alpha = 0.4f),
                            focusedTextColor = Color(0xFFF5F0E8),
                            unfocusedTextColor = Color(0xFFF5F0E8).copy(alpha = 0.8f),
                            cursorColor = LuxuryGold,
                            focusedPlaceholderColor = Color(0xFFF5F0E8).copy(alpha = 0.3f),
                            unfocusedPlaceholderColor = Color(0xFFF5F0E8).copy(alpha = 0.2f)
                        )
                    )

                    // Error message
                    AnimatedVisibility(visible = error != null) {
                        error?.let {
                            Text(
                                text = it,
                                color = BeroError,
                                style = MaterialTheme.typography.bodySmall,
                                modifier = Modifier.padding(top = 8.dp)
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(24.dp))

                    // Gold Gradient "Get OTP" Button
                    Button(
                        onClick = { onPhoneSubmit("+91$phoneNumber") },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(52.dp),
                        enabled = phoneNumber.length == 10 && !isLoading,
                        shape = RoundedCornerShape(14.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = LuxuryGold,
                            contentColor = LuxuryCharcoal,
                            disabledContainerColor = LuxuryGold.copy(alpha = 0.3f),
                            disabledContentColor = LuxuryCharcoal.copy(alpha = 0.5f)
                        )
                    ) {
                        if (isLoading) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(24.dp),
                                color = LuxuryCharcoal,
                                strokeWidth = 2.dp
                            )
                        } else {
                            Text(
                                text = "GET OTP",
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Bold,
                                letterSpacing = 2.sp
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(16.dp))

                    // Divider
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        HorizontalDivider(
                            modifier = Modifier.weight(1f),
                            color = Color(0xFFF5F0E8).copy(alpha = 0.1f)
                        )
                        Text(
                            text = "  or  ",
                            style = MaterialTheme.typography.bodySmall,
                            color = Color(0xFFF5F0E8).copy(alpha = 0.3f)
                        )
                        HorizontalDivider(
                            modifier = Modifier.weight(1f),
                            color = Color(0xFFF5F0E8).copy(alpha = 0.1f)
                        )
                    }

                    Spacer(modifier = Modifier.height(16.dp))

                    // Truecaller Button
                    OutlinedButton(
                        onClick = onTruecallerClick,
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(52.dp),
                        enabled = !isLoading,
                        shape = RoundedCornerShape(14.dp),
                        border = androidx.compose.foundation.BorderStroke(
                            1.dp,
                            LuxuryGold.copy(alpha = 0.3f)
                        ),
                        colors = ButtonDefaults.outlinedButtonColors(
                            containerColor = Color.Transparent
                        )
                    ) {
                        Text(
                            text = "📱 Continue with Truecaller",
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Medium,
                            color = LuxuryGold
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            // Terms text
            Text(
                text = "By continuing, you agree to our\nTerms of Service and Privacy Policy",
                style = MaterialTheme.typography.bodySmall,
                color = Color(0xFFF5F0E8).copy(alpha = 0.3f),
                textAlign = TextAlign.Center,
                lineHeight = 18.sp
            )

            Spacer(modifier = Modifier.height(16.dp))
        }
    }
}

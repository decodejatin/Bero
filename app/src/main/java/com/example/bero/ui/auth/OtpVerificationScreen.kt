package com.example.bero.ui.auth

import androidx.compose.animation.*
import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.scale
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.ui.theme.*
import kotlinx.coroutines.delay

/**
 * Premium OTP Verification Screen — dark luxury design with gold accents
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OtpVerificationScreen(
    phoneNumber: String,
    onOtpSubmit: (String) -> Unit,
    onResendOtp: () -> Unit,
    onBackClick: () -> Unit,
    isLoading: Boolean = false,
    error: String? = null
) {
    var otp by remember { mutableStateOf("") }
    var resendTimer by remember { mutableIntStateOf(30) }
    val focusRequester = remember { FocusRequester() }

    // Auto-submit when 6 digits entered
    LaunchedEffect(otp) {
        if (otp.length == 6) {
            onOtpSubmit(otp)
        }
    }

    // Countdown timer
    LaunchedEffect(resendTimer) {
        if (resendTimer > 0) {
            delay(1000)
            resendTimer--
        }
    }

    LaunchedEffect(Unit) {
        focusRequester.requestFocus()
    }

    // Checkmark animation
    val isComplete = otp.length == 6
    val checkScale by animateFloatAsState(
        targetValue = if (isComplete) 1f else 0f,
        animationSpec = spring(dampingRatio = 0.5f, stiffness = 300f),
        label = "checkScale"
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
        // Back button
        IconButton(
            onClick = onBackClick,
            modifier = Modifier
                .padding(16.dp)
                .align(Alignment.TopStart),
            enabled = !isLoading
        ) {
            Icon(
                imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                contentDescription = "Back",
                tint = LuxuryGold
            )
        }

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            // Animated checkmark or phone icon
            Box(
                modifier = Modifier.size(80.dp),
                contentAlignment = Alignment.Center
            ) {
                if (isComplete) {
                    Icon(
                        Icons.Default.CheckCircle,
                        contentDescription = null,
                        modifier = Modifier
                            .size(64.dp)
                            .scale(checkScale),
                        tint = LuxuryGold
                    )
                } else {
                    Text("📱", fontSize = 52.sp)
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            Text(
                text = "VERIFY OTP",
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold,
                color = Color(0xFFF5F0E8),
                letterSpacing = 3.sp
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "Code sent to ${formatPhoneNumber(phoneNumber)}",
                style = MaterialTheme.typography.bodyMedium,
                color = Color(0xFFF5F0E8).copy(alpha = 0.5f)
            )

            Spacer(modifier = Modifier.height(40.dp))

            // OTP Card
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
                        text = "Enter 6-digit code",
                        style = MaterialTheme.typography.titleSmall,
                        color = Color(0xFFF5F0E8).copy(alpha = 0.6f)
                    )

                    Spacer(modifier = Modifier.height(24.dp))

                    // Premium OTP Input
                    OtpInputField(
                        otp = otp,
                        onOtpChange = {
                            if (it.length <= 6 && it.all { char -> char.isDigit() }) {
                                otp = it
                            }
                        },
                        focusRequester = focusRequester,
                        enabled = !isLoading
                    )

                    // Error
                    AnimatedVisibility(visible = error != null) {
                        error?.let {
                            Text(
                                text = it,
                                color = BeroError,
                                style = MaterialTheme.typography.bodySmall,
                                modifier = Modifier.padding(top = 16.dp)
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(28.dp))

                    // Verify Button
                    Button(
                        onClick = { onOtpSubmit(otp) },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(52.dp),
                        enabled = otp.length == 6 && !isLoading,
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
                                text = "VERIFY",
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Bold,
                                letterSpacing = 2.sp
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(20.dp))

                    // Resend
                    if (resendTimer > 0) {
                        Text(
                            text = "Resend OTP in ${resendTimer}s",
                            style = MaterialTheme.typography.bodySmall,
                            color = Color(0xFFF5F0E8).copy(alpha = 0.4f)
                        )
                    } else {
                        TextButton(
                            onClick = {
                                onResendOtp()
                                resendTimer = 30
                            },
                            enabled = !isLoading
                        ) {
                            Text(
                                text = "Resend OTP",
                                fontSize = 14.sp,
                                fontWeight = FontWeight.Medium,
                                color = LuxuryGold
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun OtpInputField(
    otp: String,
    onOtpChange: (String) -> Unit,
    focusRequester: FocusRequester,
    enabled: Boolean
) {
    BasicTextField(
        value = otp,
        onValueChange = onOtpChange,
        modifier = Modifier.focusRequester(focusRequester),
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
        enabled = enabled,
        decorationBox = {
            Row(
                horizontalArrangement = Arrangement.spacedBy(10.dp),
                modifier = Modifier.fillMaxWidth()
            ) {
                repeat(6) { index ->
                    OtpBox(
                        char = otp.getOrNull(index),
                        isFocused = otp.length == index,
                        modifier = Modifier.weight(1f)
                    )
                }
            }
        }
    )
}

@Composable
private fun OtpBox(
    char: Char?,
    isFocused: Boolean,
    modifier: Modifier = Modifier
) {
    val borderColor by animateColorAsState(
        targetValue = when {
            char != null -> LuxuryGold
            isFocused -> LuxuryGold.copy(alpha = 0.7f)
            else -> Color(0xFFF5F0E8).copy(alpha = 0.1f)
        },
        animationSpec = tween(200),
        label = "otpBorderColor"
    )

    val bgColor by animateColorAsState(
        targetValue = if (char != null) {
            LuxuryGold.copy(alpha = 0.08f)
        } else {
            Color.Transparent
        },
        animationSpec = tween(200),
        label = "otpBgColor"
    )

    Box(
        modifier = modifier
            .aspectRatio(0.85f)
            .border(
                width = if (isFocused || char != null) 2.dp else 1.dp,
                color = borderColor,
                shape = RoundedCornerShape(14.dp)
            )
            .background(
                color = bgColor,
                shape = RoundedCornerShape(14.dp)
            ),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = char?.toString() ?: "",
            style = TextStyle(
                fontSize = 24.sp,
                fontWeight = FontWeight.Bold,
                color = Color(0xFFF5F0E8)
            ),
            textAlign = TextAlign.Center
        )
    }
}

private fun formatPhoneNumber(phone: String): String {
    val cleaned = phone.replace("+91", "").replace(" ", "")
    return if (cleaned.length == 10) {
        "+91 ${cleaned.take(5)}•••••"
    } else {
        phone
    }
}

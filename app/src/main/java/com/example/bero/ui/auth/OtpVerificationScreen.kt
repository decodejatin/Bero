package com.example.bero.ui.auth

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
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
import kotlinx.coroutines.delay

/**
 * OTP Verification screen with 6-digit input
 * Clean, focused design for easy OTP entry
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
    
    // Countdown timer for resend
    LaunchedEffect(resendTimer) {
        if (resendTimer > 0) {
            delay(1000)
            resendTimer--
        }
    }
    
    // Request focus on OTP field
    LaunchedEffect(Unit) {
        focusRequester.requestFocus()
    }
    
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                brush = Brush.verticalGradient(
                    colors = listOf(
                        MaterialTheme.colorScheme.primary,
                        MaterialTheme.colorScheme.primaryContainer
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
                tint = MaterialTheme.colorScheme.onPrimary
            )
        }
        
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            // Title
            Text(
                text = "📱",
                fontSize = 64.sp
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            Text(
                text = "Verify OTP",
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onPrimary
            )
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Text(
                text = "We sent a code to ${formatPhoneNumber(phoneNumber)}",
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onPrimary.copy(alpha = 0.8f),
                textAlign = TextAlign.Center
            )
            
            Spacer(modifier = Modifier.height(48.dp))
            
            // OTP Input Card
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(24.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surface
                ),
                elevation = CardDefaults.cardElevation(defaultElevation = 8.dp)
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text(
                        text = "Enter 6-digit OTP",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                    
                    Spacer(modifier = Modifier.height(24.dp))
                    
                    // OTP Input Boxes
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
                    
                    // Error message
                    AnimatedVisibility(visible = error != null) {
                        error?.let {
                            Text(
                                text = it,
                                color = MaterialTheme.colorScheme.error,
                                style = MaterialTheme.typography.bodySmall,
                                modifier = Modifier.padding(top = 16.dp)
                            )
                        }
                    }
                    
                    Spacer(modifier = Modifier.height(24.dp))
                    
                    // Verify Button
                    Button(
                        onClick = { onOtpSubmit(otp) },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(52.dp),
                        enabled = otp.length == 6 && !isLoading,
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        if (isLoading) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(24.dp),
                                color = MaterialTheme.colorScheme.onPrimary,
                                strokeWidth = 2.dp
                            )
                        } else {
                            Text(
                                text = "Verify",
                                fontSize = 16.sp,
                                fontWeight = FontWeight.SemiBold
                            )
                        }
                    }
                    
                    Spacer(modifier = Modifier.height(24.dp))
                    
                    // Resend OTP
                    if (resendTimer > 0) {
                        Text(
                            text = "Resend OTP in ${resendTimer}s",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
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
                                fontWeight = FontWeight.Medium
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
                horizontalArrangement = Arrangement.spacedBy(8.dp),
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
    Box(
        modifier = modifier
            .aspectRatio(1f)
            .border(
                width = 2.dp,
                color = when {
                    isFocused -> MaterialTheme.colorScheme.primary
                    char != null -> MaterialTheme.colorScheme.primary.copy(alpha = 0.5f)
                    else -> MaterialTheme.colorScheme.outline.copy(alpha = 0.3f)
                },
                shape = RoundedCornerShape(12.dp)
            )
            .background(
                color = if (char != null) {
                    MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.2f)
                } else {
                    Color.Transparent
                },
                shape = RoundedCornerShape(12.dp)
            ),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = char?.toString() ?: "",
            style = TextStyle(
                fontSize = 24.sp,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onSurface
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

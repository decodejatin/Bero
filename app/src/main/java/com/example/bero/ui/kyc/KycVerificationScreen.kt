package com.example.bero.ui.kyc

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

/**
 * KYC Verification Screen - Multi-step Aadhaar verification flow
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun KycVerificationScreen(
    onComplete: () -> Unit,
    onBackClick: () -> Unit,
    viewModel: KycViewModel
) {
    val uiState by viewModel.uiState.collectAsState()
    
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            // Top Bar
            TopAppBar(
                title = { Text("Aadhaar Verification") },
                navigationIcon = {
                    IconButton(onClick = onBackClick, enabled = !uiState.isLoading) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "Back"
                        )
                    }
                }
            )
            
            // Content based on step
            when (uiState.step) {
                KycStep.ENTER_AADHAAR -> EnterAadhaarStep(
                    onSubmit = { aadhaar -> viewModel.requestAadhaarOtp(aadhaar) },
                    isLoading = uiState.isLoading,
                    error = uiState.error
                )
                KycStep.VERIFY_OTP -> VerifyAadhaarOtpStep(
                    maskedAadhaar = uiState.maskedAadhaar ?: "",
                    onSubmit = { otp -> viewModel.verifyOtp(otp) },
                    onResend = { viewModel.resendOtp() },
                    isLoading = uiState.isLoading,
                    error = uiState.error
                )
                KycStep.SUCCESS -> KycSuccessStep(
                    fullName = uiState.verifiedName ?: "User",
                    onContinue = onComplete
                )
                KycStep.FAILED -> KycFailedStep(
                    error = uiState.error ?: "Verification failed",
                    onRetry = { viewModel.reset() }
                )
            }
        }
    }
}

@Composable
private fun EnterAadhaarStep(
    onSubmit: (String) -> Unit,
    isLoading: Boolean,
    error: String?
) {
    var aadhaarNumber by remember { mutableStateOf("") }
    val focusManager = LocalFocusManager.current
    
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Spacer(modifier = Modifier.height(32.dp))
        
        // Icon
        Text(
            text = "🪪",
            fontSize = 64.sp
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Text(
            text = "Enter Aadhaar Number",
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = "Enter your 12-digit Aadhaar number.\nWe'll send an OTP to your registered mobile.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(32.dp))
        
        // Aadhaar Input
        OutlinedTextField(
            value = formatAadhaarInput(aadhaarNumber),
            onValueChange = { 
                val digits = it.filter { char -> char.isDigit() }
                if (digits.length <= 12) {
                    aadhaarNumber = digits
                }
            },
            modifier = Modifier.fillMaxWidth(),
            label = { Text("Aadhaar Number") },
            placeholder = { Text("XXXX XXXX XXXX") },
            singleLine = true,
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Number,
                imeAction = ImeAction.Done
            ),
            keyboardActions = KeyboardActions(
                onDone = {
                    focusManager.clearFocus()
                    if (aadhaarNumber.length == 12) {
                        onSubmit(aadhaarNumber)
                    }
                }
            ),
            shape = RoundedCornerShape(12.dp),
            enabled = !isLoading
        )
        
        // Error message
        AnimatedVisibility(visible = error != null) {
            error?.let {
                Text(
                    text = it,
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.padding(top = 8.dp)
                )
            }
        }
        
        Spacer(modifier = Modifier.height(24.dp))
        
        // Submit Button
        Button(
            onClick = { onSubmit(aadhaarNumber) },
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp),
            enabled = aadhaarNumber.length == 12 && !isLoading,
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
                    text = "Get OTP",
                    fontSize = 16.sp,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
        
        Spacer(modifier = Modifier.height(32.dp))
        
        // Info Card
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.secondaryContainer.copy(alpha = 0.5f)
            )
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text(
                    text = "🔒 Your data is secure",
                    fontWeight = FontWeight.SemiBold,
                    style = MaterialTheme.typography.bodyMedium
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "We use government approved Digilocker APIs. Your Aadhaar details are encrypted and not stored.",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSecondaryContainer.copy(alpha = 0.8f)
                )
            }
        }
    }
}

@Composable
private fun VerifyAadhaarOtpStep(
    maskedAadhaar: String,
    onSubmit: (String) -> Unit,
    onResend: () -> Unit,
    isLoading: Boolean,
    error: String?
) {
    var otp by remember { mutableStateOf("") }
    var resendTimer by remember { mutableIntStateOf(30) }
    val focusManager = LocalFocusManager.current
    
    // Countdown timer
    LaunchedEffect(resendTimer) {
        if (resendTimer > 0) {
            kotlinx.coroutines.delay(1000)
            resendTimer--
        }
    }
    
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Spacer(modifier = Modifier.height(32.dp))
        
        Text(
            text = "📱",
            fontSize = 64.sp
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Text(
            text = "Verify OTP",
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = "Enter the OTP sent to mobile\nregistered with Aadhaar $maskedAadhaar",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(32.dp))
        
        // OTP Input
        OutlinedTextField(
            value = otp,
            onValueChange = { 
                if (it.length <= 6 && it.all { char -> char.isDigit() }) {
                    otp = it
                }
            },
            modifier = Modifier.fillMaxWidth(),
            label = { Text("Enter OTP") },
            placeholder = { Text("6-digit OTP") },
            singleLine = true,
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Number,
                imeAction = ImeAction.Done
            ),
            keyboardActions = KeyboardActions(
                onDone = {
                    focusManager.clearFocus()
                    if (otp.length == 6) {
                        onSubmit(otp)
                    }
                }
            ),
            shape = RoundedCornerShape(12.dp),
            enabled = !isLoading
        )
        
        // Error message
        AnimatedVisibility(visible = error != null) {
            error?.let {
                Text(
                    text = it,
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.padding(top = 8.dp)
                )
            }
        }
        
        Spacer(modifier = Modifier.height(24.dp))
        
        // Verify Button
        Button(
            onClick = { onSubmit(otp) },
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
        
        Spacer(modifier = Modifier.height(16.dp))
        
        // Resend OTP
        if (resendTimer > 0) {
            Text(
                text = "Resend OTP in ${resendTimer}s",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f)
            )
        } else {
            TextButton(
                onClick = { 
                    onResend()
                    resendTimer = 30
                },
                enabled = !isLoading
            ) {
                Text("Resend OTP")
            }
        }
    }
}

@Composable
private fun KycSuccessStep(
    fullName: String,
    onContinue: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            imageVector = Icons.Default.CheckCircle,
            contentDescription = "Success",
            modifier = Modifier.size(80.dp),
            tint = Color(0xFF4CAF50)
        )
        
        Spacer(modifier = Modifier.height(24.dp))
        
        Text(
            text = "KYC Verified!",
            style = MaterialTheme.typography.headlineMedium,
            fontWeight = FontWeight.Bold,
            color = Color(0xFF4CAF50)
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = "Welcome, $fullName!",
            style = MaterialTheme.typography.titleLarge,
            fontWeight = FontWeight.Medium
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = "Your Aadhaar has been verified successfully.\nYou can now start accepting jobs!",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(48.dp))
        
        Button(
            onClick = onContinue,
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp),
            shape = RoundedCornerShape(12.dp)
        ) {
            Text(
                text = "Continue",
                fontSize = 16.sp,
                fontWeight = FontWeight.SemiBold
            )
        }
    }
}

@Composable
private fun KycFailedStep(
    error: String,
    onRetry: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Text(
            text = "❌",
            fontSize = 64.sp
        )
        
        Spacer(modifier = Modifier.height(24.dp))
        
        Text(
            text = "Verification Failed",
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold,
            color = MaterialTheme.colorScheme.error
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = error,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(48.dp))
        
        Button(
            onClick = onRetry,
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp),
            shape = RoundedCornerShape(12.dp)
        ) {
            Text(
                text = "Try Again",
                fontSize = 16.sp,
                fontWeight = FontWeight.SemiBold
            )
        }
    }
}

/**
 * Format Aadhaar input with spaces (XXXX XXXX XXXX)
 */
private fun formatAadhaarInput(input: String): String {
    val clean = input.filter { it.isDigit() }
    return buildString {
        clean.forEachIndexed { index, char ->
            if (index > 0 && index % 4 == 0) append(' ')
            append(char)
        }
    }
}

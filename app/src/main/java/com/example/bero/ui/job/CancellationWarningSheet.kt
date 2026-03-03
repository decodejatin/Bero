@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.job

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

/**
 * Cancellation Warning Sheet — Bottom sheet when user taps Cancel.
 * Shows penalty info, cooldown warning, and cancellation count.
 * Maps to backend: GET /stability/user/:id/status
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CancellationWarningSheet(
    cancellationCount: Int, // today's count
    maxDailyCancellations: Int = 3,
    cooldownMinutes: Int = 0, // 0 = no active cooldown
    penaltyMessage: String = "",
    onConfirmCancel: () -> Unit = {},
    onKeepJob: () -> Unit = {},
    onDismiss: () -> Unit = {}
) {
    val remainingCancellations = maxDailyCancellations - cancellationCount
    val isNearLimit = remainingCancellations <= 1
    val isCoolingDown = cooldownMinutes > 0

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        shape = RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(24.dp)
        ) {
            // Warning icon
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    Icons.Default.Warning,
                    contentDescription = null,
                    tint = if (isNearLimit) Color(0xFFF44336) else Color(0xFFFF9800),
                    modifier = Modifier.size(28.dp)
                )
                Spacer(modifier = Modifier.width(12.dp))
                Text(
                    if (isCoolingDown) "Cancellation Cooldown Active" else "Cancel This Job?",
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Cooldown alert
            if (isCoolingDown) {
                Card(
                    shape = RoundedCornerShape(12.dp),
                    colors = CardDefaults.cardColors(containerColor = Color(0xFFFFEBEE))
                ) {
                    Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Icon(Icons.Default.Timer, contentDescription = null, tint = Color(0xFFF44336), modifier = Modifier.size(20.dp))
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            "You are in a cooldown period. You cannot cancel for $cooldownMinutes more minutes.",
                            fontSize = 13.sp,
                            color = Color(0xFFC62828)
                        )
                    }
                }
                Spacer(modifier = Modifier.height(12.dp))
            }

            // Cancellation count
            Card(
                shape = RoundedCornerShape(12.dp),
                colors = CardDefaults.cardColors(
                    containerColor = if (isNearLimit) Color(0xFFFFF3E0) else MaterialTheme.colorScheme.surfaceVariant
                )
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text("Cancellations Today", fontSize = 14.sp)
                        Text(
                            "$cancellationCount / $maxDailyCancellations",
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Bold,
                            color = if (isNearLimit) Color(0xFFF44336) else MaterialTheme.colorScheme.onSurface
                        )
                    }

                    if (isNearLimit) {
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            "⚠️ You have $remainingCancellations cancellation${if (remainingCancellations != 1) "s" else ""} left today. Exceeding the limit will result in a temporary suspension.",
                            fontSize = 12.sp,
                            color = Color(0xFFE65100),
                            lineHeight = 16.sp
                        )
                    }
                }
            }

            // Custom penalty
            if (penaltyMessage.isNotEmpty()) {
                Spacer(modifier = Modifier.height(12.dp))
                Card(
                    shape = RoundedCornerShape(12.dp),
                    colors = CardDefaults.cardColors(containerColor = Color(0xFFFCE4EC))
                ) {
                    Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Icon(Icons.Default.MoneyOff, contentDescription = null, tint = Color(0xFFF44336), modifier = Modifier.size(20.dp))
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(penaltyMessage, fontSize = 13.sp, color = Color(0xFFC62828))
                    }
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            // Buttons
            Button(
                onClick = onKeepJob,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(52.dp),
                shape = RoundedCornerShape(12.dp)
            ) {
                Text("Keep This Job", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
            }

            Spacer(modifier = Modifier.height(8.dp))

            OutlinedButton(
                onClick = onConfirmCancel,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(52.dp),
                shape = RoundedCornerShape(12.dp),
                enabled = !isCoolingDown,
                colors = ButtonDefaults.outlinedButtonColors(
                    contentColor = Color(0xFFF44336)
                )
            ) {
                Text(
                    if (isCoolingDown) "Cooldown Active" else "Yes, Cancel Job",
                    fontSize = 16.sp,
                    fontWeight = FontWeight.SemiBold
                )
            }

            Spacer(modifier = Modifier.height(16.dp))
        }
    }
}

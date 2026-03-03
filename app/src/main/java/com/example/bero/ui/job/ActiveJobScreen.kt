@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.job

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.models.Job
import com.example.bero.data.models.JobStatus

/**
 * Active Job Screen — Worker's view of their current in-progress job.
 * Shows: job details, elapsed timer, surge badge, "Mark Complete" button.
 * Maps to backend: POST /jobs/:id/complete-by-worker
 */
@Composable
fun ActiveJobScreen(
    job: Job,
    onBackClick: () -> Unit = {},
    onMarkComplete: (String) -> Unit = {},
    onChatClick: (String) -> Unit = {},
    onCallClient: (String) -> Unit = {}
) {
    var showCompleteDialog by remember { mutableStateOf(false) }
    var isLoading by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Active Job") },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.surface
                )
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState())
                .padding(16.dp)
        ) {
            // Status badge
            StatusBadge(job.status)

            Spacer(modifier = Modifier.height(16.dp))

            // Job title + category
            Text(
                text = job.title,
                fontSize = 22.sp,
                fontWeight = FontWeight.Bold
            )
            Text(
                text = job.category.displayName,
                fontSize = 14.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Surge indicator
            if (job.hasSurge) {
                SurgeBadge(
                    multiplier = job.surgeMultiplier ?: 1.0,
                    basePrice = job.amountRupees,
                    surgePrice = job.surgePrice ?: job.amountRupees
                )
                Spacer(modifier = Modifier.height(12.dp))
            }

            // Client info card
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant
                )
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Box(
                        modifier = Modifier
                            .size(48.dp)
                            .clip(CircleShape)
                            .background(MaterialTheme.colorScheme.primary),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = job.clientName.first().uppercase(),
                            color = MaterialTheme.colorScheme.onPrimary,
                            fontWeight = FontWeight.Bold,
                            fontSize = 20.sp
                        )
                    }
                    Spacer(modifier = Modifier.width(12.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(job.clientName, fontWeight = FontWeight.SemiBold)
                        Text(
                            "Client",
                            fontSize = 12.sp,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                    }
                    IconButton(onClick = { onChatClick(job.clientId) }) {
                        Icon(Icons.Default.Chat, contentDescription = "Chat", tint = MaterialTheme.colorScheme.primary)
                    }
                    IconButton(onClick = { onCallClient(job.clientId) }) {
                        Icon(Icons.Default.Phone, contentDescription = "Call", tint = MaterialTheme.colorScheme.primary)
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Job details
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp)
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    ActiveJobDetail(Icons.Default.LocationOn, "Location", job.address.ifEmpty { job.location })
                    ActiveJobDetail(Icons.Default.Schedule, "Scheduled", "${job.scheduledDate} • ${job.scheduledTimeSlot}")
                    ActiveJobDetail(Icons.Default.Timer, "Duration", "${job.estimatedDurationMinutes} minutes")
                    ActiveJobDetail(
                        Icons.Default.CurrencyRupee,
                        "Payment",
                        "₹${String.format("%.0f", job.displayPrice)}"
                    )
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            // Description
            if (job.description.isNotEmpty()) {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text("Description", fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(job.description, fontSize = 14.sp, lineHeight = 20.sp)
                    }
                }
            }

            Spacer(modifier = Modifier.weight(1f))
            Spacer(modifier = Modifier.height(24.dp))

            // Mark Complete button
            if (job.status == JobStatus.IN_PROGRESS || job.status == JobStatus.ASSIGNED) {
                Button(
                    onClick = { showCompleteDialog = true },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(56.dp),
                    shape = RoundedCornerShape(16.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = Color(0xFF4CAF50)
                    ),
                    enabled = !isLoading
                ) {
                    if (isLoading) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(24.dp),
                            color = Color.White,
                            strokeWidth = 2.dp
                        )
                    } else {
                        Icon(Icons.Default.CheckCircle, contentDescription = null)
                        Spacer(modifier = Modifier.width(8.dp))
                        Text("Mark as Completed", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    }
                }
            }

            // Already completed - waiting for client
            if (job.status == JobStatus.WORKER_COMPLETED) {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(
                        containerColor = Color(0xFFFFF3E0)
                    )
                ) {
                    Row(
                        modifier = Modifier.padding(16.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(Icons.Default.HourglassTop, contentDescription = null, tint = Color(0xFFFF9800))
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text("Completion Submitted", fontWeight = FontWeight.SemiBold)
                            Text("Waiting for client to confirm", fontSize = 13.sp, color = Color(0xFF795548))
                        }
                    }
                }
            }
        }
    }

    // Confirmation dialog
    if (showCompleteDialog) {
        AlertDialog(
            onDismissRequest = { showCompleteDialog = false },
            title = { Text("Mark Job Complete?") },
            text = { Text("Are you sure this job is finished? The client will be asked to confirm.") },
            confirmButton = {
                TextButton(onClick = {
                    showCompleteDialog = false
                    isLoading = true
                    onMarkComplete(job.id)
                }) {
                    Text("Yes, Complete")
                }
            },
            dismissButton = {
                TextButton(onClick = { showCompleteDialog = false }) {
                    Text("Cancel")
                }
            }
        )
    }
}

@Composable
private fun StatusBadge(status: JobStatus) {
    val (color, label) = when (status) {
        JobStatus.IN_PROGRESS -> Color(0xFF2196F3) to "In Progress"
        JobStatus.ASSIGNED -> Color(0xFF9C27B0) to "Assigned"
        JobStatus.WORKER_COMPLETED -> Color(0xFFFF9800) to "Waiting for Client"
        JobStatus.FULLY_COMPLETED -> Color(0xFF4CAF50) to "Fully Completed"
        else -> MaterialTheme.colorScheme.outline to status.name
    }
    Surface(
        shape = RoundedCornerShape(20.dp),
        color = color.copy(alpha = 0.15f)
    ) {
        Text(
            text = label,
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
            color = color,
            fontSize = 13.sp,
            fontWeight = FontWeight.SemiBold
        )
    }
}

@Composable
fun SurgeBadge(multiplier: Double, basePrice: Double, surgePrice: Double) {
    Card(
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = Color(0xFFFFF8E1))
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(Icons.Default.TrendingUp, contentDescription = null, tint = Color(0xFFFF9800))
            Spacer(modifier = Modifier.width(8.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    "Surge Pricing Active",
                    fontWeight = FontWeight.SemiBold,
                    fontSize = 13.sp,
                    color = Color(0xFFE65100)
                )
                Text(
                    "${String.format("%.1f", multiplier)}× multiplier",
                    fontSize = 12.sp,
                    color = Color(0xFF795548)
                )
            }
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    "₹${String.format("%.0f", surgePrice)}",
                    fontWeight = FontWeight.Bold,
                    fontSize = 16.sp,
                    color = Color(0xFFE65100)
                )
                Text(
                    "₹${String.format("%.0f", basePrice)}",
                    fontSize = 12.sp,
                    color = Color(0xFF9E9E9E),
                )
            }
        }
    }
}

@Composable
internal fun ActiveJobDetail(icon: ImageVector, label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            icon,
            contentDescription = null,
            tint = MaterialTheme.colorScheme.primary,
            modifier = Modifier.size(20.dp)
        )
        Spacer(modifier = Modifier.width(12.dp))
        Column {
            Text(label, fontSize = 11.sp, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
            Text(value, fontSize = 14.sp, fontWeight = FontWeight.Medium)
        }
    }
}

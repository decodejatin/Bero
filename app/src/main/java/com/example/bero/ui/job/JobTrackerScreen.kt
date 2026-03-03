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
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.models.Job
import com.example.bero.data.models.JobStatus

/**
 * Job Tracker Screen — Client's real-time view of job progress.
 * Step indicator: ASSIGNED → IN_PROGRESS → WORKER_COMPLETED → FULLY_COMPLETED
 * Shows "Confirm Completion" after worker marks done.
 * Maps to backend: POST /jobs/:id/confirm-by-client
 */
@Composable
fun JobTrackerScreen(
    job: Job,
    onBackClick: () -> Unit = {},
    onConfirmCompletion: (String) -> Unit = {},
    onChatClick: (String) -> Unit = {},
    onRateClick: (String) -> Unit = {}
) {
    var showConfirmDialog by remember { mutableStateOf(false) }
    var isLoading by remember { mutableStateOf(false) }

    val steps = listOf("Assigned", "In Progress", "Worker Done", "Confirmed")
    val currentStep = when (job.status) {
        JobStatus.ASSIGNED, JobStatus.ACCEPTED -> 0
        JobStatus.IN_PROGRESS -> 1
        JobStatus.WORKER_COMPLETED -> 2
        JobStatus.FULLY_COMPLETED, JobStatus.COMPLETED -> 3
        else -> 0
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Job Status") },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                }
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
            // Step progress indicator
            TrackerStepIndicator(steps = steps, currentStep = currentStep)

            Spacer(modifier = Modifier.height(24.dp))

            // Job title
            Text(job.title, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Text(
                job.category.displayName,
                fontSize = 14.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Price card with surge
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.3f)
                )
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column {
                        Text("Total Price", fontSize = 12.sp, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
                        Text(
                            "₹${String.format("%.0f", job.displayPrice)}",
                            fontSize = 24.sp,
                            fontWeight = FontWeight.Bold,
                            color = MaterialTheme.colorScheme.primary
                        )
                    }
                    if (job.hasSurge) {
                        Surface(
                            shape = RoundedCornerShape(8.dp),
                            color = Color(0xFFFFF3E0)
                        ) {
                            Row(modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)) {
                                Icon(
                                    Icons.Default.TrendingUp,
                                    contentDescription = null,
                                    tint = Color(0xFFFF9800),
                                    modifier = Modifier.size(16.dp)
                                )
                                Spacer(modifier = Modifier.width(4.dp))
                                Text(
                                    "${String.format("%.1f", job.surgeMultiplier)}×",
                                    fontSize = 13.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = Color(0xFFE65100)
                                )
                            }
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Worker info card
            if (job.workerName != null) {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp)
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
                                .background(MaterialTheme.colorScheme.tertiary),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = (job.workerName ?: "W").first().uppercase(),
                                color = MaterialTheme.colorScheme.onTertiary,
                                fontWeight = FontWeight.Bold,
                                fontSize = 20.sp
                            )
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Column(modifier = Modifier.weight(1f)) {
                            Text(job.workerName ?: "Worker", fontWeight = FontWeight.SemiBold)
                            if (job.workerRating != null) {
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    Icon(
                                        Icons.Default.Star,
                                        contentDescription = null,
                                        tint = Color(0xFFFFC107),
                                        modifier = Modifier.size(16.dp)
                                    )
                                    Spacer(modifier = Modifier.width(4.dp))
                                    Text(
                                        "${job.workerRating}",
                                        fontSize = 13.sp,
                                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                                    )
                                }
                            }
                        }
                        IconButton(onClick = { onChatClick(job.workerId ?: "") }) {
                            Icon(Icons.Default.Chat, contentDescription = "Chat", tint = MaterialTheme.colorScheme.primary)
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Job details
            Card(modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(12.dp)) {
                Column(modifier = Modifier.padding(16.dp)) {
                    ActiveJobDetail(Icons.Default.LocationOn, "Address", job.address.ifEmpty { job.location })
                    ActiveJobDetail(Icons.Default.Schedule, "Scheduled", "${job.scheduledDate} • ${job.scheduledTimeSlot}")
                    ActiveJobDetail(Icons.Default.Timer, "Duration", "${job.estimatedDurationMinutes} min")
                }
            }

            Spacer(modifier = Modifier.weight(1f))
            Spacer(modifier = Modifier.height(24.dp))

            // Confirm completion button (visible after WORKER_COMPLETED)
            if (job.status == JobStatus.WORKER_COMPLETED) {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = CardDefaults.cardColors(containerColor = Color(0xFFE8F5E9))
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(Icons.Default.CheckCircle, contentDescription = null, tint = Color(0xFF4CAF50))
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("Worker marked this job as complete", fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                        }
                        Spacer(modifier = Modifier.height(4.dp))
                        Text("Please confirm if the work is done to your satisfaction.", fontSize = 13.sp, color = Color(0xFF616161))
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                Button(
                    onClick = { showConfirmDialog = true },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(56.dp),
                    shape = RoundedCornerShape(16.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF4CAF50)),
                    enabled = !isLoading
                ) {
                    if (isLoading) {
                        CircularProgressIndicator(modifier = Modifier.size(24.dp), color = Color.White, strokeWidth = 2.dp)
                    } else {
                        Icon(Icons.Default.ThumbUp, contentDescription = null)
                        Spacer(modifier = Modifier.width(8.dp))
                        Text("Confirm Completion", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    }
                }
            }

            // Rate button (visible after FULLY_COMPLETED)
            if (job.status == JobStatus.FULLY_COMPLETED && !job.clientRated) {
                Button(
                    onClick = { onRateClick(job.id) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(56.dp),
                    shape = RoundedCornerShape(16.dp)
                ) {
                    Icon(Icons.Default.Star, contentDescription = null)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Rate Worker", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                }
            }
        }
    }

    if (showConfirmDialog) {
        AlertDialog(
            onDismissRequest = { showConfirmDialog = false },
            title = { Text("Confirm Completion?") },
            text = { Text("Are you satisfied with the work? You'll be asked to rate the worker after confirmation.") },
            confirmButton = {
                TextButton(onClick = {
                    showConfirmDialog = false
                    isLoading = true
                    onConfirmCompletion(job.id)
                }) { Text("Yes, Confirm") }
            },
            dismissButton = {
                TextButton(onClick = { showConfirmDialog = false }) { Text("Not Yet") }
            }
        )
    }
}

@Composable
private fun TrackerStepIndicator(steps: List<String>, currentStep: Int) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        steps.forEachIndexed { index, label ->
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                modifier = Modifier.weight(1f)
            ) {
                val isCompleted = index <= currentStep
                val isCurrent = index == currentStep

                Box(
                    modifier = Modifier
                        .size(32.dp)
                        .clip(CircleShape)
                        .background(
                            when {
                                isCompleted -> Color(0xFF4CAF50)
                                else -> MaterialTheme.colorScheme.outline.copy(alpha = 0.3f)
                            }
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    if (isCompleted && !isCurrent) {
                        Icon(Icons.Default.Check, contentDescription = null, tint = Color.White, modifier = Modifier.size(18.dp))
                    } else {
                        Text(
                            "${index + 1}",
                            fontSize = 13.sp,
                            fontWeight = FontWeight.Bold,
                            color = if (isCompleted) Color.White else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f)
                        )
                    }
                }
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    label,
                    fontSize = 10.sp,
                    textAlign = TextAlign.Center,
                    fontWeight = if (isCurrent) FontWeight.Bold else FontWeight.Normal,
                    color = if (isCompleted) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                )
            }
        }
    }
}

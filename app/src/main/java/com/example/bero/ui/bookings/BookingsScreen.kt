@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.bookings

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.bero.data.models.*
import com.example.bero.ui.jobs.JobsViewModel
import java.text.SimpleDateFormat
import java.util.*

/**
 * Bookings screen for both clients and workers to track their jobs
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BookingsScreen(
    isWorker: Boolean = false,
    jobsViewModel: JobsViewModel = viewModel(),
    onBookingClick: (Job) -> Unit = {},
    onWorkerClick: (String) -> Unit = {}, // Navigate to worker profile (for clients)
    onClientClick: (String) -> Unit = {}, // Navigate to client profile (for workers)
    onRateClick: (Job) -> Unit = {} // Navigate to rating screen
) {
    var selectedTab by remember { mutableIntStateOf(0) }
    val tabs = listOf("Active", "Completed", "Cancelled")
    
    val myJobs by jobsViewModel.myJobs.collectAsState()
    val uiState by jobsViewModel.uiState.collectAsState()
    
    // State for cancel confirmation dialog
    var jobToCancel by remember { mutableStateOf<Job?>(null) }
    
    // Refresh jobs on screen load
    LaunchedEffect(Unit) {
        jobsViewModel.loadJobs()
    }
    
    val filteredBookings = when (selectedTab) {
        0 -> myJobs.filter { it.status in listOf(JobStatus.OPEN, JobStatus.ACCEPTED, JobStatus.IN_PROGRESS, JobStatus.WORKER_COMPLETED, JobStatus.AWAITING_CONFIRMATION, JobStatus.FULLY_COMPLETED) }
        1 -> myJobs.filter { it.status in listOf(JobStatus.COMPLETED, JobStatus.FULLY_COMPLETED) }
        2 -> myJobs.filter { it.status == JobStatus.CANCELLED }
        else -> myJobs
    }
    
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        // Header
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = MaterialTheme.colorScheme.surface,
            shadowElevation = 2.dp
        ) {
            Column {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "My Bookings",
                        style = MaterialTheme.typography.headlineSmall,
                        fontWeight = FontWeight.Bold
                    )
                }
                
                // Tab row
                TabRow(
                    selectedTabIndex = selectedTab,
                    containerColor = Color.Transparent
                ) {
                    tabs.forEachIndexed { index, title ->
                        Tab(
                            selected = selectedTab == index,
                            onClick = { selectedTab = index },
                            text = { Text(title) }
                        )
                    }
                }
            }
        }
        
        if (filteredBookings.isEmpty()) {
            EmptyBookingsState(selectedTab)
        } else {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(filteredBookings) { job ->
                    BookingCard(
                        job = job,
                        isWorker = isWorker,
                        onClick = { onBookingClick(job) },
                        onWorkerClick = { workerId -> onWorkerClick(workerId) },
                        onClientClick = { clientId -> onClientClick(clientId) },
                        onCancelClick = if (selectedTab == 0 && job.status in listOf(JobStatus.OPEN, JobStatus.ACCEPTED)) {
                            { jobToCancel = job }
                        } else null,
                        onConfirmClick = if (!isWorker && job.status in listOf(JobStatus.WORKER_COMPLETED, JobStatus.AWAITING_CONFIRMATION)) {
                            { jobsViewModel.confirmJobCompletion(job.id) }
                        } else null,
                        onStartJobClick = if (isWorker && job.status == JobStatus.ACCEPTED) {
                            { jobsViewModel.startJob(job.id) }
                        } else null,
                        onMarkCompleteClick = if (isWorker && job.status == JobStatus.IN_PROGRESS) {
                            { jobsViewModel.workerMarkComplete(job.id) }
                        } else null,
                        onRateClick = if (job.status == JobStatus.FULLY_COMPLETED) {
                            { onRateClick(job) }
                        } else null
                    )
                }
                
                item {
                    Spacer(modifier = Modifier.height(80.dp))
                }
            }
        }
    }
    
    // Cancel confirmation dialog
    jobToCancel?.let { job ->
        AlertDialog(
            onDismissRequest = { jobToCancel = null },
            title = { Text("Cancel Booking?") },
            text = { 
                Text("Are you sure you want to cancel this booking? This action cannot be undone.")
            },
            confirmButton = {
                Button(
                    onClick = {
                        jobsViewModel.cancelJob(job.id)
                        jobToCancel = null
                    },
                    colors = ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.error
                    )
                ) {
                    Text("Cancel Booking")
                }
            },
            dismissButton = {
                OutlinedButton(onClick = { jobToCancel = null }) {
                    Text("Keep Booking")
                }
            }
        )
    }
}

@Composable
private fun BookingCard(
    job: Job,
    isWorker: Boolean = false,
    onClick: () -> Unit,
    onWorkerClick: (String) -> Unit = {},
    onClientClick: (String) -> Unit = {},
    onCancelClick: (() -> Unit)? = null,
    onConfirmClick: (() -> Unit)? = null,
    onStartJobClick: (() -> Unit)? = null,
    onMarkCompleteClick: (() -> Unit)? = null,
    onRateClick: (() -> Unit)? = null
) {
    var showCompleteDialog by remember { mutableStateOf(false) }
    
    val statusColor = when (job.status) {
        JobStatus.OPEN -> Color(0xFF2196F3)
        JobStatus.ACCEPTED, JobStatus.ASSIGNED -> Color(0xFF9C27B0)
        JobStatus.IN_PROGRESS -> Color(0xFFFF9800)
        JobStatus.AWAITING_CONFIRMATION, JobStatus.WORKER_COMPLETED -> Color(0xFF00BCD4)
        JobStatus.CLIENT_CONFIRMED -> Color(0xFF00BCD4)
        JobStatus.FULLY_COMPLETED -> Color(0xFFFFC107) // Amber — rating required
        JobStatus.COMPLETED -> Color(0xFF4CAF50)
        JobStatus.CANCELLED -> Color(0xFF757575)
        else -> Color(0xFF757575)
    }
    
    val statusText = when (job.status) {
        JobStatus.OPEN -> "Finding Worker"
        JobStatus.ACCEPTED, JobStatus.ASSIGNED -> if (isWorker) "You Accepted" else "Worker Assigned"
        JobStatus.IN_PROGRESS -> "In Progress"
        JobStatus.WORKER_COMPLETED -> if (isWorker) "Waiting for Client" else "Worker Completed"
        JobStatus.AWAITING_CONFIRMATION -> if (isWorker) "Pending Confirmation" else "Awaiting Your Confirmation"
        JobStatus.CLIENT_CONFIRMED -> if (isWorker) "Client Confirmed" else "You Confirmed"
        JobStatus.FULLY_COMPLETED -> "⭐ Rating Required"
        JobStatus.COMPLETED -> "Completed"
        JobStatus.CANCELLED -> "Cancelled"
        else -> "Unknown"
    }
    
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            // Header with status
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Category badge
                Surface(
                    color = Color(job.category.color).copy(alpha = 0.15f),
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(job.category.emoji, fontSize = 14.sp)
                        Spacer(modifier = Modifier.width(4.dp))
                        Text(
                            text = job.category.displayName,
                            style = MaterialTheme.typography.labelMedium,
                            color = Color(job.category.color),
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                }
                
                // Status badge
                Surface(
                    color = statusColor.copy(alpha = 0.15f),
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text(
                        text = statusText,
                        style = MaterialTheme.typography.labelSmall,
                        color = statusColor,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp)
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(12.dp))
            
            // Job title and amount
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = job.title,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.weight(1f)
                )
                Text(
                    text = "₹${job.amountRupees.toInt()}",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.primary
                )
            }
            
            Spacer(modifier = Modifier.height(8.dp))
            
            // Location
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    Icons.Outlined.LocationOn,
                    contentDescription = null,
                    modifier = Modifier.size(16.dp),
                    tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                )
                Spacer(modifier = Modifier.width(4.dp))
                Text(
                    text = job.location,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                )
            }
            
            // Worker info section - show worker details when job is accepted
            if (job.status in listOf(JobStatus.ACCEPTED, JobStatus.IN_PROGRESS, JobStatus.WORKER_COMPLETED, JobStatus.AWAITING_CONFIRMATION, JobStatus.COMPLETED, JobStatus.FULLY_COMPLETED) && job.workerId != null) {
                Spacer(modifier = Modifier.height(12.dp))
                HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
                Spacer(modifier = Modifier.height(12.dp))
                
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                // Show different profile based on user type
                if (isWorker) {
                    // Worker view: Show CLIENT info
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        modifier = Modifier
                            .weight(1f)
                            .clickable { onClientClick(job.clientId) }
                    ) {
                        // Client avatar
                        Surface(
                            shape = CircleShape,
                            color = MaterialTheme.colorScheme.secondaryContainer,
                            modifier = Modifier.size(40.dp)
                        ) {
                            Box(contentAlignment = Alignment.Center) {
                                Text(
                                    text = job.clientName.firstOrNull()?.uppercase()?.toString() ?: "C",
                                    style = MaterialTheme.typography.titleMedium,
                                    fontWeight = FontWeight.Bold,
                                    color = MaterialTheme.colorScheme.onSecondaryContainer
                                )
                            }
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text(
                                text = job.clientName,
                                style = MaterialTheme.typography.bodyMedium,
                                fontWeight = FontWeight.SemiBold
                            )
                            Text(
                                text = "Client",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                            )
                        }
                    }
                } else {
                    // Client view: Show WORKER info (clickable to view profile)
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        modifier = Modifier
                            .weight(1f)
                            .clickable { job.workerId?.let { onWorkerClick(it) } }
                    ) {
                        // Worker avatar
                        Surface(
                            shape = CircleShape,
                            color = MaterialTheme.colorScheme.primaryContainer,
                            modifier = Modifier.size(40.dp)
                        ) {
                            Box(contentAlignment = Alignment.Center) {
                                Text(
                                    text = (job.workerName?.firstOrNull()?.uppercase() ?: "W"),
                                    style = MaterialTheme.typography.titleMedium,
                                    fontWeight = FontWeight.Bold,
                                    color = MaterialTheme.colorScheme.onPrimaryContainer
                                )
                            }
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text(
                                text = job.workerName ?: "Worker Assigned",
                                style = MaterialTheme.typography.bodyMedium,
                                fontWeight = FontWeight.SemiBold
                            )
                            if (job.workerRating != null) {
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    Icon(
                                        Icons.Filled.Star,
                                        contentDescription = null,
                                        modifier = Modifier.size(14.dp),
                                        tint = Color(0xFFFFC107)
                                    )
                                    Spacer(modifier = Modifier.width(4.dp))
                                    Text(
                                        text = String.format("%.1f", job.workerRating),
                                        style = MaterialTheme.typography.bodySmall,
                                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                                    )
                                }
                            }
                        }
                    }
                }
                    
                    // Action buttons based on status
                    when (job.status) {
                        JobStatus.ACCEPTED, JobStatus.IN_PROGRESS -> {
                            Row {
                                IconButton(onClick = { }) {
                                    Icon(Icons.Default.Chat, contentDescription = "Chat")
                                }
                                IconButton(onClick = { }) {
                                    Icon(Icons.Default.Call, contentDescription = "Call")
                                }
                            }
                        }
                        else -> {}
                    }
                }
            }
            
            // Date info
            Spacer(modifier = Modifier.height(12.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Default.CalendarToday,
                        contentDescription = null,
                        modifier = Modifier.size(14.dp),
                        tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "${job.scheduledDate}, ${job.scheduledTime}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                }
                
                Text(
                    text = SimpleDateFormat("dd MMM yyyy", Locale.getDefault())
                        .format(Date(job.postedAt)),
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f)
                )
            }
            
            // Cancel button for active bookings
            onCancelClick?.let { onCancel ->
                Spacer(modifier = Modifier.height(12.dp))
                HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
                Spacer(modifier = Modifier.height(12.dp))
                
                OutlinedButton(
                    onClick = onCancel,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(8.dp),
                    colors = ButtonDefaults.outlinedButtonColors(
                        contentColor = MaterialTheme.colorScheme.error
                    ),
                    border = androidx.compose.foundation.BorderStroke(
                        1.dp,
                        MaterialTheme.colorScheme.error.copy(alpha = 0.5f)
                    )
                ) {
                    Icon(
                        Icons.Default.Close,
                        contentDescription = null,
                        modifier = Modifier.size(18.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Cancel Booking")
                }
            }
            
            // Worker action: Start Job (ACCEPTED → IN_PROGRESS)
            onStartJobClick?.let { onStart ->
                Spacer(modifier = Modifier.height(12.dp))
                if (onCancelClick == null) {
                    HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
                    Spacer(modifier = Modifier.height(12.dp))
                }
                
                Button(
                    onClick = onStart,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(8.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = Color(0xFF2196F3)
                    )
                ) {
                    Icon(
                        Icons.Default.PlayArrow,
                        contentDescription = null,
                        modifier = Modifier.size(18.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Start Job", fontWeight = FontWeight.SemiBold)
                }
            }
            
            // Worker action: Mark Complete (IN_PROGRESS → WORKER_COMPLETED)
            onMarkCompleteClick?.let { onComplete ->
                Spacer(modifier = Modifier.height(12.dp))
                HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
                Spacer(modifier = Modifier.height(12.dp))
                
                Button(
                    onClick = { showCompleteDialog = true },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(8.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = Color(0xFF4CAF50)
                    )
                ) {
                    Icon(
                        Icons.Default.CheckCircle,
                        contentDescription = null,
                        modifier = Modifier.size(18.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Mark as Completed", fontWeight = FontWeight.SemiBold)
                }
            }
            
            // Worker status: Waiting for client confirmation
            if (isWorker && job.status == JobStatus.WORKER_COMPLETED) {
                Spacer(modifier = Modifier.height(12.dp))
                HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
                Spacer(modifier = Modifier.height(12.dp))
                
                Surface(
                    color = Color(0xFFFFF3E0),
                    shape = RoundedCornerShape(8.dp),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Row(
                        modifier = Modifier.padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(Icons.Default.HourglassTop, contentDescription = null, tint = Color(0xFFFF9800), modifier = Modifier.size(20.dp))
                        Spacer(modifier = Modifier.width(8.dp))
                        Text("Waiting for client to confirm completion", fontSize = 13.sp, color = Color(0xFF795548))
                    }
                }
            }
            
            // Confirm completion button for AWAITING_CONFIRMATION jobs
            onConfirmClick?.let { onConfirm ->
                Spacer(modifier = Modifier.height(12.dp))
                if (onCancelClick == null) {
                    HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
                    Spacer(modifier = Modifier.height(12.dp))
                }
                
                Button(
                    onClick = onConfirm,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(8.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = Color(0xFF4CAF50)
                    )
                ) {
                    Icon(
                        Icons.Default.CheckCircle,
                        contentDescription = null,
                        modifier = Modifier.size(18.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Confirm Job Completed", fontWeight = FontWeight.SemiBold)
                }
            }
            
            // Rate Now button for FULLY_COMPLETED jobs
            onRateClick?.let { onRate ->
                Spacer(modifier = Modifier.height(12.dp))
                HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
                Spacer(modifier = Modifier.height(12.dp))
                
                Button(
                    onClick = onRate,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(8.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = Color(0xFFFFC107)
                    )
                ) {
                    Icon(
                        Icons.Default.Star,
                        contentDescription = null,
                        modifier = Modifier.size(18.dp),
                        tint = Color.White
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Rate Now", fontWeight = FontWeight.SemiBold, color = Color.White)
                }
            }
        }
    }
    
    // Mark complete confirmation dialog
    if (showCompleteDialog) {
        AlertDialog(
            onDismissRequest = { showCompleteDialog = false },
            title = { Text("Mark Job Complete?") },
            text = { Text("Are you sure this job is finished? The client will be asked to confirm.") },
            confirmButton = {
                Button(
                    onClick = {
                        showCompleteDialog = false
                        onMarkCompleteClick?.invoke()
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF4CAF50))
                ) {
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
private fun EmptyBookingsState(selectedTab: Int) {
    val (emoji, title, subtitle) = when (selectedTab) {
        0 -> Triple("📋", "No Active Bookings", "Your ongoing bookings will appear here")
        1 -> Triple("✅", "No Completed Bookings", "Completed jobs will be shown here")
        2 -> Triple("❌", "No Cancelled Bookings", "Cancelled bookings will appear here")
        else -> Triple("📋", "No Bookings", "")
    }
    
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.padding(32.dp)
        ) {
            Text(
                text = emoji,
                fontSize = 64.sp
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = title,
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = subtitle,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f)
            )
        }
    }
}

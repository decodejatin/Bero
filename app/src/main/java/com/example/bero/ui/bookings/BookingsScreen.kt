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
 * Bookings screen for clients to track their job requests
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BookingsScreen(
    jobsViewModel: JobsViewModel = viewModel(),
    onBookingClick: (Job) -> Unit = {}
) {
    var selectedTab by remember { mutableIntStateOf(0) }
    val tabs = listOf("Active", "Completed", "Cancelled")
    
    val myJobs by jobsViewModel.myJobs.collectAsState()
    val uiState by jobsViewModel.uiState.collectAsState()
    
    // Refresh jobs on screen load
    LaunchedEffect(Unit) {
        jobsViewModel.loadJobs()
    }
    
    val filteredBookings = when (selectedTab) {
        0 -> myJobs.filter { it.status in listOf(JobStatus.OPEN, JobStatus.ACCEPTED, JobStatus.IN_PROGRESS) }
        1 -> myJobs.filter { it.status == JobStatus.COMPLETED }
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
                        onClick = { onBookingClick(job) }
                    )
                }
                
                item {
                    Spacer(modifier = Modifier.height(80.dp))
                }
            }
        }
    }
}

@Composable
private fun BookingCard(
    job: Job,
    onClick: () -> Unit
) {
    val statusColor = when (job.status) {
        JobStatus.OPEN -> Color(0xFF2196F3)
        JobStatus.ACCEPTED, JobStatus.ASSIGNED -> Color(0xFF9C27B0)
        JobStatus.IN_PROGRESS -> Color(0xFFFF9800)
        JobStatus.COMPLETED -> Color(0xFF4CAF50)
        JobStatus.CANCELLED -> Color(0xFF757575)
        else -> Color(0xFF757575)
    }
    
    val statusText = when (job.status) {
        JobStatus.OPEN -> "Finding Worker"
        JobStatus.ACCEPTED, JobStatus.ASSIGNED -> "Worker Assigned"
        JobStatus.IN_PROGRESS -> "In Progress"
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
            
            // Worker info section - removed since Job doesn't have worker property
            // Job assignment status can be shown in the status badge above
            if (job.status in listOf(JobStatus.ACCEPTED, JobStatus.IN_PROGRESS, JobStatus.COMPLETED)) {
                Spacer(modifier = Modifier.height(12.dp))
                HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
                Spacer(modifier = Modifier.height(12.dp))
                
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        text = "Status: ${job.status.name.replace("_", " ")}",
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.Medium
                    )
                    
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
        }
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

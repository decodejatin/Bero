package com.example.bero.ui.jobs

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.models.Job
import com.example.bero.data.models.JobStatus
import com.example.bero.data.models.ServiceCategory
import java.text.SimpleDateFormat
import java.util.*

/**
 * Jobs listing screen for workers
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun JobsListScreen(
    jobs: List<Job>,
    isLoading: Boolean = false,
    selectedTab: Int = 0,
    onTabChange: (Int) -> Unit = {},
    onJobClick: (Job) -> Unit = {},
    onRefresh: () -> Unit = {}
) {
    val tabs = listOf("Available", "My Jobs")
    
    Column(modifier = Modifier.fillMaxSize()) {
        // Tab Row
        TabRow(
            selectedTabIndex = selectedTab,
            containerColor = MaterialTheme.colorScheme.surface
        ) {
            tabs.forEachIndexed { index, title ->
                Tab(
                    selected = selectedTab == index,
                    onClick = { onTabChange(index) },
                    text = { 
                        Text(
                            text = title,
                            fontWeight = if (selectedTab == index) FontWeight.Bold else FontWeight.Normal
                        ) 
                    }
                )
            }
        }
        
        // Content
        if (isLoading) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                CircularProgressIndicator()
            }
        } else if (jobs.isEmpty()) {
            EmptyJobsState(isMyJobs = selectedTab == 1)
        } else {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(jobs) { job ->
                    JobCard(
                        job = job,
                        onClick = { onJobClick(job) }
                    )
                }
            }
        }
    }
}

@Composable
private fun JobCard(
    job: Job,
    onClick: () -> Unit
) {
    val categoryIcon = getCategoryIcon(job.category)
    val categoryColor = getCategoryColor(job.category)
    
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            // Header Row
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.Top
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.weight(1f)
                ) {
                    // Category Icon
                    Box(
                        modifier = Modifier
                            .size(48.dp)
                            .clip(RoundedCornerShape(12.dp))
                            .background(categoryColor.copy(alpha = 0.15f)),
                        contentAlignment = Alignment.Center
                    ) {
                        Icon(
                            imageVector = categoryIcon,
                            contentDescription = null,
                            tint = categoryColor,
                            modifier = Modifier.size(24.dp)
                        )
                    }
                    
                    Spacer(modifier = Modifier.width(12.dp))
                    
                    Column {
                        Text(
                            text = job.title,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis
                        )
                        Text(
                            text = job.category.name.replace("_", " "),
                            style = MaterialTheme.typography.bodySmall,
                            color = categoryColor
                        )
                    }
                }
                
                // Urgent Badge
                if (job.isUrgent) {
                    Surface(
                        color = Color(0xFFF44336).copy(alpha = 0.1f),
                        shape = RoundedCornerShape(4.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                imageVector = Icons.Default.Bolt,
                                contentDescription = null,
                                tint = Color(0xFFF44336),
                                modifier = Modifier.size(14.dp)
                            )
                            Spacer(modifier = Modifier.width(2.dp))
                            Text(
                                text = "URGENT",
                                style = MaterialTheme.typography.labelSmall,
                                color = Color(0xFFF44336),
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }
                }
            }
            
            Spacer(modifier = Modifier.height(12.dp))
            
            // Description
            Text(
                text = job.description,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
                maxLines = 2,
                overflow = TextOverflow.Ellipsis
            )
            
            Spacer(modifier = Modifier.height(12.dp))
            
            // Info Row
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                InfoChip(
                    icon = Icons.Default.LocationOn,
                    text = "${job.locality}, ${job.city}"
                )
                InfoChip(
                    icon = Icons.Default.Schedule,
                    text = job.scheduledTimeSlot
                )
            }
            
            Spacer(modifier = Modifier.height(12.dp))
            
            HorizontalDivider()
            
            Spacer(modifier = Modifier.height(12.dp))
            
            // Bottom Row - Payment and Date
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Payment
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = "₹${job.paymentAmountRupees.toInt()}",
                        style = MaterialTheme.typography.titleLarge,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = " • ${job.estimatedDurationMinutes} min",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                    )
                }
                
                // Date
                Text(
                    text = "${job.scheduledDate}, ${job.scheduledTime ?: job.scheduledTimeSlot}",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                )
            }
            
            // Status badge for assigned/in-progress jobs
            if (job.status != JobStatus.OPEN) {
                Spacer(modifier = Modifier.height(12.dp))
                JobStatusBadge(status = job.status)
            }
        }
    }
}

@Composable
private fun InfoChip(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    text: String
) {
    Row(
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            imageVector = icon,
            contentDescription = null,
            modifier = Modifier.size(14.dp),
            tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
        )
        Spacer(modifier = Modifier.width(4.dp))
        Text(
            text = text,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
            maxLines = 1,
            overflow = TextOverflow.Ellipsis
        )
    }
}

@Composable
private fun JobStatusBadge(status: JobStatus) {
    val (text, color) = when (status) {
        JobStatus.ACCEPTED -> "Accepted" to Color(0xFF2196F3)
        JobStatus.ASSIGNED -> "Assigned to you" to Color(0xFF2196F3)
        JobStatus.IN_PROGRESS -> "In Progress" to Color(0xFFFF9800)
        JobStatus.WORKER_COMPLETED -> "Waiting for Client" to Color(0xFFFF9800)
        JobStatus.CLIENT_CONFIRMED -> "Client Confirmed" to Color(0xFF009688)
        JobStatus.FULLY_COMPLETED -> "⭐ Rate Required" to Color(0xFF9C27B0)
        JobStatus.COMPLETED -> "Completed" to Color(0xFF4CAF50)
        JobStatus.CANCELLED -> "Cancelled" to Color(0xFF9E9E9E)
        JobStatus.DISPUTED -> "Disputed" to Color(0xFFF44336)
        else -> "" to Color.Transparent
    }
    
    if (text.isNotEmpty()) {
        Surface(
            color = color.copy(alpha = 0.1f),
            shape = RoundedCornerShape(4.dp),
            modifier = Modifier.fillMaxWidth()
        ) {
            Text(
                text = text,
                style = MaterialTheme.typography.labelMedium,
                color = color,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.padding(8.dp)
            )
        }
    }
}

@Composable
private fun EmptyJobsState(isMyJobs: Boolean) {
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.padding(32.dp)
        ) {
            Text(
                text = if (isMyJobs) "📋" else "🔍",
                fontSize = 64.sp
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = if (isMyJobs) "No Jobs Yet" else "No Jobs Available",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = if (isMyJobs) 
                    "Accept jobs from the Available tab to see them here"
                else 
                    "New jobs will appear here. Pull down to refresh.",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
                textAlign = androidx.compose.ui.text.style.TextAlign.Center
            )
        }
    }
}

private fun getCategoryIcon(category: ServiceCategory): androidx.compose.ui.graphics.vector.ImageVector {
    return when (category) {
        ServiceCategory.PLUMBING, ServiceCategory.PLUMBER -> Icons.Default.Plumbing
        ServiceCategory.ELECTRICAL, ServiceCategory.ELECTRICIAN -> Icons.Default.ElectricalServices
        ServiceCategory.CARPENTRY, ServiceCategory.CARPENTER -> Icons.Default.Carpenter
        ServiceCategory.PAINTING, ServiceCategory.PAINTER -> Icons.Default.FormatPaint
        ServiceCategory.CLEANING, ServiceCategory.CLEANER -> Icons.Default.CleaningServices
        ServiceCategory.AC_REPAIR -> Icons.Default.AcUnit
        ServiceCategory.APPLIANCE_REPAIR -> Icons.Default.Kitchen
        ServiceCategory.PEST_CONTROL -> Icons.Default.PestControl
        ServiceCategory.GARDENING, ServiceCategory.GARDENER -> Icons.Default.Grass
        else -> Icons.Default.Build
    }
}

private fun getCategoryColor(category: ServiceCategory): Color {
    return when (category) {
        ServiceCategory.PLUMBING, ServiceCategory.PLUMBER -> Color(0xFF2196F3)
        ServiceCategory.ELECTRICAL, ServiceCategory.ELECTRICIAN -> Color(0xFFFFC107)
        ServiceCategory.CARPENTRY, ServiceCategory.CARPENTER -> Color(0xFF795548)
        ServiceCategory.PAINTING, ServiceCategory.PAINTER -> Color(0xFF9C27B0)
        ServiceCategory.CLEANING, ServiceCategory.CLEANER -> Color(0xFF00BCD4)
        ServiceCategory.AC_REPAIR -> Color(0xFF03A9F4)
        ServiceCategory.APPLIANCE_REPAIR -> Color(0xFF607D8B)
        ServiceCategory.PEST_CONTROL -> Color(0xFF4CAF50)
        ServiceCategory.GARDENING, ServiceCategory.GARDENER -> Color(0xFF8BC34A)
        else -> Color(0xFF9E9E9E)
    }
}

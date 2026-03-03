package com.example.bero.ui.jobs

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
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
import androidx.compose.ui.unit.dp
import com.example.bero.data.models.Job
import com.example.bero.data.models.JobStatus
import com.example.bero.data.models.ServiceCategory
import java.text.SimpleDateFormat
import java.util.*

/**
 * Job Detail Screen showing full job info with action buttons
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun JobDetailScreen(
    job: Job,
    isAccepting: Boolean = false,
    onAccept: () -> Unit = {},
    onStart: () -> Unit = {},
    onComplete: () -> Unit = {},
    onCancel: () -> Unit = {},
    onBack: () -> Unit = {},
    onCallClient: () -> Unit = {},
    onNavigate: () -> Unit = {},
    onRate: () -> Unit = {}
) {
    val scrollState = rememberScrollState()
    
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Job Details") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, "Back")
                    }
                }
            )
        },
        bottomBar = {
            JobActionBar(
                status = job.status,
                isLoading = isAccepting,
                onAccept = onAccept,
                onStart = onStart,
                onComplete = onComplete,
                onCancel = onCancel,
                onRate = onRate
            )
        }
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .verticalScroll(scrollState)
        ) {
            // Header Card
            JobHeaderCard(job = job)
            
            // Details Section
            JobDetailsSection(job = job)
            
            // Client Section (visible after acceptance)
            if (job.status != JobStatus.OPEN) {
                ClientSection(
                    clientName = job.clientName,
                    onCall = onCallClient,
                    onNavigate = onNavigate
                )
            }
            
            // Skills Required
            if (job.requiredSkills.isNotEmpty()) {
                SkillsRequiredSection(skills = job.requiredSkills)
            }
            
            Spacer(modifier = Modifier.height(100.dp)) // Space for bottom bar
        }
    }
}

@Composable
private fun JobHeaderCard(job: Job) {
    val categoryColor = getCategoryColor(job.category)
    
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.3f)
        )
    ) {
        Column(modifier = Modifier.padding(20.dp)) {
            // Category badge
            Surface(
                color = categoryColor.copy(alpha = 0.15f),
                shape = RoundedCornerShape(8.dp)
            ) {
                Text(
                    text = job.category.name.replace("_", " "),
                    style = MaterialTheme.typography.labelMedium,
                    color = categoryColor,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp)
                )
            }
            
            Spacer(modifier = Modifier.height(12.dp))
            
            // Title
            Text(
                text = job.title,
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold
            )
            
            Spacer(modifier = Modifier.height(8.dp))
            
            // Payment
            Row(verticalAlignment = Alignment.Bottom) {
                Text(
                    text = "₹${job.paymentAmountRupees.toInt()}",
                    style = MaterialTheme.typography.headlineMedium,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.primary
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "for ${job.estimatedDurationMinutes} min work",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                )
            }
            
            // Urgent badge
            if (job.isUrgent) {
                Spacer(modifier = Modifier.height(12.dp))
                Surface(
                    color = Color(0xFFF44336).copy(alpha = 0.1f),
                    shape = RoundedCornerShape(4.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            imageVector = Icons.Default.Bolt,
                            contentDescription = null,
                            tint = Color(0xFFF44336),
                            modifier = Modifier.size(18.dp)
                        )
                        Spacer(modifier = Modifier.width(4.dp))
                        Text(
                            text = "URGENT - Customer is waiting",
                            style = MaterialTheme.typography.bodyMedium,
                            color = Color(0xFFF44336),
                            fontWeight = FontWeight.Bold
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun JobDetailsSection(job: Job) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "Job Details",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // Description
            Text(
                text = job.description,
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            HorizontalDivider()
            Spacer(modifier = Modifier.height(16.dp))
            
            // Location
            DetailRow(
                icon = Icons.Default.LocationOn,
                label = "Location",
                value = "${job.address}\n${job.locality}, ${job.city} - ${job.pincode}"
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // Schedule
            val dateFormat = SimpleDateFormat("EEEE, d MMMM yyyy", Locale.getDefault())
            DetailRow(
                icon = Icons.Default.CalendarToday,
                label = "Date",
                value = dateFormat.format(Date(job.scheduledDate))
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            DetailRow(
                icon = Icons.Default.Schedule,
                label = "Time Slot",
                value = job.scheduledTimeSlot
            )
        }
    }
}

@Composable
private fun DetailRow(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    value: String
) {
    Row {
        Icon(
            imageVector = icon,
            contentDescription = null,
            tint = MaterialTheme.colorScheme.primary,
            modifier = Modifier.size(24.dp)
        )
        Spacer(modifier = Modifier.width(12.dp))
        Column {
            Text(
                text = label,
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
            )
            Text(
                text = value,
                style = MaterialTheme.typography.bodyLarge
            )
        }
    }
}

@Composable
private fun ClientSection(
    clientName: String,
    onCall: () -> Unit,
    onNavigate: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "Customer",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            
            Spacer(modifier = Modifier.height(12.dp))
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(48.dp)
                            .clip(RoundedCornerShape(24.dp))
                            .background(MaterialTheme.colorScheme.primaryContainer),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = clientName.first().uppercase(),
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.Bold,
                            color = MaterialTheme.colorScheme.onPrimaryContainer
                        )
                    }
                    Spacer(modifier = Modifier.width(12.dp))
                    Text(
                        text = clientName,
                        style = MaterialTheme.typography.bodyLarge,
                        fontWeight = FontWeight.Medium
                    )
                }
                
                Row {
                    IconButton(onClick = onCall) {
                        Icon(
                            imageVector = Icons.Default.Phone,
                            contentDescription = "Call",
                            tint = MaterialTheme.colorScheme.primary
                        )
                    }
                    IconButton(onClick = onNavigate) {
                        Icon(
                            imageVector = Icons.Default.Navigation,
                            contentDescription = "Navigate",
                            tint = MaterialTheme.colorScheme.primary
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun SkillsRequiredSection(skills: List<String>) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "Skills Required",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            
            Spacer(modifier = Modifier.height(12.dp))
            
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                skills.forEach { skill ->
                    SuggestionChip(
                        onClick = {},
                        label = { Text(skill) }
                    )
                }
            }
        }
    }
}

@Composable
private fun JobActionBar(
    status: JobStatus,
    isLoading: Boolean,
    onAccept: () -> Unit,
    onStart: () -> Unit,
    onComplete: () -> Unit,
    onCancel: () -> Unit,
    onRate: () -> Unit = {}
) {
    Surface(
        shadowElevation = 8.dp,
        color = MaterialTheme.colorScheme.surface
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            when (status) {
                JobStatus.OPEN -> {
                    Button(
                        onClick = onAccept,
                        modifier = Modifier.fillMaxWidth(),
                        enabled = !isLoading
                    ) {
                        if (isLoading) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(20.dp),
                                color = MaterialTheme.colorScheme.onPrimary
                            )
                        } else {
                            Icon(Icons.Default.Check, null)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("Accept Job", fontWeight = FontWeight.Bold)
                        }
                    }
                }
                JobStatus.ASSIGNED, JobStatus.ACCEPTED -> {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(12.dp)
                    ) {
                        OutlinedButton(
                            onClick = onCancel,
                            modifier = Modifier.weight(1f),
                            colors = ButtonDefaults.outlinedButtonColors(
                                contentColor = MaterialTheme.colorScheme.error
                            )
                        ) {
                            Text("Cancel")
                        }
                        Button(
                            onClick = onStart,
                            modifier = Modifier.weight(2f)
                        ) {
                            Icon(Icons.Default.PlayArrow, null)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("Start Job")
                        }
                    }
                }
                JobStatus.IN_PROGRESS -> {
                    Button(
                        onClick = onComplete,
                        modifier = Modifier.fillMaxWidth(),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = Color(0xFF4CAF50)
                        )
                    ) {
                        Icon(Icons.Default.CheckCircle, null)
                        Spacer(modifier = Modifier.width(8.dp))
                        Text("Mark Complete", fontWeight = FontWeight.Bold)
                    }
                }
                JobStatus.WORKER_COMPLETED -> {
                    // Worker waiting for client confirmation
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(12.dp),
                        colors = CardDefaults.cardColors(
                            containerColor = Color(0xFFFFF3E0)
                        )
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                Icons.Default.HourglassTop,
                                contentDescription = null,
                                tint = Color(0xFFFF9800),
                                modifier = Modifier.size(24.dp)
                            )
                            Spacer(modifier = Modifier.width(12.dp))
                            Column {
                                Text(
                                    "Completion Submitted",
                                    fontWeight = FontWeight.Bold,
                                    color = Color(0xFFE65100)
                                )
                                Text(
                                    "Waiting for client to confirm...",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = Color(0xFF795548)
                                )
                            }
                        }
                    }
                }
                JobStatus.FULLY_COMPLETED -> {
                    // Both confirmed — rate required
                    Button(
                        onClick = onRate,
                        modifier = Modifier.fillMaxWidth(),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = MaterialTheme.colorScheme.primary
                        )
                    ) {
                        Icon(Icons.Default.Star, null)
                        Spacer(modifier = Modifier.width(8.dp))
                        Text("Rate Client", fontWeight = FontWeight.Bold)
                    }
                }
                JobStatus.COMPLETED -> {
                    // Final state — all done
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(12.dp),
                        colors = CardDefaults.cardColors(
                            containerColor = Color(0xFFE8F5E9)
                        )
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                Icons.Default.CheckCircle,
                                contentDescription = null,
                                tint = Color(0xFF4CAF50),
                                modifier = Modifier.size(24.dp)
                            )
                            Spacer(modifier = Modifier.width(12.dp))
                            Text(
                                "Job Completed ✓",
                                fontWeight = FontWeight.Bold,
                                color = Color(0xFF2E7D32)
                            )
                        }
                    }
                }
                else -> {
                    // Cancelled / Disputed — no actions
                }
            }
        }
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
        ServiceCategory.OTHER -> Color(0xFF9E9E9E)
        else -> Color(0xFF9E9E9E) // Default for any other category
    }
}

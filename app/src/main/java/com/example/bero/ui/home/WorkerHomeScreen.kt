@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.home

import androidx.compose.animation.animateColorAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.DummyDataProvider
import com.example.bero.data.models.*

/**
 * Worker Home Screen - Shows available jobs, earnings summary, and quick actions
 */
@Composable
fun WorkerHomeScreen(
    onJobClick: (Job) -> Unit = {},
    onViewAllJobsClick: () -> Unit = {},
    onGoOnlineToggle: (Boolean) -> Unit = {}
) {
    var isOnline by remember { mutableStateOf(true) }
    val availableJobs = remember { DummyDataProvider.getJobs().filter { it.status == JobStatus.OPEN }.take(5) }
    val todayEarnings = remember { 1250.0 }
    val weeklyEarnings = remember { 8540.0 }
    val streakCount = remember { 7 }
    
    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background),
        contentPadding = PaddingValues(bottom = 100.dp)
    ) {
        // Header with greeting and online toggle
        item {
            WorkerHeader(
                isOnline = isOnline,
                onToggleOnline = { 
                    isOnline = it
                    onGoOnlineToggle(it)
                },
                streakCount = streakCount
            )
        }
        
        // Quick Stats
        item {
            QuickStatsSection(
                todayEarnings = todayEarnings,
                weeklyEarnings = weeklyEarnings,
                jobsCompleted = 3,
                rating = 4.8
            )
        }
        
        // Available Jobs Section
        item {
            SectionHeader(
                title = "Available Jobs Nearby",
                actionText = "View All",
                onActionClick = onViewAllJobsClick
            )
        }
        
        if (availableJobs.isEmpty()) {
            item {
                EmptyJobsCard()
            }
        } else {
            items(availableJobs) { job ->
                JobCard(
                    job = job,
                    onClick = { onJobClick(job) }
                )
            }
        }
        
        // Tips Section
        item {
            TipsSection()
        }
    }
}

@Composable
private fun WorkerHeader(
    isOnline: Boolean,
    onToggleOnline: (Boolean) -> Unit,
    streakCount: Int
) {
    val statusColor by animateColorAsState(
        targetValue = if (isOnline) Color(0xFF4CAF50) else Color(0xFF9E9E9E),
        label = "statusColor"
    )
    
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(
                brush = Brush.verticalGradient(
                    colors = listOf(
                        MaterialTheme.colorScheme.primary,
                        MaterialTheme.colorScheme.primaryContainer
                    )
                )
            )
            .padding(24.dp)
    ) {
        Column {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column {
                    Text(
                        text = "Good Morning! 👋",
                        style = MaterialTheme.typography.headlineSmall,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.onPrimary
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Box(
                            modifier = Modifier
                                .size(10.dp)
                                .clip(CircleShape)
                                .background(statusColor)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = if (isOnline) "You're Online" else "You're Offline",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onPrimary.copy(alpha = 0.8f)
                        )
                    }
                }
                
                // Online/Offline Toggle
                Switch(
                    checked = isOnline,
                    onCheckedChange = onToggleOnline,
                    colors = SwitchDefaults.colors(
                        checkedThumbColor = Color.White,
                        checkedTrackColor = Color(0xFF4CAF50),
                        uncheckedThumbColor = Color.White,
                        uncheckedTrackColor = Color(0xFF9E9E9E)
                    )
                )
            }
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // Streak Badge
            Surface(
                shape = RoundedCornerShape(20.dp),
                color = Color(0xFFFFD700).copy(alpha = 0.2f)
            ) {
                Row(
                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("🔥", fontSize = 16.sp)
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "$streakCount Day Streak!",
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFFFFD700)
                    )
                }
            }
        }
    }
}

@Composable
private fun QuickStatsSection(
    todayEarnings: Double,
    weeklyEarnings: Double,
    jobsCompleted: Int,
    rating: Double
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        StatCard(
            modifier = Modifier.weight(1f),
            title = "Today",
            value = "₹${todayEarnings.toInt()}",
            icon = Icons.Default.CurrencyRupee,
            color = Color(0xFF4CAF50)
        )
        StatCard(
            modifier = Modifier.weight(1f),
            title = "This Week",
            value = "₹${weeklyEarnings.toInt()}",
            icon = Icons.Default.TrendingUp,
            color = Color(0xFF2196F3)
        )
        StatCard(
            modifier = Modifier.weight(1f),
            title = "Jobs",
            value = "$jobsCompleted",
            icon = Icons.Default.CheckCircle,
            color = Color(0xFFFF9800)
        )
    }
}

@Composable
private fun StatCard(
    modifier: Modifier = Modifier,
    title: String,
    value: String,
    icon: ImageVector,
    color: Color
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = color.copy(alpha = 0.1f)
        )
    ) {
        Column(
            modifier = Modifier.padding(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                icon,
                contentDescription = null,
                tint = color,
                modifier = Modifier.size(24.dp)
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = value,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
                color = color
            )
            Text(
                text = title,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
            )
        }
    }
}

@Composable
private fun SectionHeader(
    title: String,
    actionText: String? = null,
    onActionClick: () -> Unit = {}
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold
        )
        if (actionText != null) {
            TextButton(onClick = onActionClick) {
                Text(actionText)
            }
        }
    }
}

@Composable
private fun JobCard(
    job: Job,
    onClick: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 6.dp)
            .clickable(onClick = onClick),
        shape = RoundedCornerShape(16.dp),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.Top
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(
                            text = job.category.emoji,
                            fontSize = 20.sp
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = job.title,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = job.location,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                }
                
                Surface(
                    shape = RoundedCornerShape(8.dp),
                    color = MaterialTheme.colorScheme.primaryContainer
                ) {
                    Text(
                        text = "₹${job.amountRupees.toInt()}",
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(12.dp))
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Default.Person,
                        contentDescription = null,
                        modifier = Modifier.size(16.dp),
                        tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = job.clientName,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "⭐ ${job.clientRating}",
                        style = MaterialTheme.typography.bodySmall
                    )
                }
                
                if (job.urgency == JobUrgency.URGENT) {
                    Surface(
                        shape = RoundedCornerShape(4.dp),
                        color = Color(0xFFFF5722).copy(alpha = 0.1f)
                    ) {
                        Text(
                            text = "URGENT",
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp),
                            style = MaterialTheme.typography.labelSmall,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFFFF5722)
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun EmptyJobsCard() {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                Icons.Default.WorkOff,
                contentDescription = null,
                modifier = Modifier.size(48.dp),
                tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f)
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = "No jobs available right now",
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Text(
                text = "Check back soon!",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.6f)
            )
        }
    }
}

@Composable
private fun TipsSection() {
    Column(modifier = Modifier.padding(16.dp)) {
        Text(
            text = "💡 Tips to earn more",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold
        )
        Spacer(modifier = Modifier.height(12.dp))
        
        LazyRow(
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            item {
                TipCard(
                    emoji = "⏰",
                    title = "Be responsive",
                    description = "Reply to job requests within 5 minutes"
                )
            }
            item {
                TipCard(
                    emoji = "⭐",
                    title = "Get 5-star ratings",
                    description = "Complete jobs professionally for better reviews"
                )
            }
            item {
                TipCard(
                    emoji = "🔥",
                    title = "Maintain streak",
                    description = "Work daily to unlock lower commission rates"
                )
            }
        }
    }
}

@Composable
private fun TipCard(
    emoji: String,
    title: String,
    description: String
) {
    Card(
        modifier = Modifier.width(200.dp),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.secondaryContainer.copy(alpha = 0.5f)
        )
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(emoji, fontSize = 24.sp)
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = title,
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold
            )
            Text(
                text = description,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSecondaryContainer.copy(alpha = 0.7f)
            )
        }
    }
}

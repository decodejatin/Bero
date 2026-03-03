package com.example.bero.ui.profile

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

/**
 * Reputation Badge Card — Reusable composable showing worker trust metrics.
 * Used on: WorkerProfileScreen, WorkerDetailsScreen, ActiveJobScreen, JobTrackerScreen.
 * Data from: GET /profile/:id, rating_avg + rating_count + tier on WorkerProfile.
 */
@Composable
fun ReputationBadgeCard(
    ratingAvg: Double,
    totalJobs: Int,
    completionRate: Double = 0.0, // 0-100
    tier: String = "STANDARD",
    cancellationRate: Double = 0.0, // 0-100
    modifier: Modifier = Modifier
) {
    val tierColor = when (tier.uppercase()) {
        "GOLD" -> Color(0xFFFFC107)
        "SILVER" -> Color(0xFF9E9E9E)
        "PLATINUM" -> Color(0xFF7C4DFF)
        else -> MaterialTheme.colorScheme.primary
    }

    Card(
        modifier = modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            // Header with tier badge
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    "Reputation",
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold
                )
                Surface(
                    shape = RoundedCornerShape(20.dp),
                    color = tierColor.copy(alpha = 0.15f)
                ) {
                    Row(modifier = Modifier.padding(horizontal = 10.dp, vertical = 4.dp)) {
                        Icon(
                            Icons.Default.Shield,
                            contentDescription = null,
                            tint = tierColor,
                            modifier = Modifier.size(16.dp)
                        )
                        Spacer(modifier = Modifier.width(4.dp))
                        Text(
                            tier.lowercase().replaceFirstChar { it.uppercase() },
                            fontSize = 12.sp,
                            fontWeight = FontWeight.Bold,
                            color = tierColor
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Stats row
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                ReputationStat(
                    icon = Icons.Default.Star,
                    value = String.format("%.1f", ratingAvg),
                    label = "Rating",
                    color = Color(0xFFFFC107)
                )
                ReputationStat(
                    icon = Icons.Default.WorkHistory,
                    value = "$totalJobs",
                    label = "Jobs",
                    color = MaterialTheme.colorScheme.primary
                )
                if (completionRate > 0) {
                    ReputationStat(
                        icon = Icons.Default.CheckCircle,
                        value = "${completionRate.toInt()}%",
                        label = "Completion",
                        color = Color(0xFF4CAF50)
                    )
                }
                if (cancellationRate > 0) {
                    ReputationStat(
                        icon = Icons.Default.Cancel,
                        value = "${cancellationRate.toInt()}%",
                        label = "Cancel Rate",
                        color = Color(0xFFF44336)
                    )
                }
            }
        }
    }
}

@Composable
private fun ReputationStat(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    value: String,
    label: String,
    color: Color
) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Icon(icon, contentDescription = null, tint = color, modifier = Modifier.size(20.dp))
        Spacer(modifier = Modifier.height(4.dp))
        Text(value, fontSize = 18.sp, fontWeight = FontWeight.Bold)
        Text(
            label,
            fontSize = 11.sp,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
        )
    }
}

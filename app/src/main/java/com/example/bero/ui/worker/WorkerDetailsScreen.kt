@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.worker

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.Chat
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
import com.example.bero.data.models.ServiceCategory
import com.example.bero.data.models.WorkerDisplayProfile
import com.example.bero.data.models.WorkerTier
import com.example.bero.data.models.Review
import com.example.bero.data.network.BeroApiClient
import com.example.bero.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WorkerDetailsScreen(
    workerId: String? = null,
    apiClient: BeroApiClient? = null,
    onBackClick: () -> Unit = {},
    onBookClick: (WorkerDisplayProfile) -> Unit = {},
    onChatClick: (String) -> Unit = {}
) {
    // Handle system back press
    androidx.activity.compose.BackHandler {
        onBackClick()
    }

    // Loading and error state
    var isLoading by remember { mutableStateOf(true) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    
    // Worker profile state - starts with placeholder, updated from API
    var worker by remember {
        mutableStateOf(
            WorkerDisplayProfile(
                userId = workerId ?: "",
                name = "Loading...",
                phoneNumber = "",
                rating = 0.0,
                totalJobs = 0,
                skills = emptyList(),
                isOnline = false,
                tier = WorkerTier.BRONZE,
                isKycVerified = false,
                hasVideoBio = false,
                streakCount = 0,
                distance = 0.0,
                location = "",
                memberSince = ""
            )
        )
    }
    
    // Fetch worker profile from API
    LaunchedEffect(workerId) {
        if (workerId != null && apiClient != null) {
            isLoading = true
            val result = apiClient.getWorkerProfile(workerId)
            result.fold(
                onSuccess = { profileDto ->
                    worker = WorkerDisplayProfile(
                        userId = profileDto.id,
                        name = profileDto.full_name ?: "Worker",
                        phoneNumber = profileDto.phone_number ?: "",
                        rating = 4.5, // TODO: Get from reviews API
                        totalJobs = 0, // TODO: Get from stats API
                        skills = listOf(ServiceCategory.PLUMBING), // TODO: Get from profile
                        isOnline = true,
                        tier = WorkerTier.BRONZE,
                        isKycVerified = profileDto.kyc_status == "VERIFIED",
                        hasVideoBio = false,
                        streakCount = 0,
                        distance = null,
                        location = profileDto.address ?: "Unknown",
                        memberSince = "2024"
                    )
                    isLoading = false
                },
                onFailure = { error ->
                    errorMessage = error.message
                    isLoading = false
                }
            )
        } else {
            isLoading = false
        }
    }

    val reviews = remember { emptyList<Review>() }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Worker Profile", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { /* Share */ }) {
                        Icon(Icons.Default.Share, contentDescription = "Share")
                    }
                    IconButton(onClick = { /* Favorite */ }) {
                        Icon(Icons.Default.FavoriteBorder, contentDescription = "Favorite")
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
        ) {
            // Profile Header
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                shape = RoundedCornerShape(20.dp),
                colors = CardDefaults.cardColors(containerColor = Color.Transparent)
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(
                            brush = Brush.verticalGradient(
                                colors = listOf(
                                    MaterialTheme.colorScheme.primaryContainer,
                                    MaterialTheme.colorScheme.surface
                                )
                            ),
                            shape = RoundedCornerShape(20.dp)
                        )
                        .padding(24.dp)
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        // Profile Image
                        Box {
                            Box(
                                modifier = Modifier
                                    .size(100.dp)
                                    .clip(CircleShape)
                                    .background(MaterialTheme.colorScheme.primary),
                                contentAlignment = Alignment.Center
                            ) {
                                Text(
                                    text = worker.name.first().uppercase(),
                                    fontSize = 40.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = Color.White
                                )
                            }

                            // Online indicator
                            if (worker.isOnline) {
                                Box(
                                    modifier = Modifier
                                        .size(24.dp)
                                        .align(Alignment.BottomEnd)
                                        .clip(CircleShape)
                                        .background(Color.White)
                                        .padding(3.dp)
                                        .clip(CircleShape)
                                        .background(LuxuryOnline)
                                )
                            }
                        }

                        Spacer(modifier = Modifier.height(16.dp))

                        // Name and badges
                        Row(
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(
                                text = worker.name,
                                fontSize = 24.sp,
                                fontWeight = FontWeight.Bold
                            )

                            if (worker.isKycVerified) {
                                Spacer(modifier = Modifier.width(8.dp))
                                Icon(
                                    Icons.Default.Verified,
                                    contentDescription = "Verified",
                                    tint = LuxuryVerified,
                                    modifier = Modifier.size(24.dp)
                                )
                            }
                        }

                        // Tier Badge
                        TierBadge(tier = worker.tier)

                        Spacer(modifier = Modifier.height(16.dp))

                        // Stats Row
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceEvenly
                        ) {
                            StatItem(
                                value = "${worker.rating}",
                                label = "Rating",
                                icon = Icons.Default.Star,
                                iconColor = LuxuryGold
                            )

                            StatItem(
                                value = "${worker.totalJobs}",
                                label = "Jobs Done",
                                icon = Icons.Default.CheckCircle,
                                iconColor = BeroSuccess
                            )

                            StatItem(
                                value = worker.distance?.let { "${it}km" } ?: "N/A",
                                label = "Distance",
                                icon = Icons.Default.LocationOn,
                                iconColor = BeroError
                            )
                        }
                    }
                }
            }

            // Quick Actions
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                OutlinedButton(
                    onClick = { onChatClick(worker.userId) },
                    modifier = Modifier.weight(1f),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Icon(Icons.AutoMirrored.Filled.Chat, contentDescription = null)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Message")
                }

                OutlinedButton(
                    onClick = { /* Call */ },
                    modifier = Modifier.weight(1f),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Icon(Icons.Default.Phone, contentDescription = null)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Call")
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Services/Skills
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp),
                shape = RoundedCornerShape(16.dp)
            ) {
                Column(
                    modifier = Modifier.padding(20.dp)
                ) {
                    Text(
                        text = "Services Offered",
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold
                    )

                    Spacer(modifier = Modifier.height(16.dp))

                    worker.skills.forEach { skill ->
                        ServiceRow(skill = skill)
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // About
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp),
                shape = RoundedCornerShape(16.dp)
            ) {
                Column(
                    modifier = Modifier.padding(20.dp)
                ) {
                    Text(
                        text = "About",
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold
                    )

                    Spacer(modifier = Modifier.height(12.dp))

                    InfoRow(Icons.Default.LocationOn, "Location", worker.location)
                    InfoRow(Icons.Default.CalendarMonth, "Member Since", worker.memberSince)
                    InfoRow(Icons.Default.LocalFireDepartment, "Streak", "${worker.streakCount} days")

                    if (worker.hasVideoBio) {
                        Spacer(modifier = Modifier.height(12.dp))

                        OutlinedButton(
                            onClick = { /* Play video */ },
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp)
                        ) {
                            Icon(Icons.Default.PlayCircle, contentDescription = null)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("Watch Video Bio")
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Reviews Section
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp),
                shape = RoundedCornerShape(16.dp)
            ) {
                Column(
                    modifier = Modifier.padding(20.dp)
                ) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = "Reviews",
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Bold
                        )

                        TextButton(onClick = { /* See all */ }) {
                            Text("See All")
                        }
                    }

                    reviews.forEach { review ->
                        ReviewItem(
                            reviewerName = review.reviewerName,
                            rating = review.rating,
                            comment = review.comment ?: "",
                            timeAgo = "2d ago"
                        )

                        if (review != reviews.last()) {
                            HorizontalDivider(modifier = Modifier.padding(vertical = 12.dp))
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(100.dp))
        }
    }

    // Bottom Book Button
    Box(
        modifier = Modifier.fillMaxSize()
    ) {
        Surface(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth(),
            shadowElevation = 16.dp,
            color = MaterialTheme.colorScheme.surface
        ) {
            Button(
                onClick = { onBookClick(worker) },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp)
                    .height(56.dp),
                shape = RoundedCornerShape(16.dp)
            ) {
                Icon(Icons.Default.CalendarToday, contentDescription = null)
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "Book Now",
                    fontSize = 18.sp,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}

@Composable
private fun TierBadge(tier: WorkerTier) {
    val (color, label) = when (tier) {
        WorkerTier.BRONZE -> TierBronze to "Bronze"
        WorkerTier.SILVER -> TierSilver to "Silver"
        WorkerTier.GOLD -> TierGold to "Gold"
        WorkerTier.PLATINUM -> TierPlatinum to "Platinum"
    }

    Surface(
        shape = RoundedCornerShape(16.dp),
        color = color.copy(alpha = 0.2f),
        modifier = Modifier.padding(top = 8.dp)
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                Icons.Default.EmojiEvents,
                contentDescription = null,
                tint = color,
                modifier = Modifier.size(16.dp)
            )
            Spacer(modifier = Modifier.width(4.dp))
            Text(
                text = "$label Worker",
                color = color,
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold
            )
        }
    }
}

@Composable
private fun StatItem(
    value: String,
    label: String,
    icon: ImageVector,
    iconColor: Color
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                icon,
                contentDescription = null,
                tint = iconColor,
                modifier = Modifier.size(20.dp)
            )
            Spacer(modifier = Modifier.width(4.dp))
            Text(
                text = value,
                fontSize = 20.sp,
                fontWeight = FontWeight.Bold
            )
        }
        Text(
            text = label,
            fontSize = 12.sp,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
        )
    }
}

@Composable
private fun ServiceRow(skill: ServiceCategory) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(44.dp)
                .clip(RoundedCornerShape(10.dp))
                .background(Color(skill.color.toInt()).copy(alpha = 0.2f)),
            contentAlignment = Alignment.Center
        ) {
            Text(text = skill.emoji, fontSize = 22.sp)
        }

        Spacer(modifier = Modifier.width(16.dp))

        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = skill.displayName,
                fontWeight = FontWeight.Medium,
                fontSize = 16.sp
            )
            Text(
                text = "Starting from ₹200/hour",
                fontSize = 12.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
            )
        }
    }
}

@Composable
private fun InfoRow(
    icon: ImageVector,
    label: String,
    value: String
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            icon,
            contentDescription = null,
            tint = MaterialTheme.colorScheme.primary,
            modifier = Modifier.size(24.dp)
        )

        Spacer(modifier = Modifier.width(16.dp))

        Column {
            Text(
                text = label,
                fontSize = 12.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
            )
            Text(
                text = value,
                fontSize = 14.sp
            )
        }
    }
}

@Composable
private fun ReviewItem(
    reviewerName: String,
    rating: Float,
    comment: String,
    timeAgo: String
) {
    Column {
        Row(
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .clip(CircleShape)
                    .background(MaterialTheme.colorScheme.primaryContainer),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = reviewerName.first().uppercase(),
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.primary
                )
            }

            Spacer(modifier = Modifier.width(12.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = reviewerName,
                    fontWeight = FontWeight.Medium
                )

                Row {
                    repeat(5) { index ->
                        Icon(
                            if (index < rating.toInt()) Icons.Default.Star else Icons.Default.StarBorder,
                            contentDescription = null,
                            tint = LuxuryGold,
                            modifier = Modifier.size(14.dp)
                        )
                    }
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = timeAgo,
                        fontSize = 12.sp,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                    )
                }
            }
        }

        if (comment.isNotBlank()) {
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = comment,
                fontSize = 14.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
            )
        }
    }
}

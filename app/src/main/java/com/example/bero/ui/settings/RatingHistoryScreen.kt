@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.settings

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.TokenManager
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
data class RatingHistoryItemDto(
    val job_id: String = "",
    val job_title: String = "",
    val other_party: String = "",
    val rating: Int = 0,
    val review: String = "",
    val is_given: Boolean = true,
    val created_at: String = ""
)

/**
 * Screen showing rating history — ratings given and received.
 */
@Composable
fun RatingHistoryScreen(
    onBackClick: () -> Unit = {}
) {
    val context = LocalContext.current
    val apiClient = remember { BeroApiClient(TokenManager(context)) }
    val scope = rememberCoroutineScope()
    val json = remember { Json { ignoreUnknownKeys = true } }

    var ratings by remember { mutableStateOf<List<RatingHistoryItemDto>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var error by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(Unit) {
        try {
            val response = apiClient.getMyRatings()
            ratings = json.decodeFromString<List<RatingHistoryItemDto>>(response)
        } catch (e: Exception) {
            error = e.message
        }
        isLoading = false
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Rating History", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        when {
            isLoading -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator()
                }
            }
            error != null -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    contentAlignment = Alignment.Center
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Icon(
                            Icons.Default.ErrorOutline,
                            contentDescription = null,
                            modifier = Modifier.size(48.dp),
                            tint = MaterialTheme.colorScheme.error
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text("Failed to load ratings", fontWeight = FontWeight.SemiBold)
                        Text(
                            error ?: "",
                            fontSize = 12.sp,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
            }
            ratings.isEmpty() -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    contentAlignment = Alignment.Center
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Icon(
                            Icons.Default.Star,
                            contentDescription = null,
                            modifier = Modifier.size(56.dp),
                            tint = Color(0xFFFFC107)
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        Text("No ratings yet", fontWeight = FontWeight.SemiBold, fontSize = 18.sp)
                        Text(
                            "Complete jobs and rate each other to see history here",
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            fontSize = 13.sp
                        )
                    }
                }
            }
            else -> {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .padding(horizontal = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                    contentPadding = PaddingValues(vertical = 12.dp)
                ) {
                    items(ratings) { item ->
                        RatingHistoryCard(item)
                    }
                }
            }
        }
    }
}

@Composable
private fun RatingHistoryCard(item: RatingHistoryItemDto) {
    val direction = if (item.is_given) "You rated" else "Rated by"
    val directionColor = if (item.is_given) Color(0xFF2196F3) else Color(0xFF4CAF50)
    val directionIcon = if (item.is_given) Icons.Default.ArrowUpward else Icons.Default.ArrowDownward

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        )
    ) {
        Column(modifier = Modifier.padding(14.dp)) {
            // Header: direction + job title
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Surface(
                        color = directionColor.copy(alpha = 0.12f),
                        shape = RoundedCornerShape(6.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                directionIcon,
                                contentDescription = null,
                                modifier = Modifier.size(14.dp),
                                tint = directionColor
                            )
                            Spacer(modifier = Modifier.width(4.dp))
                            Text(
                                direction,
                                fontSize = 11.sp,
                                fontWeight = FontWeight.SemiBold,
                                color = directionColor
                            )
                        }
                    }
                }

                // Stars
                Row {
                    repeat(5) { idx ->
                        Icon(
                            if (idx < item.rating) Icons.Default.Star else Icons.Default.StarBorder,
                            contentDescription = null,
                            modifier = Modifier.size(18.dp),
                            tint = Color(0xFFFFC107)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            // Job title
            if (item.job_title.isNotBlank()) {
                Text(
                    item.job_title,
                    fontWeight = FontWeight.SemiBold,
                    fontSize = 14.sp
                )
            }

            // Other party name
            if (item.other_party.isNotBlank()) {
                Text(
                    "${if (item.is_given) "To" else "From"}: ${item.other_party}",
                    fontSize = 13.sp,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            // Review text
            if (item.review.isNotBlank()) {
                Spacer(modifier = Modifier.height(6.dp))
                Surface(
                    color = MaterialTheme.colorScheme.surfaceContainerHigh,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text(
                        "\"${item.review}\"",
                        fontSize = 13.sp,
                        modifier = Modifier.padding(8.dp),
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
    }
}

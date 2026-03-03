@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.rating

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.StarBorder
import androidx.compose.material.icons.filled.Warning
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

/**
 * Mandatory Rating Screen — shown after FULLY_COMPLETED.
 * NO skip button. User must rate 1-5 before proceeding.
 * Maps to backend: POST /jobs/:id/rate (mutual rating)
 */
@Composable
fun MandatoryRatingScreen(
    jobId: String,
    rateeName: String,
    rateeRole: String, // "worker" or "client"
    onSubmit: (jobId: String, rating: Int, review: String) -> Unit = { _, _, _ -> }
) {
    var rating by remember { mutableIntStateOf(0) }
    var reviewText by remember { mutableStateOf("") }
    var isSubmitting by remember { mutableStateOf(false) }

    val tags = if (rateeRole == "worker") {
        listOf("Professional", "On Time", "Skilled", "Friendly", "Clean Work", "Good Value")
    } else {
        listOf("Clear Instructions", "Polite", "Easy Location", "Paid on Time", "Respectful")
    }
    var selectedTags by remember { mutableStateOf(setOf<String>()) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Spacer(modifier = Modifier.height(32.dp))

        // Warning banner
        Card(
            shape = RoundedCornerShape(12.dp),
            colors = CardDefaults.cardColors(containerColor = Color(0xFFFFF3E0))
        ) {
            Row(
                modifier = Modifier.padding(12.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(Icons.Default.Warning, contentDescription = null, tint = Color(0xFFFF9800), modifier = Modifier.size(20.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    "Rating is mandatory to continue using the app",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Medium,
                    color = Color(0xFFE65100)
                )
            }
        }

        Spacer(modifier = Modifier.height(32.dp))

        // Avatar
        Box(
            modifier = Modifier
                .size(80.dp)
                .clip(CircleShape)
                .background(MaterialTheme.colorScheme.primaryContainer),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = rateeName.first().uppercase(),
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.primary
            )
        }

        Spacer(modifier = Modifier.height(16.dp))

        Text(
            text = "Rate $rateeName",
            fontSize = 24.sp,
            fontWeight = FontWeight.Bold
        )
        Text(
            text = if (rateeRole == "worker") "How was the service?" else "How was the client?",
            fontSize = 14.sp,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
        )

        Spacer(modifier = Modifier.height(32.dp))

        // Stars
        Row(
            horizontalArrangement = Arrangement.Center,
            modifier = Modifier.fillMaxWidth()
        ) {
            repeat(5) { index ->
                val star = index + 1
                IconButton(
                    onClick = { rating = star },
                    modifier = Modifier.size(52.dp)
                ) {
                    Icon(
                        imageVector = if (rating >= star) Icons.Default.Star else Icons.Default.StarBorder,
                        contentDescription = "$star star",
                        tint = if (rating >= star) Color(0xFFFFC107) else MaterialTheme.colorScheme.outline,
                        modifier = Modifier.fillMaxSize()
                    )
                }
            }
        }

        Text(
            text = when (rating) {
                1 -> "Poor"
                2 -> "Fair"
                3 -> "Good"
                4 -> "Very Good"
                5 -> "Excellent!"
                else -> "Tap to rate"
            },
            fontSize = 16.sp,
            fontWeight = FontWeight.Medium,
            color = if (rating > 0) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.outline,
            modifier = Modifier.height(24.dp)
        )

        Spacer(modifier = Modifier.height(24.dp))

        // Tags
        Text("What stood out?", fontSize = 14.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.align(Alignment.Start))
        Spacer(modifier = Modifier.height(8.dp))

        @OptIn(ExperimentalLayoutApi::class)
        FlowRow(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            tags.forEach { tag ->
                FilterChip(
                    selected = selectedTags.contains(tag),
                    onClick = {
                        selectedTags = if (selectedTags.contains(tag)) selectedTags - tag else selectedTags + tag
                    },
                    label = { Text(tag) }
                )
            }
        }

        Spacer(modifier = Modifier.height(24.dp))

        // Review
        OutlinedTextField(
            value = reviewText,
            onValueChange = { reviewText = it },
            label = { Text("Write a review (optional)") },
            modifier = Modifier.fillMaxWidth(),
            minLines = 3,
            maxLines = 5
        )

        Spacer(modifier = Modifier.weight(1f))

        // Submit — no skip
        Button(
            onClick = {
                isSubmitting = true
                val review = buildString {
                    append(reviewText)
                    if (selectedTags.isNotEmpty()) {
                        if (isNotEmpty()) append(" ")
                        append("[${selectedTags.joinToString(", ")}]")
                    }
                }
                onSubmit(jobId, rating, review)
            },
            enabled = rating > 0 && !isSubmitting,
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp),
            shape = RoundedCornerShape(16.dp)
        ) {
            if (isSubmitting) {
                CircularProgressIndicator(modifier = Modifier.size(24.dp), color = Color.White, strokeWidth = 2.dp)
            } else {
                Text("Submit Rating", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
            }
        }
    }
}

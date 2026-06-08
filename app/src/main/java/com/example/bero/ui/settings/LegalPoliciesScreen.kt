@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.settings

import androidx.compose.foundation.clickable
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
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

/**
 * Data class representing a legal document item
 */
data class LegalDocumentItem(
    val slug: String,
    val title: String,
    val version: String = "v1.0",
    val lastUpdated: String = "March 2026",
    val icon: ImageVector,
    val workerOnly: Boolean = false
)

/**
 * Legal & Policies screen — lists all legal documents
 * Shows in Settings → Legal & Policies
 */
@Composable
fun LegalPoliciesScreen(
    onBackClick: () -> Unit = {},
    onDocumentClick: (slug: String, title: String) -> Unit = { _, _ -> },
    isWorker: Boolean = false
) {
    val documents = remember {
        listOf(
            LegalDocumentItem(
                slug = "terms-conditions",
                title = "Terms & Conditions",
                icon = Icons.Default.Description
            ),
            LegalDocumentItem(
                slug = "privacy-policy",
                title = "Privacy Policy",
                icon = Icons.Default.Security
            ),
            LegalDocumentItem(
                slug = "liability-disclaimer",
                title = "Liability Disclaimer",
                icon = Icons.Default.Shield
            ),
            LegalDocumentItem(
                slug = "dispute-resolution",
                title = "Dispute Resolution Policy",
                icon = Icons.Default.Gavel
            ),
            LegalDocumentItem(
                slug = "worker-responsibility",
                title = "Worker Responsibility Policy",
                icon = Icons.Default.Engineering,
                workerOnly = true
            )
        )
    }

    // Filter worker-only docs when not in worker mode
    val visibleDocuments = if (isWorker) documents else documents.filter { !it.workerOnly }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Legal & Policies", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            item {
                Text(
                    text = "Review our legal documents and policies",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                    modifier = Modifier.padding(bottom = 8.dp)
                )
            }

            items(visibleDocuments) { doc ->
                LegalDocumentCard(
                    document = doc,
                    onClick = { onDocumentClick(doc.slug, doc.title) }
                )
            }

            item {
                Spacer(modifier = Modifier.height(32.dp))
                Text(
                    text = "All documents are governed under Indian laws including the Digital Personal Data Protection Act, 2023 and Consumer Protection Act, 2019.",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
                    lineHeight = 18.sp,
                    modifier = Modifier.padding(horizontal = 4.dp)
                )
            }

            item {
                Spacer(modifier = Modifier.height(80.dp))
            }
        }
    }
}

@Composable
private fun LegalDocumentCard(
    document: LegalDocumentItem,
    onClick: () -> Unit
) {
    Card(
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Surface(
                shape = RoundedCornerShape(10.dp),
                color = MaterialTheme.colorScheme.primaryContainer
            ) {
                Icon(
                    document.icon,
                    contentDescription = null,
                    modifier = Modifier.padding(10.dp),
                    tint = MaterialTheme.colorScheme.primary
                )
            }

            Spacer(modifier = Modifier.width(16.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = document.title,
                    fontWeight = FontWeight.Medium,
                    style = MaterialTheme.typography.bodyLarge
                )
                Spacer(modifier = Modifier.height(2.dp))
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Text(
                        text = document.version,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = "•",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f)
                    )
                    Text(
                        text = "Updated ${document.lastUpdated}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                    )
                }
                if (document.workerOnly) {
                    Spacer(modifier = Modifier.height(4.dp))
                    Surface(
                        shape = RoundedCornerShape(4.dp),
                        color = MaterialTheme.colorScheme.tertiaryContainer
                    ) {
                        Text(
                            text = "Worker Only",
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.tertiary,
                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                        )
                    }
                }
            }

            Icon(
                Icons.Default.ChevronRight,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f)
            )
        }
    }
}

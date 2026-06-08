@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.settings

import android.annotation.SuppressLint
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.viewinterop.AndroidView

/**
 * Maps a document slug to the local asset PDF file path
 */
private fun slugToAssetPath(slug: String): String {
    return when (slug) {
        "terms-conditions" -> "legal/terms_conditions.pdf"
        "privacy-policy" -> "legal/privacy_policy.pdf"
        "liability-disclaimer" -> "legal/liability_disclaimer.pdf"
        "dispute-resolution" -> "legal/dispute_resolution.pdf"
        "worker-responsibility" -> "legal/worker_responsibility.pdf"
        else -> "legal/terms_conditions.pdf"
    }
}

/**
 * PDF Viewer screen using WebView
 * Uses Google Docs viewer for PDF rendering inside WebView
 * Falls back to file:///android_asset for offline support
 */
@SuppressLint("SetJavaScriptEnabled")
@Composable
fun PdfViewerScreen(
    slug: String,
    title: String,
    onBackClick: () -> Unit = {}
) {
    var isLoading by remember { mutableStateOf(true) }
    val assetPath = slugToAssetPath(slug)

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text(title, fontWeight = FontWeight.Bold, maxLines = 1)
                        Text(
                            text = "v1.0 • March 2026",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                        )
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            AndroidView(
                factory = { context ->
                    WebView(context).apply {
                        webViewClient = object : WebViewClient() {
                            override fun onPageFinished(view: WebView?, url: String?) {
                                isLoading = false
                            }
                        }
                        settings.javaScriptEnabled = true
                        settings.builtInZoomControls = true
                        settings.displayZoomControls = false
                        settings.loadWithOverviewMode = true
                        settings.useWideViewPort = true
                        settings.setSupportZoom(true)

                        // Load PDF from assets using file:///android_asset/
                        loadUrl("file:///android_asset/$assetPath")
                    }
                },
                modifier = Modifier.fillMaxSize()
            )

            if (isLoading) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(16.dp)
                    ) {
                        CircularProgressIndicator(
                            color = MaterialTheme.colorScheme.primary
                        )
                        Text(
                            text = "Loading document...",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                    }
                }
            }
        }
    }
}

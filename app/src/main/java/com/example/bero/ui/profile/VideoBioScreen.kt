package com.example.bero.ui.profile

import android.Manifest
import android.net.Uri
import android.view.ViewGroup
import android.widget.LinearLayout
import androidx.camera.core.CameraSelector
import androidx.camera.view.LifecycleCameraController
import androidx.camera.view.PreviewView
import androidx.camera.view.CameraController
import androidx.camera.video.Recording
import androidx.camera.video.VideoRecordEvent
import androidx.camera.view.video.AudioConfig
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.content.ContextCompat
import com.google.accompanist.permissions.ExperimentalPermissionsApi
import com.google.accompanist.permissions.rememberMultiplePermissionsState
import java.io.File
import androidx.camera.video.FileOutputOptions

@OptIn(ExperimentalPermissionsApi::class)
@Composable
fun VideoBioScreen(
    onVideoSaved: (Uri) -> Unit,
    onBack: () -> Unit
) {
    val context = LocalContext.current
    
    val permissionsState = rememberMultiplePermissionsState(
        permissions = listOf(
            Manifest.permission.CAMERA,
            Manifest.permission.RECORD_AUDIO
        )
    )
    
    var isRecording by remember { mutableStateOf(false) }
    var recordingTimeSeconds by remember { mutableLongStateOf(0L) }
    var recordedUri by remember { mutableStateOf<Uri?>(null) }
    
    val lifecycleOwner = LocalLifecycleOwner.current
    
    // Camera controller
    val cameraController = remember {
        LifecycleCameraController(context).apply {
            setEnabledUseCases(CameraController.VIDEO_CAPTURE)
            cameraSelector = CameraSelector.DEFAULT_FRONT_CAMERA
            bindToLifecycle(lifecycleOwner)
        }
    }
    
    // Recording object
    var activeRecording by remember { mutableStateOf<Recording?>(null) }
    
    LaunchedEffect(Unit) {
        if (!permissionsState.allPermissionsGranted) {
            permissionsState.launchMultiplePermissionRequest()
        }
    }
    
    if (!permissionsState.allPermissionsGranted) {
        Box(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center
        ) {
            Text("Camera and Audio permissions are required to record video bio")
        }
        return
    }
    
    if (recordedUri != null) {
        // Review Screen
        VideoReviewScreen(
            uri = recordedUri!!,
            onRetake = { recordedUri = null },
            onSave = { onVideoSaved(recordedUri!!) }
        )
    } else {
        // Recording Screen
        Box(modifier = Modifier.fillMaxSize()) {
            AndroidView(
                factory = { ctx ->
                    PreviewView(ctx).apply {
                        layoutParams = LinearLayout.LayoutParams(
                            ViewGroup.LayoutParams.MATCH_PARENT,
                            ViewGroup.LayoutParams.MATCH_PARENT
                        )
                        scaleType = PreviewView.ScaleType.FILL_CENTER
                        controller = cameraController
                    }
                },
                modifier = Modifier.fillMaxSize()
            )
            
            // Overlays
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(32.dp),
                verticalArrangement = Arrangement.SpaceBetween,
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                // Top bar
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    IconButton(
                        onClick = onBack,
                        modifier = Modifier
                            .background(Color.Black.copy(alpha = 0.5f), CircleShape)
                    ) {
                        Icon(Icons.Default.Close, "Close", tint = Color.White)
                    }
                    
                    // Timer
                    if (isRecording) {
                        Text(
                            text = String.format("00:%02d", recordingTimeSeconds),
                            style = MaterialTheme.typography.titleLarge,
                            color = Color.Red,
                            modifier = Modifier
                                .background(Color.Black.copy(alpha = 0.5f), RoundedCornerShape(8.dp))
                                .padding(horizontal = 12.dp, vertical = 4.dp)
                        )
                    }
                }
                
                // Instructions
                if (!isRecording) {
                    Text(
                        text = "Introduce yourself in 15 seconds",
                        style = MaterialTheme.typography.titleMedium,
                        color = Color.White,
                        modifier = Modifier
                            .background(Color.Black.copy(alpha = 0.5f), RoundedCornerShape(8.dp))
                            .padding(8.dp)
                    )
                }
                
                // Record Button
                Box(
                    modifier = Modifier
                        .size(80.dp)
                        .border(4.dp, Color.White, CircleShape)
                        .clip(CircleShape)
                        .background(if (isRecording) Color.Red else Color.White)
                        .clickable {
                            if (isRecording) {
                                // Stop recording
                                activeRecording?.stop()
                                activeRecording = null
                                isRecording = false
                            } else {
                                // Start recording
                                val outputFile = File(context.cacheDir, "video_bio_${System.currentTimeMillis()}.mp4")
                                val outputOptions = FileOutputOptions.Builder(outputFile).build()
                                
                                activeRecording = cameraController.startRecording(
                                    outputOptions,
                                    AudioConfig.create(true),
                                    ContextCompat.getMainExecutor(context)
                                ) { event ->
                                    if (event is VideoRecordEvent.Finalize) {
                                        if (!event.hasError()) {
                                            recordedUri = Uri.fromFile(outputFile)
                                        } else {
                                            // Handle error
                                            activeRecording?.close()
                                            activeRecording = null
                                            isRecording = false
                                        }
                                    } else if (event is VideoRecordEvent.Status) {
                                        recordingTimeSeconds = event.recordingStats.recordedDurationNanos / 1_000_000_000
                                        if (recordingTimeSeconds >= 15) {
                                            activeRecording?.stop()
                                        }
                                    }
                                }
                                isRecording = true
                            }
                        }
                )
            }
        }
    }
}

@Composable
fun VideoReviewScreen(
    uri: Uri,
    onRetake: () -> Unit,
    onSave: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Text(
            text = "Video Recorded!",
            color = Color.White,
            style = MaterialTheme.typography.headlineMedium
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Text(
            text = "File saved at: $uri",
            color = Color.Gray,
            style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.padding(16.dp)
        )
        
        Spacer(modifier = Modifier.height(32.dp))
        
        Row(
            horizontalArrangement = Arrangement.spacedBy(32.dp)
        ) {
            Button(
                onClick = onRetake,
                colors = ButtonDefaults.buttonColors(containerColor = Color.Gray)
            ) {
                Icon(Icons.Default.Refresh, "Retake")
                Spacer(modifier = Modifier.width(8.dp))
                Text("Retake")
            }
            
            Button(
                onClick = onSave,
                colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.primary)
            ) {
                Icon(Icons.Default.Check, "Save")
                Spacer(modifier = Modifier.width(8.dp))
                Text("Save")
            }
        }
    }
}

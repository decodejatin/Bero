package com.example.bero.ui.settings

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
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

/**
 * Settings screen with app preferences
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    onBackClick: () -> Unit = {}
) {
    var notificationsEnabled by remember { mutableStateOf(true) }
    var soundEnabled by remember { mutableStateOf(true) }
    var darkMode by remember { mutableStateOf(false) }
    var locationEnabled by remember { mutableStateOf(true) }
    var selectedLanguage by remember { mutableStateOf("English") }
    var showLanguageDialog by remember { mutableStateOf(false) }
    
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Settings", fontWeight = FontWeight.Bold) },
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
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Notifications Section
            item {
                SettingsSection(title = "Notifications") {
                    SettingsSwitchItem(
                        icon = Icons.Default.Notifications,
                        title = "Push Notifications",
                        subtitle = "Receive job alerts and updates",
                        checked = notificationsEnabled,
                        onCheckedChange = { notificationsEnabled = it }
                    )
                    HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))
                    SettingsSwitchItem(
                        icon = Icons.Default.VolumeUp,
                        title = "Sound",
                        subtitle = "Play sound for notifications",
                        checked = soundEnabled,
                        onCheckedChange = { soundEnabled = it }
                    )
                }
            }
            
            // Appearance Section
            item {
                SettingsSection(title = "Appearance") {
                    SettingsSwitchItem(
                        icon = Icons.Default.DarkMode,
                        title = "Dark Mode",
                        subtitle = "Use dark theme",
                        checked = darkMode,
                        onCheckedChange = { darkMode = it }
                    )
                    HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))
                    SettingsClickItem(
                        icon = Icons.Default.Language,
                        title = "Language",
                        subtitle = selectedLanguage,
                        onClick = { showLanguageDialog = true }
                    )
                }
            }
            
            // Privacy Section
            item {
                SettingsSection(title = "Privacy & Security") {
                    SettingsSwitchItem(
                        icon = Icons.Default.LocationOn,
                        title = "Location Access",
                        subtitle = "Allow app to access location",
                        checked = locationEnabled,
                        onCheckedChange = { locationEnabled = it }
                    )
                    HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))
                    SettingsClickItem(
                        icon = Icons.Default.Security,
                        title = "Privacy Policy",
                        subtitle = "View our privacy policy",
                        onClick = { }
                    )
                    HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))
                    SettingsClickItem(
                        icon = Icons.Default.Description,
                        title = "Terms of Service",
                        subtitle = "View terms and conditions",
                        onClick = { }
                    )
                }
            }
            
            // About Section
            item {
                SettingsSection(title = "About") {
                    SettingsClickItem(
                        icon = Icons.Default.Info,
                        title = "App Version",
                        subtitle = "1.0.0 (Build 1)",
                        onClick = { }
                    )
                    HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))
                    SettingsClickItem(
                        icon = Icons.Default.Star,
                        title = "Rate Us",
                        subtitle = "Love the app? Rate us!",
                        onClick = { }
                    )
                    HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))
                    SettingsClickItem(
                        icon = Icons.Default.Share,
                        title = "Share App",
                        subtitle = "Tell your friends about Bero",
                        onClick = { }
                    )
                }
            }
            
            // Danger Zone
            item {
                SettingsSection(title = "Danger Zone") {
                    SettingsClickItem(
                        icon = Icons.Default.DeleteForever,
                        title = "Delete Account",
                        subtitle = "Permanently delete your account",
                        onClick = { },
                        isDestructive = true
                    )
                }
            }
            
            item {
                Spacer(modifier = Modifier.height(80.dp))
            }
        }
    }
    
    // Language Selection Dialog
    if (showLanguageDialog) {
        AlertDialog(
            onDismissRequest = { showLanguageDialog = false },
            title = { Text("Select Language") },
            text = {
                Column {
                    listOf("English", "हिंदी (Hindi)", "தமிழ் (Tamil)", "বাংলা (Bengali)", "తెలుగు (Telugu)").forEach { language ->
                        ListItem(
                            headlineContent = { Text(language) },
                            leadingContent = {
                                RadioButton(
                                    selected = selectedLanguage == language.split(" ").first(),
                                    onClick = {
                                        selectedLanguage = language.split(" ").first()
                                        showLanguageDialog = false
                                    }
                                )
                            },
                            modifier = Modifier.clickable {
                                selectedLanguage = language.split(" ").first()
                                showLanguageDialog = false
                            }
                        )
                    }
                }
            },
            confirmButton = {
                TextButton(onClick = { showLanguageDialog = false }) {
                    Text("Cancel")
                }
            }
        )
    }
}

@Composable
private fun SettingsSection(
    title: String,
    content: @Composable ColumnScope.() -> Unit
) {
    Column {
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(bottom = 8.dp)
        )
        
        Card(
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.surface
            )
        ) {
            Column {
                content()
            }
        }
    }
}

@Composable
private fun SettingsSwitchItem(
    icon: ImageVector,
    title: String,
    subtitle: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit
) {
    ListItem(
        headlineContent = {
            Text(text = title, fontWeight = FontWeight.Medium)
        },
        supportingContent = {
            Text(
                text = subtitle,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
            )
        },
        leadingContent = {
            Surface(
                shape = RoundedCornerShape(10.dp),
                color = MaterialTheme.colorScheme.primaryContainer
            ) {
                Icon(
                    icon,
                    contentDescription = null,
                    modifier = Modifier.padding(10.dp),
                    tint = MaterialTheme.colorScheme.primary
                )
            }
        },
        trailingContent = {
            Switch(
                checked = checked,
                onCheckedChange = onCheckedChange
            )
        }
    )
}

@Composable
private fun SettingsClickItem(
    icon: ImageVector,
    title: String,
    subtitle: String,
    onClick: () -> Unit,
    isDestructive: Boolean = false
) {
    val contentColor = if (isDestructive) MaterialTheme.colorScheme.error
                       else MaterialTheme.colorScheme.primary
    
    ListItem(
        headlineContent = {
            Text(
                text = title,
                fontWeight = FontWeight.Medium,
                color = if (isDestructive) MaterialTheme.colorScheme.error
                        else MaterialTheme.colorScheme.onSurface
            )
        },
        supportingContent = {
            Text(
                text = subtitle,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
            )
        },
        leadingContent = {
            Surface(
                shape = RoundedCornerShape(10.dp),
                color = if (isDestructive) MaterialTheme.colorScheme.errorContainer
                        else MaterialTheme.colorScheme.primaryContainer
            ) {
                Icon(
                    icon,
                    contentDescription = null,
                    modifier = Modifier.padding(10.dp),
                    tint = contentColor
                )
            }
        },
        trailingContent = {
            Icon(
                Icons.Default.ChevronRight,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f)
            )
        },
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
    )
}

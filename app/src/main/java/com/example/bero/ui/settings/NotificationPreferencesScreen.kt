@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.settings

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.models.NotificationPreference
import com.example.bero.data.models.NotificationPreferenceType

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationPreferencesScreen(
    onBackClick: () -> Unit = {}
) {
    var preferences by remember {
        mutableStateOf(
            listOf(
                NotificationPreference(NotificationPreferenceType.JOB_ALERTS, true),
                NotificationPreference(NotificationPreferenceType.BOOKING_UPDATES, true),
                NotificationPreference(NotificationPreferenceType.PAYMENT_NOTIFICATIONS, true),
                NotificationPreference(NotificationPreferenceType.CHAT_MESSAGES, true),
                NotificationPreference(NotificationPreferenceType.PROMOTIONAL_OFFERS, false),
                NotificationPreference(NotificationPreferenceType.APP_UPDATES, true)
            )
        )
    }

    var quietHours by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Notifications", fontWeight = FontWeight.Bold) },
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
            verticalArrangement = Arrangement.spacedBy(24.dp)
        ) {
            item {
                Text(
                    text = "Notification Types",
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.Bold,
                    fontSize = 14.sp
                )
            }

            items(preferences) { pref ->
                NotificationSwitchRow(
                    preference = pref,
                    onToggle = { enabled ->
                        preferences = preferences.map { 
                            if (it.type == pref.type) it.copy(enabled = enabled) else it 
                        }
                    }
                )
            }

            item {
                 HorizontalDivider(modifier = Modifier.padding(vertical = 8.dp))
                 Text(
                    text = "Advanced Settings",
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.Bold,
                    fontSize = 14.sp
                )
            }

            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = "Quiet Hours",
                            fontSize = 16.sp,
                            fontWeight = FontWeight.Medium
                        )
                        Text(
                            text = "Mute notifications from 10 PM to 7 AM",
                            fontSize = 12.sp,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                    }
                    Switch(
                        checked = quietHours,
                        onCheckedChange = { quietHours = it }
                    )
                }
            }
        }
    }
}

@Composable
private fun NotificationSwitchRow(
    preference: NotificationPreference,
    onToggle: (Boolean) -> Unit
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = getLabelForType(preference.type),
                fontSize = 16.sp,
                fontWeight = FontWeight.Medium
            )
            Text(
                text = getDescriptionForType(preference.type),
                fontSize = 12.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
            )
        }
        Switch(
            checked = preference.enabled,
            onCheckedChange = onToggle
        )
    }
}

private fun getLabelForType(type: NotificationPreferenceType): String {
    return when (type) {
        NotificationPreferenceType.JOB_ALERTS -> "New Jobs"
        NotificationPreferenceType.BOOKING_UPDATES -> "Booking Status"
        NotificationPreferenceType.PAYMENT_NOTIFICATIONS -> "Payments"
        NotificationPreferenceType.CHAT_MESSAGES -> "Messages"
        NotificationPreferenceType.PROMOTIONAL_OFFERS -> "Offers & Promos"
        NotificationPreferenceType.APP_UPDATES -> "App Updates"
    }
}

private fun getDescriptionForType(type: NotificationPreferenceType): String {
    return when (type) {
        NotificationPreferenceType.JOB_ALERTS -> "Get notified when new jobs match your skills"
        NotificationPreferenceType.BOOKING_UPDATES -> "Updates on your booking requests"
        NotificationPreferenceType.PAYMENT_NOTIFICATIONS -> "Receipts and payout alerts"
        NotificationPreferenceType.CHAT_MESSAGES -> "Messages from clients or workers"
        NotificationPreferenceType.PROMOTIONAL_OFFERS -> "Discounts and special deals"
        NotificationPreferenceType.APP_UPDATES -> "New features and improvements"
    }
}

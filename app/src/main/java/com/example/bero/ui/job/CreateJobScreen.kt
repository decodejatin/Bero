@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.job

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.bero.data.models.ServiceCategory

/**
 * Screen for clients to create/post a new job
 */
@Composable
fun CreateJobScreen(
    clientName: String = "Client",
    viewModel: CreateJobViewModel = viewModel(),
    onBackClick: () -> Unit = {},
    onJobCreated: () -> Unit = {}
) {
    val uiState by viewModel.uiState.collectAsState()
    
    var title by remember { mutableStateOf("") }
    var description by remember { mutableStateOf("") }
    var selectedCategory by remember { mutableStateOf<ServiceCategory?>(null) }
    var address by remember { mutableStateOf("") }
    var locality by remember { mutableStateOf("") }
    var pincode by remember { mutableStateOf("") }
    var paymentAmount by remember { mutableStateOf("") }
    var scheduledDate by remember { mutableStateOf("") }
    var scheduledTimeSlot by remember { mutableStateOf("9 AM - 12 PM") }
    var isUrgent by remember { mutableStateOf(false) }
    var showCategoryDropdown by remember { mutableStateOf(false) }
    var showTimeSlotDropdown by remember { mutableStateOf(false) }
    
    // Category options
    val categories = listOf(
        ServiceCategory.PLUMBER,
        ServiceCategory.ELECTRICIAN,
        ServiceCategory.CARPENTER,
        ServiceCategory.PAINTER,
        ServiceCategory.CLEANER,
        ServiceCategory.AC_REPAIR,
        ServiceCategory.APPLIANCE_REPAIR,
        ServiceCategory.GARDENER
    )
    
    val timeSlots = listOf(
        "6 AM - 9 AM",
        "9 AM - 12 PM",
        "12 PM - 3 PM",
        "3 PM - 6 PM",
        "6 PM - 9 PM"
    )
    
    // Handle success
    LaunchedEffect(uiState.isSuccess) {
        if (uiState.isSuccess) {
            onJobCreated()
            viewModel.resetState()
        }
    }
    
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Post a Job", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
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
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Title
            OutlinedTextField(
                value = title,
                onValueChange = { title = it },
                label = { Text("Job Title") },
                placeholder = { Text("e.g., Kitchen Sink Repair") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                leadingIcon = { Icon(Icons.Default.Work, contentDescription = null) }
            )
            
            // Description
            OutlinedTextField(
                value = description,
                onValueChange = { description = it },
                label = { Text("Description") },
                placeholder = { Text("Describe the work needed...") },
                modifier = Modifier.fillMaxWidth().heightIn(min = 100.dp),
                maxLines = 5,
                leadingIcon = { Icon(Icons.Default.Description, contentDescription = null) }
            )
            
            // Category Dropdown
            ExposedDropdownMenuBox(
                expanded = showCategoryDropdown,
                onExpandedChange = { showCategoryDropdown = it }
            ) {
                OutlinedTextField(
                    value = selectedCategory?.displayName ?: "",
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Service Category") },
                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = showCategoryDropdown) },
                    modifier = Modifier.fillMaxWidth().menuAnchor(),
                    leadingIcon = { 
                        Text(selectedCategory?.emoji ?: "📋", fontSize = 20.sp) 
                    }
                )
                ExposedDropdownMenu(
                    expanded = showCategoryDropdown,
                    onDismissRequest = { showCategoryDropdown = false }
                ) {
                    categories.forEach { category ->
                        DropdownMenuItem(
                            text = { Text("${category.emoji} ${category.displayName}") },
                            onClick = {
                                selectedCategory = category
                                showCategoryDropdown = false
                            }
                        )
                    }
                }
            }
            
            // Address
            OutlinedTextField(
                value = address,
                onValueChange = { address = it },
                label = { Text("Full Address") },
                placeholder = { Text("House no., Building, Street") },
                modifier = Modifier.fillMaxWidth(),
                leadingIcon = { Icon(Icons.Default.Home, contentDescription = null) }
            )
            
            // Locality and Pincode Row
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                OutlinedTextField(
                    value = locality,
                    onValueChange = { locality = it },
                    label = { Text("Locality") },
                    placeholder = { Text("Sector 15") },
                    modifier = Modifier.weight(1f),
                    singleLine = true
                )
                OutlinedTextField(
                    value = pincode,
                    onValueChange = { if (it.length <= 6) pincode = it },
                    label = { Text("Pincode") },
                    placeholder = { Text("110001") },
                    modifier = Modifier.weight(1f),
                    singleLine = true,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number)
                )
            }
            
            // Payment Amount
            OutlinedTextField(
                value = paymentAmount,
                onValueChange = { paymentAmount = it.filter { ch -> ch.isDigit() } },
                label = { Text("Budget (₹)") },
                placeholder = { Text("500") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                leadingIcon = { Icon(Icons.Default.CurrencyRupee, contentDescription = null) },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number)
            )
            
            // Scheduled Date
            OutlinedTextField(
                value = scheduledDate,
                onValueChange = { scheduledDate = it },
                label = { Text("Scheduled Date") },
                placeholder = { Text("Today, Tomorrow, or 30 Jan") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                leadingIcon = { Icon(Icons.Default.CalendarToday, contentDescription = null) }
            )
            
            // Time Slot Dropdown
            ExposedDropdownMenuBox(
                expanded = showTimeSlotDropdown,
                onExpandedChange = { showTimeSlotDropdown = it }
            ) {
                OutlinedTextField(
                    value = scheduledTimeSlot,
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Time Slot") },
                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = showTimeSlotDropdown) },
                    modifier = Modifier.fillMaxWidth().menuAnchor(),
                    leadingIcon = { Icon(Icons.Default.Schedule, contentDescription = null) }
                )
                ExposedDropdownMenu(
                    expanded = showTimeSlotDropdown,
                    onDismissRequest = { showTimeSlotDropdown = false }
                ) {
                    timeSlots.forEach { slot ->
                        DropdownMenuItem(
                            text = { Text(slot) },
                            onClick = {
                                scheduledTimeSlot = slot
                                showTimeSlotDropdown = false
                            }
                        )
                    }
                }
            }
            
            // Urgent Toggle
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(
                    containerColor = if (isUrgent) Color(0xFFFFE0E0) else MaterialTheme.colorScheme.surfaceVariant
                )
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column {
                        Text(
                            text = "🔥 Mark as Urgent",
                            fontWeight = FontWeight.SemiBold
                        )
                        Text(
                            text = "Workers will be notified immediately",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                    }
                    Switch(
                        checked = isUrgent,
                        onCheckedChange = { isUrgent = it }
                    )
                }
            }
            
            // Error Message
            uiState.error?.let { error ->
                Card(
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.errorContainer)
                ) {
                    Text(
                        text = error,
                        modifier = Modifier.padding(16.dp),
                        color = MaterialTheme.colorScheme.onErrorContainer
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // Submit Button
            Button(
                onClick = {
                    selectedCategory?.let { category ->
                        viewModel.createJob(
                            title = title,
                            description = description,
                            category = category,
                            clientName = clientName,
                            address = address,
                            locality = locality,
                            pincode = pincode,
                            paymentAmount = paymentAmount.toDoubleOrNull() ?: 0.0,
                            scheduledDate = scheduledDate,
                            scheduledTimeSlot = scheduledTimeSlot,
                            isUrgent = isUrgent
                        )
                    }
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                enabled = !uiState.isLoading && 
                         title.isNotBlank() && 
                         description.isNotBlank() && 
                         selectedCategory != null &&
                         address.isNotBlank() &&
                         locality.isNotBlank() &&
                         paymentAmount.isNotBlank() &&
                         scheduledDate.isNotBlank(),
                shape = RoundedCornerShape(16.dp)
            ) {
                if (uiState.isLoading) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(24.dp),
                        color = MaterialTheme.colorScheme.onPrimary
                    )
                } else {
                    Icon(Icons.Default.Check, contentDescription = null)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Post Job", fontWeight = FontWeight.Bold, fontSize = 16.sp)
                }
            }
            
            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

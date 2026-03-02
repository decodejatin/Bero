package com.example.bero.ui.profile

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
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.CreateAddressRequest
import com.example.bero.data.network.SavedAddressDto
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SavedAddressesScreen(
    apiClient: BeroApiClient? = null,
    onBackClick: () -> Unit = {}
) {
    var addresses by remember { mutableStateOf<List<SavedAddressDto>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var showAddDialog by remember { mutableStateOf(false) }
    var editingAddress by remember { mutableStateOf<SavedAddressDto?>(null) }
    val scope = rememberCoroutineScope()

    fun loadAddresses() {
        scope.launch {
            isLoading = true
            apiClient?.getAddresses()?.onSuccess { addresses = it }
            isLoading = false
        }
    }

    LaunchedEffect(Unit) { loadAddresses() }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Saved Addresses", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { showAddDialog = true }) {
                        Icon(Icons.Default.Add, contentDescription = "Add Address")
                    }
                }
            )
        }
    ) { padding ->
        if (isLoading) {
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else if (addresses.isEmpty()) {
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Icon(
                        Icons.Default.LocationOff,
                        contentDescription = null,
                        modifier = Modifier.size(64.dp),
                        tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f)
                    )
                    Spacer(Modifier.height(16.dp))
                    Text(
                        "No saved addresses",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                    Spacer(Modifier.height(8.dp))
                    Text(
                        "Tap + to add your first address",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f)
                    )
                }
            }
        } else {
            LazyColumn(
                modifier = Modifier.fillMaxSize().padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(addresses) { address ->
                    AddressCard(
                        address = address,
                        onEdit = { editingAddress = address },
                        onDelete = {
                            scope.launch {
                                apiClient?.deleteAddress(address.id)
                                loadAddresses()
                            }
                        }
                    )
                }
                item { Spacer(Modifier.height(80.dp)) }
            }
        }
    }

    // Add/Edit Dialog
    if (showAddDialog || editingAddress != null) {
        AddressDialog(
            existing = editingAddress,
            onDismiss = {
                showAddDialog = false
                editingAddress = null
            },
            onSave = { label, fullAddress, isDefault ->
                scope.launch {
                    val request = CreateAddressRequest(
                        label = label,
                        full_address = fullAddress,
                        is_default = isDefault
                    )
                    if (editingAddress != null) {
                        apiClient?.updateAddress(editingAddress!!.id, request)
                    } else {
                        apiClient?.createAddress(request)
                    }
                    showAddDialog = false
                    editingAddress = null
                    loadAddresses()
                }
            }
        )
    }
}

@Composable
private fun AddressCard(
    address: SavedAddressDto,
    onEdit: () -> Unit,
    onDelete: () -> Unit
) {
    Card(
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        modifier = Modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Surface(
                shape = RoundedCornerShape(12.dp),
                color = if (address.is_default) MaterialTheme.colorScheme.primaryContainer
                        else MaterialTheme.colorScheme.surfaceVariant
            ) {
                Icon(
                    when (address.label.lowercase()) {
                        "home" -> Icons.Default.Home
                        "work" -> Icons.Default.Work
                        else -> Icons.Default.LocationOn
                    },
                    contentDescription = null,
                    modifier = Modifier.padding(12.dp),
                    tint = if (address.is_default) MaterialTheme.colorScheme.primary
                           else MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            
            Spacer(Modifier.width(16.dp))
            
            Column(modifier = Modifier.weight(1f)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        address.label,
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.Bold
                    )
                    if (address.is_default) {
                        Spacer(Modifier.width(8.dp))
                        Surface(
                            shape = RoundedCornerShape(4.dp),
                            color = MaterialTheme.colorScheme.primary
                        ) {
                            Text(
                                "DEFAULT",
                                style = MaterialTheme.typography.labelSmall,
                                color = MaterialTheme.colorScheme.onPrimary,
                                modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                            )
                        }
                    }
                }
                Spacer(Modifier.height(4.dp))
                Text(
                    address.full_address,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                    maxLines = 2
                )
            }
            
            IconButton(onClick = onEdit) {
                Icon(Icons.Default.Edit, contentDescription = "Edit", tint = MaterialTheme.colorScheme.primary)
            }
            IconButton(onClick = onDelete) {
                Icon(Icons.Default.Delete, contentDescription = "Delete", tint = MaterialTheme.colorScheme.error)
            }
        }
    }
}

@Composable
private fun AddressDialog(
    existing: SavedAddressDto?,
    onDismiss: () -> Unit,
    onSave: (label: String, fullAddress: String, isDefault: Boolean) -> Unit
) {
    var label by remember { mutableStateOf(existing?.label ?: "Home") }
    var fullAddress by remember { mutableStateOf(existing?.full_address ?: "") }
    var isDefault by remember { mutableStateOf(existing?.is_default ?: false) }
    val labels = listOf("Home", "Work", "Other")

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(if (existing != null) "Edit Address" else "Add Address") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                // Label selector
                Text("Label", style = MaterialTheme.typography.labelMedium)
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    labels.forEach { l ->
                        FilterChip(
                            selected = label == l,
                            onClick = { label = l },
                            label = { Text(l) }
                        )
                    }
                }
                
                OutlinedTextField(
                    value = fullAddress,
                    onValueChange = { fullAddress = it },
                    label = { Text("Full Address") },
                    modifier = Modifier.fillMaxWidth(),
                    minLines = 2,
                    maxLines = 4
                )
                
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.fillMaxWidth().clickable { isDefault = !isDefault }
                ) {
                    Checkbox(checked = isDefault, onCheckedChange = { isDefault = it })
                    Spacer(Modifier.width(8.dp))
                    Text("Set as default address")
                }
            }
        },
        confirmButton = {
            TextButton(
                onClick = { if (fullAddress.isNotBlank()) onSave(label, fullAddress, isDefault) },
                enabled = fullAddress.isNotBlank()
            ) {
                Text("Save")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text("Cancel") }
        }
    )
}

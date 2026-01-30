@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.payment

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
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
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.models.PaymentMethod
import com.example.bero.data.models.PaymentMethodType
import java.util.UUID

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PaymentScreen(
    onBackClick: () -> Unit = {}
) {
    var paymentMethods by remember {
        mutableStateOf(
            listOf(
                PaymentMethod(
                    type = PaymentMethodType.UPI,
                    displayName = "Google Pay",
                    upiId = "user@okaxis",
                    isDefault = true
                ),
                PaymentMethod(
                    type = PaymentMethodType.CREDIT_CARD,
                    displayName = "HDFC Bank Credit Card",
                    lastFourDigits = "4589"
                )
            )
        )
    }

    var showAddDialog by remember { mutableStateOf(false) }

    if (showAddDialog) {
        AddPaymentMethodDialog(
            onDismiss = { showAddDialog = false },
            onAdd = { method ->
                paymentMethods = paymentMethods + method
                showAddDialog = false
            }
        )
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Payments", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { showAddDialog = true },
                containerColor = MaterialTheme.colorScheme.primary
            ) {
                Icon(Icons.Default.Add, contentDescription = "Add Payment Method")
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(20.dp)
            ) {
                // Payment Methods Section
                item {
                    Text(
                        text = "Saved Payment Methods",
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.padding(bottom = 8.dp)
                    )
                }

                items(paymentMethods) { method ->
                    PaymentMethodCard(
                        method = method,
                        onDelete = {
                            paymentMethods = paymentMethods.filter { it.id != method.id }
                        },
                        onSetDefault = {
                            paymentMethods = paymentMethods.map {
                                it.copy(isDefault = it.id == method.id)
                            }
                        }
                    )
                }

                if (paymentMethods.isEmpty()) {
                    item {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(32.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = "No payment methods added yet",
                                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                            )
                        }
                    }
                }

                // Recent Transactions (Dummy)
                item {
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        text = "Recent Payments",
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.padding(bottom = 8.dp)
                    )
                }

                item {
                    TransactionItem(
                        title = "Plumbing Service",
                        date = "Today, 10:30 AM",
                        amount = "₹450",
                        status = "Success"
                    )
                    HorizontalDivider(modifier = Modifier.padding(vertical = 12.dp))
                    TransactionItem(
                        title = "AC Repair Advance",
                        date = "Yesterday, 4:15 PM",
                        amount = "₹200",
                        status = "Success"
                    )
                    HorizontalDivider(modifier = Modifier.padding(vertical = 12.dp))
                    TransactionItem(
                        title = "House Cleaning",
                        date = "28 Jan, 11:00 AM",
                        amount = "₹1,200",
                        status = "Failed",
                        statusColor = Color(0xFFF44336)
                    )
                }
            }
        }
    }
}

@Composable
private fun PaymentMethodCard(
    method: PaymentMethod,
    onDelete: () -> Unit,
    onSetDefault: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface,
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Icon based on type
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .background(
                            MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.3f),
                            RoundedCornerShape(8.dp)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        imageVector = when (method.type) {
                            PaymentMethodType.CREDIT_CARD, PaymentMethodType.DEBIT_CARD -> Icons.Default.CreditCard
                            PaymentMethodType.UPI -> Icons.Default.QrCode
                            PaymentMethodType.WALLET -> Icons.Default.AccountBalanceWallet
                            PaymentMethodType.NET_BANKING -> Icons.Default.AccountBalance
                        },
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.primary
                    )
                }

                Spacer(modifier = Modifier.width(16.dp))

                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = method.displayName,
                        fontWeight = FontWeight.SemiBold,
                        fontSize = 16.sp
                    )
                    Text(
                        text = when (method.type) {
                            PaymentMethodType.CREDIT_CARD, PaymentMethodType.DEBIT_CARD -> "**** **** **** ${method.lastFourDigits}"
                            PaymentMethodType.UPI -> method.upiId ?: ""
                            else -> method.type.name
                        },
                        fontSize = 14.sp,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                }

                if (method.isDefault) {
                    Surface(
                        color = Color(0xFFE8F5E9),
                        shape = RoundedCornerShape(4.dp)
                    ) {
                        Text(
                            text = "DEFAULT",
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFF2E7D32),
                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End
            ) {
                if (!method.isDefault) {
                    TextButton(onClick = onSetDefault) {
                        Text("Set as Default")
                    }
                }
                IconButton(onClick = onDelete) {
                    Icon(
                        Icons.Default.Delete,
                        contentDescription = "Delete",
                        tint = Color(0xFFF44336)
                    )
                }
            }
        }
    }
}

@Composable
private fun TransactionItem(
    title: String,
    date: String,
    amount: String,
    status: String,
    statusColor: Color = Color(0xFF4CAF50)
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = title,
                fontWeight = FontWeight.Medium,
                fontSize = 16.sp
            )
            Text(
                text = date,
                fontSize = 12.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
            )
        }

        Column(horizontalAlignment = Alignment.End) {
            Text(
                text = "-$amount",
                fontWeight = FontWeight.Bold,
                fontSize = 16.sp
            )
            Text(
                text = status,
                fontSize = 12.sp,
                color = statusColor
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun AddPaymentMethodDialog(
    onDismiss: () -> Unit,
    onAdd: (PaymentMethod) -> Unit
) {
    var selectedType by remember { mutableStateOf(PaymentMethodType.UPI) }
    var upiId by remember { mutableStateOf("") }
    var cardNumber by remember { mutableStateOf("") }
    var cardName by remember { mutableStateOf("") }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Add Payment Method") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                // Type selection could be chips or dropdown, simple buttons for now
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceEvenly) {
                     FilterChip(
                         selected = selectedType == PaymentMethodType.UPI,
                         onClick = { selectedType = PaymentMethodType.UPI },
                         label = { Text("UPI") }
                     )
                    FilterChip(
                        selected = selectedType == PaymentMethodType.CREDIT_CARD,
                        onClick = { selectedType = PaymentMethodType.CREDIT_CARD },
                        label = { Text("Card") }
                    )
                }

                if (selectedType == PaymentMethodType.UPI) {
                    OutlinedTextField(
                        value = upiId,
                        onValueChange = { upiId = it },
                        label = { Text("UPI ID") },
                        placeholder = { Text("example@upi") },
                        modifier = Modifier.fillMaxWidth()
                    )
                } else {
                    OutlinedTextField(
                        value = cardNumber,
                        onValueChange = { if (it.length <= 16) cardNumber = it },
                        label = { Text("Card Number") },
                        modifier = Modifier.fillMaxWidth()
                    )
                    OutlinedTextField(
                        value = cardName,
                        onValueChange = { cardName = it },
                        label = { Text("Name on Card") },
                        modifier = Modifier.fillMaxWidth()
                    )
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    val method = if (selectedType == PaymentMethodType.UPI) {
                        PaymentMethod(
                            type = PaymentMethodType.UPI,
                            displayName = "UPI - $upiId",
                            upiId = upiId
                        )
                    } else {
                        PaymentMethod(
                            type = PaymentMethodType.CREDIT_CARD,
                            displayName = cardName.ifBlank { "Card" },
                            lastFourDigits = cardNumber.takeLast(4)
                        )
                    }
                    onAdd(method)
                },
                enabled = (selectedType == PaymentMethodType.UPI && upiId.isNotBlank()) ||
                        (selectedType == PaymentMethodType.CREDIT_CARD && cardNumber.length >= 16)
            ) {
                Text("Add")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel")
            }
        }
    )
}

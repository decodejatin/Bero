package com.example.bero.ui.wallet

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.DummyDataProvider
import com.example.bero.data.models.*
import java.text.SimpleDateFormat
import java.util.*

/**
 * Wallet screen showing balance and transaction history
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WalletScreen() {
    var selectedTab by remember { mutableIntStateOf(0) }
    
    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        // Balance Card
        item {
            BalanceCard()
        }
        
        // Quick Actions
        item {
            QuickActionsRow()
        }
        
        // Stats Cards
        item {
            StatsRow()
        }
        
        // Transaction History Header
        item {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Transaction History",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold
                )
                TextButton(onClick = { }) {
                    Text("See All")
                }
            }
        }
        
        // Transactions
        items(DummyDataProvider.sampleTransactions) { transaction ->
            TransactionItem(transaction)
        }
        
        item {
            Spacer(modifier = Modifier.height(100.dp))
        }
    }
}

@Composable
private fun BalanceCard() {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        shape = RoundedCornerShape(24.dp),
        colors = CardDefaults.cardColors(
            containerColor = Color.Transparent
        )
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(
                    brush = Brush.linearGradient(
                        colors = listOf(
                            Color(0xFF667eea),
                            Color(0xFF764ba2)
                        )
                    ),
                    shape = RoundedCornerShape(24.dp)
                )
                .padding(24.dp)
        ) {
            Column {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.Top
                ) {
                    Column {
                        Text(
                            text = "Available Balance",
                            style = MaterialTheme.typography.bodyMedium,
                            color = Color.White.copy(alpha = 0.8f)
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "₹${String.format("%,.2f", DummyDataProvider.currentWalletBalance)}",
                            style = MaterialTheme.typography.displaySmall,
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )
                    }
                    
                    Surface(
                        shape = CircleShape,
                        color = Color.White.copy(alpha = 0.2f)
                    ) {
                        Icon(
                            Icons.Default.AccountBalanceWallet,
                            contentDescription = null,
                            modifier = Modifier.padding(12.dp),
                            tint = Color.White
                        )
                    }
                }
                
                Spacer(modifier = Modifier.height(20.dp))
                
                // Pending balance
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(12.dp))
                        .background(Color.White.copy(alpha = 0.15f))
                        .padding(12.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(
                            Icons.Outlined.Schedule,
                            contentDescription = null,
                            tint = Color.White.copy(alpha = 0.8f),
                            modifier = Modifier.size(18.dp)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "Pending",
                            style = MaterialTheme.typography.bodyMedium,
                            color = Color.White.copy(alpha = 0.8f)
                        )
                    }
                    Text(
                        text = "₹${String.format("%,.2f", DummyDataProvider.pendingBalance)}",
                        style = MaterialTheme.typography.bodyLarge,
                        fontWeight = FontWeight.SemiBold,
                        color = Color.White
                    )
                }
            }
        }
    }
}

@Composable
private fun QuickActionsRow() {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        QuickActionButton(
            icon = Icons.Outlined.AccountBalance,
            label = "Withdraw",
            onClick = { },
            modifier = Modifier.weight(1f),
            containerColor = MaterialTheme.colorScheme.primaryContainer
        )
        QuickActionButton(
            icon = Icons.Outlined.Add,
            label = "Add Money",
            onClick = { },
            modifier = Modifier.weight(1f),
            containerColor = MaterialTheme.colorScheme.secondaryContainer
        )
        QuickActionButton(
            icon = Icons.Outlined.History,
            label = "History",
            onClick = { },
            modifier = Modifier.weight(1f),
            containerColor = MaterialTheme.colorScheme.tertiaryContainer
        )
    }
}

@Composable
private fun QuickActionButton(
    icon: ImageVector,
    label: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    containerColor: Color
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = containerColor),
        onClick = onClick
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                icon,
                contentDescription = null,
                modifier = Modifier.size(28.dp)
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = label,
                style = MaterialTheme.typography.labelMedium,
                fontWeight = FontWeight.Medium
            )
        }
    }
}

@Composable
private fun StatsRow() {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        StatCard(
            title = "This Month",
            value = "₹${String.format("%,.0f", DummyDataProvider.totalEarningsThisMonth)}",
            icon = Icons.Default.TrendingUp,
            trendUp = true,
            modifier = Modifier.weight(1f)
        )
        StatCard(
            title = "Jobs Done",
            value = "${DummyDataProvider.totalJobsThisMonth}",
            icon = Icons.Default.CheckCircle,
            modifier = Modifier.weight(1f)
        )
    }
}

@Composable
private fun StatCard(
    title: String,
    value: String,
    icon: ImageVector,
    trendUp: Boolean = false,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
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
                    text = title,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f)
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = value,
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold
                )
            }
            
            Surface(
                shape = CircleShape,
                color = if (trendUp) Color(0xFF4CAF50).copy(alpha = 0.15f) 
                        else MaterialTheme.colorScheme.primary.copy(alpha = 0.15f)
            ) {
                Icon(
                    icon,
                    contentDescription = null,
                    modifier = Modifier.padding(10.dp),
                    tint = if (trendUp) Color(0xFF4CAF50) else MaterialTheme.colorScheme.primary
                )
            }
        }
    }
}

@Composable
private fun TransactionItem(transaction: Transaction) {
    val (icon, iconColor, bgColor) = when (transaction.type) {
        TransactionType.JOB_EARNING -> Triple(Icons.Default.WorkOutline, Color(0xFF4CAF50), Color(0xFF4CAF50).copy(alpha = 0.1f))
        TransactionType.COMMISSION_DEDUCTION -> Triple(Icons.Default.Percent, Color(0xFFFF9800), Color(0xFFFF9800).copy(alpha = 0.1f))
        TransactionType.TDS_DEDUCTION -> Triple(Icons.Default.Receipt, Color(0xFFFF5722), Color(0xFFFF5722).copy(alpha = 0.1f))
        TransactionType.GST_DEDUCTION -> Triple(Icons.Default.Receipt, Color(0xFFFF5722), Color(0xFFFF5722).copy(alpha = 0.1f))
        TransactionType.WALLET_RECHARGE -> Triple(Icons.Default.Add, Color(0xFF2196F3), Color(0xFF2196F3).copy(alpha = 0.1f))
        TransactionType.WITHDRAWAL -> Triple(Icons.Default.AccountBalance, Color(0xFF9C27B0), Color(0xFF9C27B0).copy(alpha = 0.1f))
        TransactionType.BONUS -> Triple(Icons.Default.CardGiftcard, Color(0xFFE91E63), Color(0xFFE91E63).copy(alpha = 0.1f))
        TransactionType.STREAK_REWARD -> Triple(Icons.Default.LocalFireDepartment, Color(0xFFFF5722), Color(0xFFFF5722).copy(alpha = 0.1f))
    }
    
    val isPositive = transaction.amountRupees > 0
    val timeFormat = SimpleDateFormat("hh:mm a", Locale.getDefault())
    val dateFormat = SimpleDateFormat("dd MMM", Locale.getDefault())
    
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        // Icon
        Surface(
            modifier = Modifier.size(48.dp),
            shape = CircleShape,
            color = bgColor
        ) {
            Icon(
                icon,
                contentDescription = null,
                modifier = Modifier.padding(12.dp),
                tint = iconColor
            )
        }
        
        Spacer(modifier = Modifier.width(12.dp))
        
        // Details
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = transaction.description,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                maxLines = 1
            )
            Spacer(modifier = Modifier.height(2.dp))
            Text(
                text = "${dateFormat.format(Date(transaction.timestamp))} • ${timeFormat.format(Date(transaction.timestamp))}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
            )
        }
        
        // Amount
        Text(
            text = "${if (isPositive) "+" else ""}₹${String.format("%,.0f", kotlin.math.abs(transaction.amountRupees))}",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold,
            color = if (isPositive) Color(0xFF4CAF50) else Color(0xFFE53935)
        )
    }
}

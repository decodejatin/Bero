package com.example.bero

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Build
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.bero.data.auth.AuthState
import com.example.bero.data.models.WorkerProfile
import com.example.bero.data.models.WorkerTier
import com.example.bero.utils.getPlatform
import com.example.bero.data.payment.TaxCalculator
import com.example.bero.ui.auth.AuthViewModel
import com.example.bero.ui.auth.LoginScreen
import com.example.bero.ui.auth.OtpVerificationScreen
import com.example.bero.ui.auth.RoleSelectionScreen
import com.example.bero.ui.job.PostJobScreen
import com.example.bero.data.models.UserType
import com.example.bero.ui.profile.VideoBioScreen
import com.example.bero.data.models.KycStatus
import com.example.bero.ui.kyc.KycVerificationScreen
import com.example.bero.ui.kyc.KycViewModel
import com.example.bero.ui.theme.BeroTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            BeroTheme {
                BeroApp()
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BeroApp() {
    val authViewModel: AuthViewModel = viewModel()
    val authState by authViewModel.authState.collectAsState()
    val uiState by authViewModel.uiState.collectAsState()
    
    // Route based on auth state
    when {
        // Show OTP screen if OTP was sent
        uiState.otpSent && !uiState.otpVerified -> {
            OtpVerificationScreen(
                phoneNumber = uiState.phoneNumber ?: "",
                onOtpSubmit = { otp -> authViewModel.verifyOtp(otp) },
                onResendOtp = { authViewModel.resendOtp() },
                onBackClick = { authViewModel.goBackToLogin() },
                isLoading = uiState.isLoading,
                error = uiState.error
            )
        }
        // Show login for unauthenticated or authenticating states
        authState is AuthState.NotAuthenticated || 
        authState is AuthState.Authenticating -> {
            LoginScreen(
                onPhoneSubmit = { phone -> authViewModel.requestOtp(phone) },
                onTruecallerClick = { 
                    // TODO: Implement Truecaller SDK integration
                    // For now, show a placeholder
                },
                isLoading = uiState.isLoading,
                error = uiState.error
            )
        }
        // Show role selection if needed
        authState is AuthState.RequiresRoleSelection -> {
            RoleSelectionScreen(
                onRoleSelected = { role -> authViewModel.selectRole(role) }
            )
        }
        // Show KYC screen for users who need KYC
        authState is AuthState.RequiresKyc -> {
            val kycViewModel: KycViewModel = viewModel()
            var isVerifyingKyc by remember { mutableStateOf(false) }

            if (isVerifyingKyc) {
                KycVerificationScreen(
                    onComplete = {
                         authViewModel.updateKycStatus(KycStatus.VERIFIED)
                         isVerifyingKyc = false
                    },
                    onBackClick = { isVerifyingKyc = false },
                    viewModel = kycViewModel
                )
            } else {
                KycPendingScreen(
                    onStartKyc = { 
                        kycViewModel.initSession((authState as AuthState.RequiresKyc).user.id)
                        isVerifyingKyc = true 
                    },
                    onLogout = { authViewModel.logout() }
                )
            }
        }
        // Show video bio screen for users who need to record video
        authState is AuthState.RequiresVideoBio -> {
            var isRecordingVideo by remember { mutableStateOf(false) }
            
            if (isRecordingVideo) {
                VideoBioScreen(
                    onVideoSaved = { uri -> 
                        isRecordingVideo = false
                        authViewModel.updateVideoBioStatus(true) 
                    },
                    onBack = { isRecordingVideo = false }
                )
            } else {
                VideoBioPendingScreen(
                    onRecordVideo = { isRecordingVideo = true },
                    onLogout = { authViewModel.logout() }
                )
            }
        }
        // Show main app for authenticated users
        authState is AuthState.Authenticated -> {
            val user = (authState as AuthState.Authenticated).user
            MainAppScreen(
                userType = user.userType,
                onLogout = { authViewModel.logout() }
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MainAppScreen(
    userType: UserType,
    onLogout: () -> Unit
) {
    var selectedTab by remember { mutableIntStateOf(0) }
    
    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = {
                    Text(
                        text = "Bero",
                        fontWeight = FontWeight.Bold,
                        fontSize = 24.sp
                    )
                },
                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                    containerColor = MaterialTheme.colorScheme.primary,
                    titleContentColor = MaterialTheme.colorScheme.onPrimary
                )
            )
        },
        bottomBar = {
            NavigationBar {
                NavigationBarItem(
                    icon = { Icon(Icons.Default.Home, contentDescription = "Home") },
                    label = { Text("Home") },
                    selected = selectedTab == 0,
                    onClick = { selectedTab = 0 }
                )
                
                if (userType == UserType.WORKER) {
                    NavigationBarItem(
                        icon = { Icon(Icons.Default.Build, contentDescription = "Jobs") },
                        label = { Text("Jobs") },
                        selected = selectedTab == 1,
                        onClick = { selectedTab = 1 }
                    )
                } else {
                    NavigationBarItem(
                        icon = { Icon(Icons.Default.Add, contentDescription = "Post") },
                        label = { Text("Post Job") },
                        selected = selectedTab == 1,
                        onClick = { selectedTab = 1 }
                    )
                }
                
                NavigationBarItem(
                    icon = { Icon(Icons.Default.Person, contentDescription = "Profile") },
                    label = { Text("Profile") },
                    selected = selectedTab == 2,
                    onClick = { selectedTab = 2 }
                )
            }
        }
    ) { innerPadding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
        ) {
            when (selectedTab) {
                0 -> HomeScreen(userType)
                1 -> if (userType == UserType.WORKER) JobsScreen() else PostJobScreen()
                2 -> ProfileScreen(userType, onLogout = onLogout)
            }
        }
    }
}

@Composable
fun KycPendingScreen(
    onStartKyc: () -> Unit,
    onLogout: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Text(
            text = "🪪",
            fontSize = 80.sp
        )
        
        Spacer(modifier = Modifier.height(24.dp))
        
        Text(
            text = "Complete KYC Verification",
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold,
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = "Verify your Aadhaar to start accepting jobs\nand build trust with customers",
            style = MaterialTheme.typography.bodyLarge,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(32.dp))
        
        Button(
            onClick = onStartKyc,
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp)
        ) {
            Text("Start Aadhaar KYC", fontWeight = FontWeight.SemiBold)
        }
        
        Spacer(modifier = Modifier.height(16.dp))
        
        TextButton(onClick = onLogout) {
            Text("Logout")
        }
    }
}

@Composable
fun VideoBioPendingScreen(
    onRecordVideo: () -> Unit,
    onLogout: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Text(
            text = "🎥",
            fontSize = 80.sp
        )
        
        Spacer(modifier = Modifier.height(24.dp))
        
        Text(
            text = "Record Your Video Bio",
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold,
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = "A short 15-second video helps customers\ntrust you and increases your job bookings!",
            style = MaterialTheme.typography.bodyLarge,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(32.dp))
        
        Button(
            onClick = onRecordVideo,
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp)
        ) {
            Text("Record Video Bio", fontWeight = FontWeight.SemiBold)
        }
        
        Spacer(modifier = Modifier.height(16.dp))
        
        TextButton(onClick = onLogout) {
            Text("Logout")
        }
    }
}

@Composable
fun HomeScreen(userType: UserType) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(24.dp)
    ) {
        Spacer(modifier = Modifier.height(32.dp))
        
        // Welcome Card
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.primaryContainer
            )
        ) {
            Column(
                modifier = Modifier.padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text(
                    text = "Welcome to Bero",
                    style = MaterialTheme.typography.headlineMedium,
                    fontWeight = FontWeight.Bold
                )
                Spacer(modifier = Modifier.height(8.dp))
                // Change message based on role
                Text(
                    text = if (userType == UserType.WORKER) "Ready to find your next job?" else "Find the best workers for your needs",
                    style = MaterialTheme.typography.bodyLarge,
                    textAlign = TextAlign.Center,
                    color = MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.8f)
                )
            }
        }
        
        if (userType == UserType.WORKER) {
            // Platform Info (from shared module)
            Text(
                text = "Running on: ${getPlatform().name}",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f)
            )
            
            // Demo: Tax Calculator from shared module
            DemoTaxCalculatorCard()
        } else {
            // Client specific home content
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.secondaryContainer
                )
            ) {
                Row(
                    modifier = Modifier.padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("🔍", fontSize = 32.sp)
                    Spacer(modifier = Modifier.width(16.dp))
                    Column {
                        Text(
                            text = "Browse Categories",
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold
                        )
                        Text("Plumbers, Electricians, and more")
                    }
                }
            }
        }
    }
}

@Composable
fun DemoTaxCalculatorCard() {
    val demoWorker = WorkerProfile(
        userId = "demo-user",
        skills = listOf("Plumber", "Electrician"),
        streakCount = 0,
        tier = WorkerTier.BRONZE
    )
    val breakdown = TaxCalculator.calculateBreakdown(1000.0, demoWorker)
    
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "💰 Transaction Example (₹1000 Job)",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(12.dp))
            
            TransactionRow("Gross Amount", TaxCalculator.microsToRupees(breakdown.grossAmountMicros))
            TransactionRow("Commission (${(breakdown.commissionRate * 100).toInt()}%)", 
                -TaxCalculator.microsToRupees(breakdown.commissionMicros))
            TransactionRow("GST (18%)", -TaxCalculator.microsToRupees(breakdown.gstOnCommissionMicros))
            TransactionRow("TDS (1%)", -TaxCalculator.microsToRupees(breakdown.tdsDeductionMicros))
            
            HorizontalDivider(modifier = Modifier.padding(vertical = 8.dp))
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "Worker Payout",
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.primary
                )
                Text(
                    text = TaxCalculator.formatRupees(TaxCalculator.microsToRupees(breakdown.workerPayoutMicros)),
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.primary
                )
            }
        }
    }
}

@Composable
fun TransactionRow(label: String, amount: Double) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 2.dp),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(text = label, style = MaterialTheme.typography.bodyMedium)
        Text(
            text = if (amount >= 0) TaxCalculator.formatRupees(amount) 
                   else "-${TaxCalculator.formatRupees(-amount)}",
            style = MaterialTheme.typography.bodyMedium
        )
    }
}

@Composable
fun JobsScreen() {
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(
                text = "🔧",
                fontSize = 64.sp
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = "Jobs Coming Soon",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold
            )
            Text(
                text = "Worker job matching will appear here",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f)
            )
        }
    }
}

@Composable
fun ProfileScreen(userType: UserType, onLogout: () -> Unit = {}) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Spacer(modifier = Modifier.height(32.dp))
        
        Text(
            text = "👤",
            fontSize = 64.sp
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Text(
            text = if (userType == UserType.WORKER) "Worker Profile" else "Client Profile",
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        if (userType == UserType.WORKER) {
            // KYC Status Badge
            Card(
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.secondaryContainer
                )
            ) {
                Text(
                    text = "✅ KYC Verified",
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSecondaryContainer
                )
            }
        }
        
        Spacer(modifier = Modifier.weight(1f))
        
        // Logout Button
        OutlinedButton(
            onClick = onLogout,
            modifier = Modifier.fillMaxWidth()
        ) {
            Text("Logout")
        }
        
        Spacer(modifier = Modifier.height(16.dp))
    }
}

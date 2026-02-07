@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.animation.*
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.text.font.FontWeight
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.bero.data.auth.AuthState
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.TokenManager
import com.example.bero.data.models.UserType
import com.example.bero.data.models.Job
import com.example.bero.data.models.WorkerDisplayProfile
import com.example.bero.ui.auth.AuthViewModel
import com.example.bero.ui.auth.LoginScreen
import com.example.bero.ui.auth.OtpVerificationScreen
import com.example.bero.ui.auth.RoleSelectionScreen
import com.example.bero.ui.profile.CreateProfileScreen
import com.example.bero.ui.profile.EditProfileScreen
// KYC and Video Bio screens removed - now optional features
import com.example.bero.ui.theme.BeroTheme

// Navigation Screens
import com.example.bero.ui.onboarding.LanguageSelectionScreen
import com.example.bero.ui.job.JobsScreen as EnhancedJobsScreen
import com.example.bero.ui.job.JobDetailsScreen
import com.example.bero.ui.wallet.WalletScreen
import com.example.bero.ui.chat.ConversationsScreen
import com.example.bero.ui.notifications.NotificationsScreen
import com.example.bero.ui.categories.CategoriesScreen
import com.example.bero.ui.bookings.BookingsScreen
import com.example.bero.ui.profile.WorkerProfileScreen
import com.example.bero.ui.profile.ClientProfileScreen
import com.example.bero.ui.settings.SettingsScreen
import com.example.bero.ui.settings.NotificationPreferencesScreen
import com.example.bero.ui.settings.PrivacyPolicyScreen
import com.example.bero.ui.settings.TermsOfServiceScreen
import com.example.bero.ui.help.HelpSupportScreen
import com.example.bero.ui.search.SearchScreen
import com.example.bero.ui.earnings.EarningsAnalyticsScreen
import com.example.bero.ui.reviews.ReviewsManagementScreen
import com.example.bero.ui.skills.SkillManagementScreen
import com.example.bero.ui.worker.WorkerDetailsScreen
import com.example.bero.ui.booking.BookingFlowScreen
import com.example.bero.ui.payment.PaymentScreen
import com.example.bero.ui.rating.RatingFlowScreen
import com.example.bero.ui.settings.SettingsViewModel
import com.example.bero.ui.home.WorkerHomeScreen
import com.example.bero.ui.home.ClientHomeScreen
import com.example.bero.ui.job.CreateJobScreen

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            // Create SettingsViewModel at the top level
            val settingsViewModel: SettingsViewModel = viewModel()
            val settings by settingsViewModel.settings.collectAsState()
            
            BeroTheme(darkTheme = settings.isDarkMode) {
                BeroApp(settingsViewModel = settingsViewModel)
            }
        }
    }
}

// Simple Screen Navigation State
sealed class Screen {
    object Main : Screen()
    object LanguageSelection : Screen()
    object Settings : Screen()
    object Wallet : Screen()
    object NotificationPreferences : Screen()
    object HelpSupport : Screen()
    object Search : Screen()
    object EditProfile : Screen()
    
    // Worker Specific
    data class JobDetails(val job: Job) : Screen()
    object EarningsAnalytics : Screen()
    object ReviewsManagement : Screen()
    object SkillManagement : Screen()
    
    // Client Specific
    data class WorkerDetails(val workerId: String) : Screen()
    data class BookingFlow(val worker: WorkerDisplayProfile) : Screen()
    object Payment : Screen()
    data class RatingFlow(val worker: WorkerDisplayProfile) : Screen()
    object CreateJob : Screen()
    object PrivacyPolicy : Screen()
    object TermsOfService : Screen()
}

@Composable
fun BeroApp(settingsViewModel: SettingsViewModel = viewModel()) {
    val authViewModel: AuthViewModel = viewModel()
    val authState by authViewModel.authState.collectAsState()
    val uiState by authViewModel.uiState.collectAsState()
    val settings by settingsViewModel.settings.collectAsState()
    
    // Check if language has been selected (not first launch)
    val hasSelectedLanguage = settings.languageCode.isNotEmpty()
    var showLanguageSelection by remember { mutableStateOf(!hasSelectedLanguage || settings.languageCode == "en") }
    
    // For demo, always show language selection on first compose
    // In production, check SharedPreferences for "language_selected" flag
    var isFirstLaunch by remember { mutableStateOf(true) }
    
    if (isFirstLaunch && showLanguageSelection) {
        LanguageSelectionScreen(
            onLanguageSelected = { language ->
                settingsViewModel.setLanguage(language.code)
                showLanguageSelection = false
                isFirstLaunch = false
            }
        )
        return
    }

    when {
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
        authState is AuthState.NotAuthenticated || authState is AuthState.Authenticating -> {
            LoginScreen(
                onPhoneSubmit = { phone -> authViewModel.requestOtp(phone) },
                onTruecallerClick = { },
                isLoading = uiState.isLoading,
                error = uiState.error
            )
        }
        authState is AuthState.RequiresRoleSelection -> {
            RoleSelectionScreen(
                onRoleSelected = { role -> authViewModel.selectRole(role) },
                isLoading = uiState.isLoading,
                error = uiState.error
            )
        }
        authState is AuthState.RequiresProfileCreation -> {
            CreateProfileScreen(
                apiClient = authViewModel.apiClient,
                onProfileCreated = { authViewModel.completeProfileCreation() }
            )
        }
        // KYC and Video Bio are now optional - can be done later from profile settings
        // The RequiresKyc and RequiresVideoBio states are kept for future use if needed
        authState is AuthState.Authenticated -> {
            val user = (authState as AuthState.Authenticated).user
            MainAppScreen(
                userType = user.userType,
                userName = user.fullName ?: "Client",
                onLogout = { authViewModel.logout() },
                settingsViewModel = settingsViewModel,
                apiClient = authViewModel.apiClient
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MainAppScreen(
    userType: UserType,
    userName: String,
    onLogout: () -> Unit,
    settingsViewModel: SettingsViewModel = viewModel(),
    apiClient: BeroApiClient
) {
    var selectedTab by remember { mutableIntStateOf(0) }
    var currentScreen by remember { mutableStateOf<Screen>(Screen.Main) }
    
    // Helper to handle back press or back navigation
    val onBack = { currentScreen = Screen.Main }
    
    // Handle system back press for sub-screens (Settings, Notifications, etc.)
    BackHandler(enabled = currentScreen !is Screen.Main) {
        when (currentScreen) {
            is Screen.NotificationPreferences -> currentScreen = Screen.Settings
            is Screen.BookingFlow -> currentScreen = Screen.WorkerDetails((currentScreen as Screen.BookingFlow).worker.userId)
            else -> currentScreen = Screen.Main
        }
    }

    // Handle system back press on main tabs
    // If on a tab other than Home (0), go back to Home
    BackHandler(enabled = currentScreen is Screen.Main && selectedTab != 0) {
        selectedTab = 0
    }

    // Navigation Stack Logic
    when (val screen = currentScreen) {
        is Screen.Main -> {
            Scaffold(
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
                                icon = { Icon(Icons.Default.WorkOutline, contentDescription = "Jobs") },
                                label = { Text("Jobs") },
                                selected = selectedTab == 1,
                                onClick = { selectedTab = 1 }
                            )
                        } else {
                            NavigationBarItem(
                                icon = { Icon(Icons.Default.GridView, contentDescription = "Services") },
                                label = { Text("Services") },
                                selected = selectedTab == 1,
                                onClick = { selectedTab = 1 }
                            )
                        }
                        
                        NavigationBarItem(
                            icon = { 
                                Icon(Icons.Default.Chat, contentDescription = "Chat")
                            },
                            label = { Text("Chat") },
                            selected = selectedTab == 2,
                            onClick = { selectedTab = 2 }
                        )
                        
                        NavigationBarItem(
                            icon = { Icon(Icons.Default.ListAlt, contentDescription = "Bookings") },
                            label = { Text("Bookings") },
                            selected = selectedTab == 3,
                            onClick = { selectedTab = 3 }
                        )
                        
                        NavigationBarItem(
                            icon = { Icon(Icons.Default.Person, contentDescription = "Profile") },
                            label = { Text("Profile") },
                            selected = selectedTab == 4,
                            onClick = { selectedTab = 4 }
                        )
                    }
                }
            ) { innerPadding ->
                Box(modifier = Modifier.padding(innerPadding)) {
                    when (selectedTab) {
                        0 -> {
                            if (userType == UserType.WORKER) {
                                WorkerHomeScreen(
                                    onJobClick = { job -> currentScreen = Screen.JobDetails(job) },
                                    onViewAllJobsClick = { selectedTab = 1 }
                                )
                            } else {
                                ClientHomeScreen(
                                    onCategoryClick = { currentScreen = Screen.Search },
                                    onWorkerClick = { workerId -> currentScreen = Screen.WorkerDetails(workerId) },
                                    onViewAllCategoriesClick = { selectedTab = 1 },
                                    onPostJobClick = { currentScreen = Screen.CreateJob }
                                )
                            }
                        }
                        1 -> if (userType == UserType.WORKER) {
                            EnhancedJobsScreen(
                                onJobClick = { job -> currentScreen = Screen.JobDetails(job) }
                            )
                        } else {
                            // On Category Click from CategoriesScreen -> Search/List (Simplified to Search for now)
                            CategoriesScreen(
                                onCategoryClick = { currentScreen = Screen.Search }
                            )
                        }
                        2 -> ConversationsScreen()
                        3 -> BookingsScreen(
                            isWorker = userType == UserType.WORKER,
                            onWorkerClick = { workerId -> currentScreen = Screen.WorkerDetails(workerId) }
                        )
                        4 -> if (userType == UserType.WORKER) {
                            WorkerProfileScreen(
                                apiClient = apiClient,
                                onLogout = onLogout,
                                onEditProfileClick = { currentScreen = Screen.EditProfile },
                                onSettingsClick = { currentScreen = Screen.Settings },
                                onWalletClick = { currentScreen = Screen.Wallet },
                                onEarningsClick = { currentScreen = Screen.EarningsAnalytics },
                                onReviewsClick = { currentScreen = Screen.ReviewsManagement },
                                onSkillsClick = { currentScreen = Screen.SkillManagement },
                                onHelpClick = { currentScreen = Screen.HelpSupport }
                            )
                        } else {
                            ClientProfileScreen(
                                apiClient = apiClient,
                                onLogout = onLogout,
                                onEditProfileClick = { currentScreen = Screen.EditProfile },
                                onSettingsClick = { currentScreen = Screen.Settings },
                                onPaymentMethodsClick = { currentScreen = Screen.Payment },
                                onHelpClick = { currentScreen = Screen.HelpSupport }
                            )
                        }
                    }
                }
            }
        }
        
        // Full Screen Overlays
        is Screen.JobDetails -> {
           JobDetailsScreen(
               job = screen.job,
               onBackClick = onBack,
               onChatClick = { /* Go to Chat */ }
           )
        }
        
        is Screen.Settings -> {
            SettingsScreen(
                onBackClick = onBack,
                onNotificationsClick = { currentScreen = Screen.NotificationPreferences },
                onHelpClick = { currentScreen = Screen.HelpSupport },
                onPrivacyPolicyClick = { currentScreen = Screen.PrivacyPolicy },
                onTermsOfServiceClick = { currentScreen = Screen.TermsOfService },
                settingsViewModel = settingsViewModel
            )
        }
        
        is Screen.Wallet -> {
            Scaffold(topBar = {
                TopAppBar(title = { Text("Wallet") }, navigationIcon = {
                    IconButton(onClick = onBack) { Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back") }
                })
            }) { p -> Box(Modifier.padding(p)) { WalletScreen() } }
        }
        
        is Screen.NotificationPreferences -> {
            NotificationPreferencesScreen(onBackClick = { currentScreen = Screen.Settings })
        }
        
        is Screen.PrivacyPolicy -> {
            PrivacyPolicyScreen(onBackClick = { currentScreen = Screen.Settings })
        }
        
        is Screen.TermsOfService -> {
            TermsOfServiceScreen(onBackClick = { currentScreen = Screen.Settings })
        }
        
        is Screen.HelpSupport -> {
            HelpSupportScreen(onBackClick = onBack)
        }
        
        is Screen.Search -> {
            SearchScreen(
                onBackClick = onBack,
                onCategoryClick = { /* Handle category selection from search */ }
            )
        }

        is Screen.EditProfile -> {
            EditProfileScreen(
                apiClient = apiClient,
                onBackClick = onBack
            )
        }
        
        is Screen.EarningsAnalytics -> {
            EarningsAnalyticsScreen(onBackClick = onBack)
        }
        
        is Screen.ReviewsManagement -> {
            ReviewsManagementScreen(onBackClick = onBack)
        }
        
        is Screen.SkillManagement -> {
            SkillManagementScreen(onBackClick = onBack)
        }
        
        is Screen.Payment -> {
            PaymentScreen(onBackClick = onBack)
        }
        
        // Client Flow Handling (Simplified trigger)
        is Screen.WorkerDetails -> {
            WorkerDetailsScreen(
                workerId = screen.workerId,
                apiClient = apiClient,
                onBackClick = onBack,
                onBookClick = { worker -> currentScreen = Screen.BookingFlow(worker) },
                onChatClick = { /* Navigate to chat */ }
            )
        }
        
        is Screen.BookingFlow -> {
            BookingFlowScreen(
                worker = screen.worker,
                onBackClick = { currentScreen = Screen.WorkerDetails(screen.worker.userId) },
                onBookingComplete = { 
                    // Go to rating or home
                    onBack()
                }
            )
        }
        
        is Screen.CreateJob -> {
            CreateJobScreen(
                onBackClick = onBack,
                onJobCreated = {
                    // Navigate back to home after job is created
                    onBack()
                }
            )
        }
         
         // Other screens...
         else -> {
             // Fallback
             Box(Modifier.fillMaxSize(), Alignment.Center) {
                 Text("Screen not implemented yet: $screen")
                 Button(onClick = onBack) { Text("Go Back") }
             }
         }
    }
}


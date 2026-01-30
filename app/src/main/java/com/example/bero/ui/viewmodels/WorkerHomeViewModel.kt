package com.example.bero.ui.viewmodels

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.bero.data.models.EarningsSummary
import com.example.bero.data.models.DailyEarning
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * ViewModel for worker home screen and earnings
 */
class WorkerHomeViewModel : ViewModel() {
    
    // Worker stats
    private val _isOnline = MutableStateFlow(false)
    val isOnline: StateFlow<Boolean> = _isOnline.asStateFlow()
    
    private val _streakCount = MutableStateFlow(0)
    val streakCount: StateFlow<Int> = _streakCount.asStateFlow()
    
    private val _todayEarnings = MutableStateFlow(0.0)
    val todayEarnings: StateFlow<Double> = _todayEarnings.asStateFlow()
    
    private val _weeklyEarnings = MutableStateFlow(0.0)
    val weeklyEarnings: StateFlow<Double> = _weeklyEarnings.asStateFlow()
    
    private val _jobsCompletedToday = MutableStateFlow(0)
    val jobsCompletedToday: StateFlow<Int> = _jobsCompletedToday.asStateFlow()
    
    private val _rating = MutableStateFlow(0.0)
    val rating: StateFlow<Double> = _rating.asStateFlow()
    
    private val _earningsSummary = MutableStateFlow<EarningsSummary?>(null)
    val earningsSummary: StateFlow<EarningsSummary?> = _earningsSummary.asStateFlow()
    
    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()
    
    /**
     * Toggle online/offline status
     */
    fun toggleOnlineStatus(online: Boolean) {
        viewModelScope.launch {
            _isOnline.value = online
            // TODO: Call API to update worker online status
        }
    }
    
    /**
     * Load worker dashboard data
     */
    fun loadDashboard() {
        viewModelScope.launch {
            _isLoading.value = true
            
            // TODO: Replace with actual API calls
            // For now, using mock data
            _todayEarnings.value = 1250.0
            _weeklyEarnings.value = 8540.0
            _jobsCompletedToday.value = 3
            _streakCount.value = 7
            _rating.value = 4.8
            _isOnline.value = true
            
            _isLoading.value = false
        }
    }
    
    /**
     * Load earnings summary for analytics
     */
    fun loadEarningsSummary(period: String = "WEEK") {
        viewModelScope.launch {
            _isLoading.value = true
            
            // TODO: Replace with actual API call
            // Mock data for now
            _earningsSummary.value = EarningsSummary(
                period = period,
                totalEarnings = 8540.0,
                jobsCompleted = 12,
                averagePerJob = 712.0,
                commissionPaid = 1281.0,
                tdsPaid = 85.4,
                gstPaid = 0.0,
                netEarnings = 7173.6,
                dailyEarnings = listOf(
                    DailyEarning("Mon", 1200.0, 2),
                    DailyEarning("Tue", 850.0, 1),
                    DailyEarning("Wed", 1500.0, 2),
                    DailyEarning("Thu", 1100.0, 2),
                    DailyEarning("Fri", 1890.0, 3),
                    DailyEarning("Sat", 1250.0, 2),
                    DailyEarning("Sun", 750.0, 1)
                )
            )
            
            _isLoading.value = false
        }
    }
}

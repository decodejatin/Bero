package com.example.bero.ui.job

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.bero.data.models.Job
import com.example.bero.data.models.ServiceCategory
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.TokenManager
import com.example.bero.data.repository.JobRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * ViewModel for creating new jobs (client side)
 */
class CreateJobViewModel(application: Application) : AndroidViewModel(application) {
    
    private val tokenManager = TokenManager(application)
    private val apiClient = BeroApiClient(tokenManager)
    private val jobRepository = JobRepository(apiClient)
    
    private val _uiState = MutableStateFlow(CreateJobUiState())
    val uiState: StateFlow<CreateJobUiState> = _uiState.asStateFlow()
    
    fun createJob(
        title: String,
        description: String,
        category: ServiceCategory,
        clientName: String,
        address: String,
        locality: String,
        city: String = "Delhi",
        pincode: String,
        paymentAmount: Double,
        scheduledDate: String,
        scheduledTimeSlot: String,
        isUrgent: Boolean = false
    ) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)
            
            // Format date to RFC3339 (Backend requirement)
            val formattedDate = try {
                if (scheduledDate.equals("Today", ignoreCase = true)) {
                    java.time.format.DateTimeFormatter.ISO_INSTANT.format(java.time.Instant.now())
                } else if (scheduledDate.equals("Tomorrow", ignoreCase = true)) {
                    java.time.format.DateTimeFormatter.ISO_INSTANT.format(java.time.Instant.now().plus(java.time.Duration.ofDays(1)))
                } else {
                     // Try basic parsing or fallback to now
                    // For safety in this dev iteration, we'll default to "Now" if it's not a standard parsable string
                    // In a real app, we'd use a DatePicker
                     java.time.format.DateTimeFormatter.ISO_INSTANT.format(java.time.Instant.now().plus(java.time.Duration.ofDays(1)))
                }
            } catch (e: Exception) {
                 // Fallback
                 java.time.format.DateTimeFormatter.ISO_INSTANT.format(java.time.Instant.now())
            }

            jobRepository.createJob(
                title = title,
                description = description,
                category = category,
                clientName = clientName,
                address = address,
                locality = locality,
                city = city,
                pincode = pincode,
                paymentAmount = paymentAmount,
                scheduledDate = formattedDate,
                scheduledTimeSlot = scheduledTimeSlot,
                isUrgent = isUrgent
            ).fold(
                onSuccess = { job ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        createdJob = job,
                        isSuccess = true
                    )
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = error.message ?: "Failed to create job"
                    )
                }
            )
        }
    }
    
    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }
    
    fun resetState() {
        _uiState.value = CreateJobUiState()
    }
}

/**
 * UI state for job creation
 */
data class CreateJobUiState(
    val isLoading: Boolean = false,
    val error: String? = null,
    val isSuccess: Boolean = false,
    val createdJob: Job? = null
)

package com.example.bero.ui.jobs

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.bero.data.models.Job
import com.example.bero.data.models.JobStatus
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.TokenManager
import com.example.bero.data.repository.JobRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * ViewModel for Jobs screens
 */
class JobsViewModel(application: Application) : AndroidViewModel(application) {
    
    private val tokenManager = TokenManager(application)
    private val apiClient = BeroApiClient(tokenManager)
    private val jobRepository = JobRepository(apiClient)
    
    private val _uiState = MutableStateFlow(JobsUiState())
    val uiState: StateFlow<JobsUiState> = _uiState.asStateFlow()
    
    private val _availableJobs = MutableStateFlow<List<Job>>(emptyList())
    val availableJobs: StateFlow<List<Job>> = _availableJobs.asStateFlow()
    
    private val _myJobs = MutableStateFlow<List<Job>>(emptyList())
    val myJobs: StateFlow<List<Job>> = _myJobs.asStateFlow()
    
    init {
        loadJobs()
    }
    
    fun loadJobs() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            // Load available jobs
            jobRepository.getAvailableJobs().fold(
                onSuccess = { jobs ->
                    _availableJobs.value = jobs
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        error = error.message ?: "Failed to load jobs"
                    )
                }
            )
            
            // Load my jobs
            jobRepository.getMyJobs().fold(
                onSuccess = { jobs ->
                    _myJobs.value = jobs
                },
                onFailure = { /* Ignore */ }
            )
            
            _uiState.value = _uiState.value.copy(isLoading = false)
        }
    }
    
    fun acceptJob(jobId: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isAccepting = true)
            
            jobRepository.acceptJob(jobId).fold(
                onSuccess = {
                    _uiState.value = _uiState.value.copy(
                        isAccepting = false,
                        acceptedJobId = jobId
                    )
                    // Refresh jobs
                    loadJobs()
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        isAccepting = false,
                        error = error.message ?: "Failed to accept job"
                    )
                }
            )
        }
    }
    
    fun startJob(jobId: String) {
        viewModelScope.launch {
            jobRepository.startJob(jobId).fold(
                onSuccess = { loadJobs() },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        error = error.message ?: "Failed to start job"
                    )
                }
            )
        }
    }
    
    fun completeJob(jobId: String, notes: String? = null) {
        viewModelScope.launch {
            jobRepository.completeJob(jobId, workerNotes = notes).fold(
                onSuccess = { 
                    _uiState.value = _uiState.value.copy(completedJobId = jobId)
                    loadJobs() 
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        error = error.message ?: "Failed to complete job"
                    )
                }
            )
        }
    }
    
    fun cancelJob(jobId: String, reason: String = "Cancelled by worker") {
        viewModelScope.launch {
            jobRepository.cancelJob(jobId).fold(
                onSuccess = { loadJobs() },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        error = error.message ?: "Failed to cancel job"
                    )
                }
            )
        }
    }
    
    fun confirmJobCompletion(jobId: String) {
        viewModelScope.launch {
            jobRepository.confirmJobCompletion(jobId).fold(
                onSuccess = { 
                    _uiState.value = _uiState.value.copy(completedJobId = jobId)
                    loadJobs() 
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        error = error.message ?: "Failed to confirm job completion"
                    )
                }
            )
        }
    }
    
    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }
    
    fun setSelectedTab(tab: Int) {
        _uiState.value = _uiState.value.copy(selectedTab = tab)
    }
}

/**
 * UI state for jobs screens
 */
data class JobsUiState(
    val isLoading: Boolean = false,
    val isAccepting: Boolean = false,
    val error: String? = null,
    val selectedTab: Int = 0,
    val acceptedJobId: String? = null,
    val completedJobId: String? = null
)

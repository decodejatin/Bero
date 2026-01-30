package com.example.bero.ui.jobs

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.bero.domain.models.Job
import com.bero.domain.models.JobStatus
import com.bero.jobs.JobRepository
import com.bero.jobs.JobRepositoryImpl
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * ViewModel for Jobs screens
 */
class JobsViewModel : ViewModel() {
    
    private val jobRepository: JobRepository = JobRepositoryImpl()
    
    private val _uiState = MutableStateFlow(JobsUiState())
    val uiState: StateFlow<JobsUiState> = _uiState.asStateFlow()
    
    private val _availableJobs = MutableStateFlow<List<Job>>(emptyList())
    val availableJobs: StateFlow<List<Job>> = _availableJobs.asStateFlow()
    
    private val _myJobs = MutableStateFlow<List<Job>>(emptyList())
    val myJobs: StateFlow<List<Job>> = _myJobs.asStateFlow()
    
    private val workerId = "worker-001" // TODO: Get from auth
    
    init {
        loadJobs()
    }
    
    fun loadJobs() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            // Load available jobs
            jobRepository.getAvailableJobs(workerId).fold(
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
            jobRepository.getMyJobs(workerId).fold(
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
            
            jobRepository.acceptJob(jobId, workerId).fold(
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
            jobRepository.startJob(jobId, workerId).fold(
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
            jobRepository.completeJob(jobId, workerId, notes).fold(
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
    
    fun cancelJob(jobId: String, reason: String) {
        viewModelScope.launch {
            jobRepository.cancelJobAcceptance(jobId, workerId, reason).fold(
                onSuccess = { loadJobs() },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        error = error.message ?: "Failed to cancel job"
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

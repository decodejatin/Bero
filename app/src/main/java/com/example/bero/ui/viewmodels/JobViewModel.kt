package com.example.bero.ui.viewmodels

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.bero.data.models.Job
import com.example.bero.data.models.ServiceCategory
import com.example.bero.data.repository.JobRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * ViewModel for job-related screens
 */
class JobViewModel(private val jobRepository: JobRepository) : ViewModel() {
    
    // UI States
    private val _availableJobs = MutableStateFlow<List<Job>>(emptyList())
    val availableJobs: StateFlow<List<Job>> = _availableJobs.asStateFlow()
    
    private val _myJobs = MutableStateFlow<List<Job>>(emptyList())
    val myJobs: StateFlow<List<Job>> = _myJobs.asStateFlow()
    
    private val _selectedJob = MutableStateFlow<Job?>(null)
    val selectedJob: StateFlow<Job?> = _selectedJob.asStateFlow()
    
    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()
    
    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error.asStateFlow()
    
    private val _actionSuccess = MutableStateFlow<String?>(null)
    val actionSuccess: StateFlow<String?> = _actionSuccess.asStateFlow()
    
    /**
     * Load available jobs for workers
     */
    fun loadAvailableJobs(
        locality: String? = null,
        category: ServiceCategory? = null
    ) {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            
            jobRepository.getAvailableJobs(locality, category)
                .onSuccess { jobs ->
                    _availableJobs.value = jobs
                }
                .onFailure { e ->
                    _error.value = e.message ?: "Failed to load jobs"
                }
            
            _isLoading.value = false
        }
    }
    
    /**
     * Load user's own jobs (assigned for worker, created for client)
     */
    fun loadMyJobs(status: String? = null) {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            
            jobRepository.getMyJobs(status)
                .onSuccess { jobs ->
                    _myJobs.value = jobs
                }
                .onFailure { e ->
                    _error.value = e.message ?: "Failed to load your jobs"
                }
            
            _isLoading.value = false
        }
    }
    
    /**
     * Load job details
     */
    fun loadJobDetails(jobId: String) {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            
            jobRepository.getJobDetails(jobId)
                .onSuccess { job ->
                    _selectedJob.value = job
                }
                .onFailure { e ->
                    _error.value = e.message ?: "Failed to load job details"
                }
            
            _isLoading.value = false
        }
    }
    
    /**
     * Accept a job (worker action)
     */
    fun acceptJob(jobId: String, estimatedArrival: Int = 30) {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            
            jobRepository.acceptJob(jobId, estimatedArrival)
                .onSuccess {
                    _actionSuccess.value = "Job accepted successfully!"
                    loadMyJobs() // Refresh my jobs
                    loadAvailableJobs() // Refresh available jobs
                }
                .onFailure { e ->
                    _error.value = e.message ?: "Failed to accept job"
                }
            
            _isLoading.value = false
        }
    }
    
    /**
     * Start a job (worker action)
     */
    fun startJob(jobId: String) {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            
            jobRepository.startJob(jobId)
                .onSuccess {
                    _actionSuccess.value = "Job started!"
                    loadJobDetails(jobId) // Refresh job details
                    loadMyJobs() // Refresh my jobs
                }
                .onFailure { e ->
                    _error.value = e.message ?: "Failed to start job"
                }
            
            _isLoading.value = false
        }
    }
    
    /**
     * Complete a job (worker action)
     */
    fun completeJob(
        jobId: String,
        notes: String? = null,
        photoUrls: List<String> = emptyList()
    ) {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            
            jobRepository.completeJob(jobId, notes, photoUrls)
                .onSuccess {
                    _actionSuccess.value = "Job completed! Great work!"
                    loadMyJobs() // Refresh my jobs
                }
                .onFailure { e ->
                    _error.value = e.message ?: "Failed to complete job"
                }
            
            _isLoading.value = false
        }
    }
    
    /**
     * Cancel a job (client action)
     */
    fun cancelJob(jobId: String) {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            
            jobRepository.cancelJob(jobId)
                .onSuccess {
                    _actionSuccess.value = "Job cancelled"
                    loadMyJobs() // Refresh my jobs
                }
                .onFailure { e ->
                    _error.value = e.message ?: "Failed to cancel job"
                }
            
            _isLoading.value = false
        }
    }
    
    /**
     * Create a new job (client action)
     */
    fun createJob(
        title: String,
        description: String,
        category: ServiceCategory,
        clientName: String,
        address: String,
        locality: String,
        pincode: String,
        paymentAmount: Double,
        scheduledDate: String,
        scheduledTimeSlot: String,
        isUrgent: Boolean = false
    ) {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            
            jobRepository.createJob(
                title = title,
                description = description,
                category = category,
                clientName = clientName,
                address = address,
                locality = locality,
                pincode = pincode,
                paymentAmount = paymentAmount,
                scheduledDate = scheduledDate,
                scheduledTimeSlot = scheduledTimeSlot,
                isUrgent = isUrgent
            )
                .onSuccess { job ->
                    _actionSuccess.value = "Job posted successfully!"
                    loadMyJobs() // Refresh my jobs
                }
                .onFailure { e ->
                    _error.value = e.message ?: "Failed to create job"
                }
            
            _isLoading.value = false
        }
    }
    
    /**
     * Clear error message
     */
    fun clearError() {
        _error.value = null
    }
    
    /**
     * Clear success message
     */
    fun clearSuccess() {
        _actionSuccess.value = null
    }
}

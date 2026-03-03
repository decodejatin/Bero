package com.example.bero.data.repository

import com.example.bero.data.models.Job
import com.example.bero.data.network.ApiMapper.toJob
import com.example.bero.data.network.ApiMapper.toJobs
import com.example.bero.data.network.ApiMapper.toApiCategory
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.CompleteJobRequest
import com.example.bero.data.network.CreateJobRequest
import com.example.bero.data.models.ServiceCategory

/**
 * Repository for job operations - connects UI to backend API
 */
class JobRepository(private val apiClient: BeroApiClient) {
    
    /**
     * Get available jobs for workers
     */
    suspend fun getAvailableJobs(
        locality: String? = null,
        category: ServiceCategory? = null,
        limit: Int = 20,
        offset: Int = 0
    ): Result<List<Job>> {
        return apiClient.getAvailableJobs(
            locality = locality,
            category = category?.toApiCategory(),
            limit = limit,
            offset = offset
        ).map { it.toJobs() }
    }
    
    /**
     * Get jobs for the current user (worker's assigned jobs or client's posted jobs)
     */
    suspend fun getMyJobs(
        status: String? = null,
        limit: Int = 20,
        offset: Int = 0
    ): Result<List<Job>> {
        return apiClient.getMyJobs(status, limit, offset).map { it.toJobs() }
    }
    
    /**
     * Get job details by ID
     */
    suspend fun getJobDetails(jobId: String): Result<Job> {
        return apiClient.getJobDetails(jobId).map { it.toJob() }
    }
    
    /**
     * Create a new job (client only)
     */
    suspend fun createJob(
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
        isUrgent: Boolean = false,
        requiredSkills: List<String> = emptyList()
    ): Result<Job> {
        val request = CreateJobRequest(
            title = title,
            description = description,
            category = category.toApiCategory(),
            client_name = clientName,
            address = address,
            locality = locality,
            city = city,
            pincode = pincode,
            payment_amount_rupees = paymentAmount,
            scheduled_date = scheduledDate,
            scheduled_time_slot = scheduledTimeSlot,
            is_urgent = isUrgent,
            required_skills = requiredSkills
        )
        return apiClient.createJob(request).map { it.toJob() }
    }
    
    /**
     * Accept a job (worker only)
     */
    suspend fun acceptJob(jobId: String, estimatedArrivalMinutes: Int = 30): Result<Unit> {
        return apiClient.acceptJob(jobId, estimatedArrivalMinutes).map { }
    }
    
    /**
     * Start working on a job
     */
    suspend fun startJob(jobId: String): Result<Unit> {
        return apiClient.startJob(jobId).map { }
    }
    
    /**
     * Complete a job
     */
    suspend fun completeJob(
        jobId: String,
        workerNotes: String? = null,
        photoProofUrls: List<String> = emptyList()
    ): Result<Unit> {
        val request = CompleteJobRequest(
            worker_notes = workerNotes,
            photo_proof_urls = photoProofUrls
        )
        return apiClient.completeJob(jobId, request).map { }
    }
    
    /**
     * Cancel a job
     */
    suspend fun cancelJob(jobId: String): Result<Unit> {
        return apiClient.cancelJob(jobId).map { }
    }
    
    /**
     * Client confirms job completion (dual-sided completion flow)
     * Calls POST /jobs/:id/confirm-by-client
     */
    suspend fun confirmJobCompletion(jobId: String): Result<Unit> {
        return try {
            apiClient.clientConfirmComplete(jobId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    /**
     * Worker marks job complete (dual-sided completion flow)
     * Calls POST /jobs/:id/complete-by-worker
     */
    suspend fun workerMarkComplete(jobId: String): Result<Unit> {
        return try {
            apiClient.workerMarkComplete(jobId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    /**
     * Submit mutual rating after job is fully completed
     * Calls POST /jobs/:id/rate
     */
    suspend fun submitMutualRating(jobId: String, rating: Int, review: String = ""): Result<Unit> {
        return try {
            apiClient.submitMutualRating(jobId, rating, review)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

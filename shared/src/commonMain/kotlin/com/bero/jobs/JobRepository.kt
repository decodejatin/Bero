package com.bero.jobs

import com.bero.domain.models.Job
import com.bero.domain.models.JobAcceptance
import com.bero.domain.models.JobCompletion
import com.bero.domain.models.JobStatus
import com.bero.domain.models.ServiceCategory
import com.bero.auth.currentTimeMillis

/**
 * Repository for job operations
 */
interface JobRepository {
    /**
     * Get available jobs for a worker based on their skills and location
     */
    suspend fun getAvailableJobs(
        workerId: String,
        latitude: Double? = null,
        longitude: Double? = null,
        radiusKm: Double = 10.0
    ): Result<List<Job>>
    
    /**
     * Get jobs assigned to a worker
     */
    suspend fun getMyJobs(workerId: String): Result<List<Job>>
    
    /**
     * Accept a job
     */
    suspend fun acceptJob(jobId: String, workerId: String): Result<JobAcceptance>
    
    /**
     * Mark job as in progress (worker has arrived)
     */
    suspend fun startJob(jobId: String, workerId: String): Result<Job>
    
    /**
     * Complete a job
     */
    suspend fun completeJob(
        jobId: String,
        workerId: String,
        notes: String? = null,
        photoProofUrls: List<String> = emptyList()
    ): Result<JobCompletion>
    
    /**
     * Cancel job acceptance
     */
    suspend fun cancelJobAcceptance(jobId: String, workerId: String, reason: String): Result<Unit>
}

/**
 * Mock implementation for development
 */
class JobRepositoryImpl : JobRepository {
    
    // Mock jobs database
    private val mockJobs = mutableListOf(
        Job(
            id = "job-001",
            title = "Leaky Tap Repair",
            description = "Kitchen tap is leaking continuously. Need urgent fix.",
            category = ServiceCategory.PLUMBING,
            status = JobStatus.OPEN,
            clientName = "Rahul Sharma",
            address = "B-42, Sector 15",
            locality = "Sector 15",
            city = "Noida",
            pincode = "201301",
            paymentAmountRupees = 350.0,
            scheduledDate = currentTimeMillis() + (2 * 60 * 60 * 1000), // 2 hours from now
            scheduledTimeSlot = "10:00 AM - 12:00 PM",
            isUrgent = true,
            requiredSkills = listOf("Plumbing", "Tap Repair")
        ),
        Job(
            id = "job-002",
            title = "AC Not Cooling",
            description = "1.5 ton split AC not cooling properly. Might need gas refill.",
            category = ServiceCategory.AC_REPAIR,
            status = JobStatus.OPEN,
            clientName = "Priya Verma",
            address = "Flat 301, Green Valley Apartments",
            locality = "Indirapuram",
            city = "Ghaziabad",
            pincode = "201014",
            paymentAmountRupees = 800.0,
            scheduledDate = currentTimeMillis() + (24 * 60 * 60 * 1000), // Tomorrow
            scheduledTimeSlot = "2:00 PM - 4:00 PM",
            isUrgent = false,
            requiredSkills = listOf("AC Repair", "HVAC"),
            estimatedDurationMinutes = 90
        ),
        Job(
            id = "job-003",
            title = "Electrical Socket Replacement",
            description = "Need to replace 4 damaged electrical sockets in bedroom.",
            category = ServiceCategory.ELECTRICAL,
            status = JobStatus.OPEN,
            clientName = "Amit Kumar",
            address = "House No. 23, Block C",
            locality = "Vasant Kunj",
            city = "Delhi",
            pincode = "110070",
            paymentAmountRupees = 500.0,
            scheduledDate = currentTimeMillis() + (3 * 60 * 60 * 1000), // 3 hours from now
            scheduledTimeSlot = "11:00 AM - 1:00 PM",
            isUrgent = false,
            requiredSkills = listOf("Electrical", "Wiring")
        ),
        Job(
            id = "job-004",
            title = "Full House Deep Cleaning",
            description = "3BHK flat needs deep cleaning including kitchen and bathrooms.",
            category = ServiceCategory.CLEANING,
            status = JobStatus.OPEN,
            clientName = "Neha Gupta",
            address = "Tower A, DLF Phase 2",
            locality = "DLF Phase 2",
            city = "Gurgaon",
            pincode = "122002",
            paymentAmountRupees = 2500.0,
            scheduledDate = currentTimeMillis() + (48 * 60 * 60 * 1000), // Day after tomorrow
            scheduledTimeSlot = "9:00 AM - 1:00 PM",
            isUrgent = false,
            requiredSkills = listOf("Cleaning", "Deep Clean"),
            estimatedDurationMinutes = 240
        )
    )
    
    private val workerJobs = mutableMapOf<String, MutableList<String>>() // workerId -> list of jobIds
    
    override suspend fun getAvailableJobs(
        workerId: String,
        latitude: Double?,
        longitude: Double?,
        radiusKm: Double
    ): Result<List<Job>> {
        // Return open jobs (in real impl, filter by location and skills)
        val availableJobs = mockJobs.filter { it.status == JobStatus.OPEN }
        return Result.success(availableJobs)
    }
    
    override suspend fun getMyJobs(workerId: String): Result<List<Job>> {
        val jobIds = workerJobs[workerId] ?: emptyList()
        val jobs = mockJobs.filter { it.id in jobIds }
        return Result.success(jobs)
    }
    
    override suspend fun acceptJob(jobId: String, workerId: String): Result<JobAcceptance> {
        val job = mockJobs.find { it.id == jobId }
            ?: return Result.failure(Exception("Job not found"))
        
        if (job.status != JobStatus.OPEN) {
            return Result.failure(Exception("Job is no longer available"))
        }
        
        // Update job status
        val index = mockJobs.indexOfFirst { it.id == jobId }
        mockJobs[index] = job.copy(status = JobStatus.ASSIGNED)
        
        // Track assignment
        workerJobs.getOrPut(workerId) { mutableListOf() }.add(jobId)
        
        return Result.success(
            JobAcceptance(
                jobId = jobId,
                workerId = workerId,
                acceptedAt = currentTimeMillis(),
                estimatedArrivalMinutes = 30
            )
        )
    }
    
    override suspend fun startJob(jobId: String, workerId: String): Result<Job> {
        val job = mockJobs.find { it.id == jobId }
            ?: return Result.failure(Exception("Job not found"))
        
        val index = mockJobs.indexOfFirst { it.id == jobId }
        val updatedJob = job.copy(status = JobStatus.IN_PROGRESS)
        mockJobs[index] = updatedJob
        
        return Result.success(updatedJob)
    }
    
    override suspend fun completeJob(
        jobId: String,
        workerId: String,
        notes: String?,
        photoProofUrls: List<String>
    ): Result<JobCompletion> {
        val job = mockJobs.find { it.id == jobId }
            ?: return Result.failure(Exception("Job not found"))
        
        val index = mockJobs.indexOfFirst { it.id == jobId }
        mockJobs[index] = job.copy(status = JobStatus.COMPLETED)
        
        return Result.success(
            JobCompletion(
                jobId = jobId,
                workerId = workerId,
                completedAt = currentTimeMillis(),
                workerNotes = notes,
                photoProofUrls = photoProofUrls
            )
        )
    }
    
    override suspend fun cancelJobAcceptance(
        jobId: String,
        workerId: String,
        reason: String
    ): Result<Unit> {
        val job = mockJobs.find { it.id == jobId }
            ?: return Result.failure(Exception("Job not found"))
        
        val index = mockJobs.indexOfFirst { it.id == jobId }
        mockJobs[index] = job.copy(status = JobStatus.OPEN)
        
        workerJobs[workerId]?.remove(jobId)
        
        return Result.success(Unit)
    }
}

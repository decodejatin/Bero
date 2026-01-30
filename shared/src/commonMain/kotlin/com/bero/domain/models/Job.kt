package com.bero.domain.models

import kotlinx.serialization.Serializable

/**
 * Job status in the worker flow
 */
@Serializable
enum class JobStatus {
    OPEN,           // Available for workers to accept
    ASSIGNED,       // Worker has been assigned
    IN_PROGRESS,    // Worker is actively working
    COMPLETED,      // Job finished successfully
    CANCELLED,      // Job was cancelled
    DISPUTED        // There's a dispute to resolve
}

/**
 * Service categories
 */
@Serializable
enum class ServiceCategory {
    PLUMBING,
    ELECTRICAL,
    CARPENTRY,
    PAINTING,
    CLEANING,
    AC_REPAIR,
    APPLIANCE_REPAIR,
    PEST_CONTROL,
    GARDENING,
    OTHER
}

/**
 * Job listing model
 * Represents a job available for workers
 */
@Serializable
data class Job(
    val id: String,
    val title: String,
    val description: String,
    val category: ServiceCategory,
    val status: JobStatus = JobStatus.OPEN,
    val clientName: String,
    val clientPhone: String? = null, // Only visible after acceptance
    val address: String,
    val locality: String,
    val city: String = "Delhi",
    val pincode: String,
    val latitude: Double? = null,
    val longitude: Double? = null,
    val estimatedDurationMinutes: Int = 60,
    val paymentAmountRupees: Double,
    val scheduledDate: Long, // Unix timestamp
    val scheduledTimeSlot: String, // e.g., "10:00 AM - 12:00 PM"
    val isUrgent: Boolean = false,
    val requiredSkills: List<String> = emptyList(),
    val createdAt: Long = 0L
)

/**
 * Job acceptance by worker
 */
@Serializable
data class JobAcceptance(
    val jobId: String,
    val workerId: String,
    val acceptedAt: Long,
    val estimatedArrivalMinutes: Int = 30
)

/**
 * Job completion record
 */
@Serializable
data class JobCompletion(
    val jobId: String,
    val workerId: String,
    val completedAt: Long,
    val clientRating: Int? = null, // 1-5
    val clientReview: String? = null,
    val workerNotes: String? = null,
    val photoProofUrls: List<String> = emptyList()
)

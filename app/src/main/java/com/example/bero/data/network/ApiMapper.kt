package com.example.bero.data.network

import com.example.bero.data.models.*

/**
 * Mapper functions to convert API DTOs to app models
 */
object ApiMapper {
    
    fun JobDto.toJob(): Job {
        return Job(
            id = this.id,
            title = this.title,
            description = this.description,
            category = mapCategory(this.category),
            location = "${this.address}, ${this.locality}",
            landmark = null,
            amountRupees = this.payment_amount_rupees,
            status = mapJobStatus(this.status),
            clientId = this.client_id,
            clientName = this.client_name,
            clientRating = 4.5, // Default, would come from user endpoint
            workerId = this.assigned_worker_id,
            workerName = null,
            postedAt = System.currentTimeMillis(), // TODO: Parse created_at
            scheduledDate = this.scheduled_date,
            scheduledTime = this.scheduled_time_slot,
            distance = 0.0, // TODO: Calculate from lat/lng
            urgency = if (this.is_urgent) JobUrgency.URGENT else JobUrgency.NORMAL
        )
    }
    
    fun List<JobDto>.toJobs(): List<Job> = map { it.toJob() }
    
    fun UserDto.toWorkerDisplayProfile(): WorkerDisplayProfile {
        return WorkerDisplayProfile(
            userId = this.id,
            name = this.full_name ?: "Worker",
            phoneNumber = this.phone_number,
            skills = emptyList(), // Would be fetched from worker profile endpoint
            rating = 4.5,
            totalJobs = 0,
            isOnline = false,
            isKycVerified = this.aadhaar_kyc_status == "VERIFIED",
            hasVideoBio = false,
            tier = WorkerTier.BRONZE,
            streakCount = 0,
            memberSince = "2024",
            location = "Delhi NCR"
        )
    }
    
    fun EarningsSummaryDto.toEarningsSummary(): EarningsSummary {
        return EarningsSummary(
            period = this.period,
            totalEarnings = this.total_earnings,
            jobsCompleted = this.jobs_completed,
            averagePerJob = this.average_per_job,
            commissionPaid = this.commission_paid,
            tdsPaid = this.tds_paid,
            gstPaid = this.gst_paid,
            netEarnings = this.net_earnings,
            dailyEarnings = this.daily_earnings.map { 
                DailyEarning(it.date, it.amount, it.job_count) 
            }
        )
    }
    
    private fun mapCategory(category: String): ServiceCategory {
        return when (category.uppercase()) {
            "PLUMBING" -> ServiceCategory.PLUMBER
            "ELECTRICAL" -> ServiceCategory.ELECTRICIAN
            "CARPENTRY" -> ServiceCategory.CARPENTER
            "PAINTING" -> ServiceCategory.PAINTER
            "CLEANING" -> ServiceCategory.CLEANER
            "AC_REPAIR" -> ServiceCategory.AC_REPAIR
            "APPLIANCE_REPAIR" -> ServiceCategory.APPLIANCE_REPAIR
            "PEST_CONTROL" -> ServiceCategory.PEST_CONTROL
            "GARDENING" -> ServiceCategory.GARDENER
            else -> ServiceCategory.HELPER
        }
    }
    
    private fun mapJobStatus(status: String): JobStatus {
        return when (status.uppercase()) {
            "OPEN" -> JobStatus.OPEN
            "ASSIGNED" -> JobStatus.ACCEPTED
            "IN_PROGRESS" -> JobStatus.IN_PROGRESS
            "COMPLETED" -> JobStatus.COMPLETED
            "CANCELLED" -> JobStatus.CANCELLED
            else -> JobStatus.OPEN
        }
    }
    
    // Reverse mapping for creating jobs
    fun ServiceCategory.toApiCategory(): String {
        return when (this) {
            ServiceCategory.PLUMBER, ServiceCategory.PLUMBING -> "PLUMBING"
            ServiceCategory.ELECTRICIAN, ServiceCategory.ELECTRICAL -> "ELECTRICAL"
            ServiceCategory.CARPENTER, ServiceCategory.CARPENTRY -> "CARPENTRY"
            ServiceCategory.PAINTER, ServiceCategory.PAINTING -> "PAINTING"
            ServiceCategory.CLEANER, ServiceCategory.CLEANING -> "CLEANING"
            ServiceCategory.AC_REPAIR -> "AC_REPAIR"
            ServiceCategory.APPLIANCE_REPAIR -> "APPLIANCE_REPAIR"
            ServiceCategory.PEST_CONTROL -> "PEST_CONTROL"
            ServiceCategory.GARDENER, ServiceCategory.GARDENING -> "GARDENING"
            else -> "OTHER"
        }
    }
}

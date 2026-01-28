package com.bero.domain.models

import kotlinx.serialization.Serializable

/**
 * Worker tier levels for the streak system
 */
enum class WorkerTier {
    BRONZE,  // Default: 15% commission
    SILVER,  // 7 day streak: 12% commission
    GOLD     // 30 day streak: 10% commission + free accident insurance
}

/**
 * Worker profile with supply-side attributes
 * Maps to the `worker_profiles` table in CockroachDB
 */
@Serializable
data class WorkerProfile(
    val userId: String,
    val skills: List<String> = emptyList(),
    val h3IndexRes9: String? = null, // Current hex location (Uber H3)
    val isOnline: Boolean = false,
    val walletBalanceMicros: Long = 0L, // Stored in smallest currency unit (paisa)
    val ratingAvg: Double = 0.0,
    val streakCount: Int = 0,
    val lastActiveDate: Long? = null,
    val tier: WorkerTier = WorkerTier.BRONZE,
    val videoBioUrl: String? = null
) {
    /**
     * Get commission rate based on streak tier
     */
    fun getCommissionRate(): Double = when (tier) {
        WorkerTier.BRONZE -> 0.15
        WorkerTier.SILVER -> 0.12
        WorkerTier.GOLD -> 0.10
    }

    /**
     * Check if worker can accept new jobs
     * Blocked if wallet balance is below -500 rupees (-50000000 micros)
     */
    fun canAcceptJobs(): Boolean = walletBalanceMicros >= -50000000L
}

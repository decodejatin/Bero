package com.bero.domain.models

import kotlinx.serialization.Serializable

/**
 * User types in the Bero platform
 */
@Serializable
enum class UserType {
    WORKER,
    CLIENT
}

/**
 * KYC verification status
 */
@Serializable
enum class KycStatus {
    NONE,
    PENDING,
    VERIFIED,
    REJECTED
}

/**
 * Core user identity model
 * Maps to the `users` table in CockroachDB
 */
@Serializable
data class User(
    val id: String,
    val phoneNumber: String,
    val fullName: String? = null,
    val aadhaarKycStatus: KycStatus = KycStatus.NONE,
    val userType: UserType,
    val createdAt: Long = 0L
)


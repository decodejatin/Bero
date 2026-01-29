package com.example.bero.data.models

/**
 * User types in the Bero platform
 */
enum class UserType {
    WORKER,
    CLIENT
}

/**
 * KYC verification status
 */
enum class KycStatus {
    NONE,
    PENDING,
    VERIFIED,
    REJECTED
}

/**
 * Core user identity model
 */
data class User(
    val id: String,
    val phoneNumber: String,
    val fullName: String? = null,
    val aadhaarKycStatus: KycStatus = KycStatus.NONE,
    val userType: UserType,
    val createdAt: Long = 0L
)

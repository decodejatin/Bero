package com.bero.kyc

import com.bero.domain.models.KycStatus
import kotlinx.serialization.Serializable

/**
 * KYC session initiated when user starts verification
 */
@Serializable
data class KycSession(
    val sessionId: String,
    val userId: String,
    val status: KycStatus = KycStatus.PENDING,
    val aadhaarLastFour: String? = null,
    val fullName: String? = null,
    val createdAt: Long = 0L,
    val expiresAt: Long = 0L
)

/**
 * Aadhaar OTP request response
 */
@Serializable
data class AadhaarOtpRequest(
    val transactionId: String,
    val aadhaarNumber: String,
    val otpSent: Boolean = true
)

/**
 * Aadhaar verification result
 */
@Serializable
data class AadhaarVerificationResult(
    val verified: Boolean,
    val fullName: String? = null,
    val gender: String? = null,
    val dateOfBirth: String? = null,
    val address: String? = null,
    val photoBase64: String? = null, // Aadhaar photo
    val errorMessage: String? = null
)

/**
 * KYC document types that can be submitted
 */
enum class KycDocumentType {
    AADHAAR,
    PAN,
    DRIVING_LICENSE,
    VOTER_ID
}

/**
 * KYC error types
 */
sealed class KycError : Exception() {
    data object InvalidAadhaarNumber : KycError()
    data object InvalidOtp : KycError()
    data object OtpExpired : KycError()
    data object SessionExpired : KycError()
    data object VerificationFailed : KycError()
    data object NetworkError : KycError()
    data object AlreadyVerified : KycError()
    data class Unknown(override val message: String) : KycError()
}

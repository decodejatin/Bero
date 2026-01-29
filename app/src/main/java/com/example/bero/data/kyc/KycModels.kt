package com.example.bero.data.kyc

import com.example.bero.data.models.KycStatus

/**
 * KYC session initiated when user starts verification
 */
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
data class AadhaarOtpRequest(
    val transactionId: String,
    val aadhaarNumber: String,
    val otpSent: Boolean = true
)

/**
 * Aadhaar verification result
 */
data class AadhaarVerificationResult(
    val verified: Boolean,
    val fullName: String? = null,
    val gender: String? = null,
    val dateOfBirth: String? = null,
    val address: String? = null,
    val photoBase64: String? = null,
    val errorMessage: String? = null
)

/**
 * KYC document types
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

/**
 * Repository for KYC verification operations.
 */
interface KycRepository {
    suspend fun initiateKyc(userId: String): Result<KycSession>
    suspend fun requestAadhaarOtp(sessionId: String, aadhaarNumber: String): Result<AadhaarOtpRequest>
    suspend fun verifyAadhaarOtp(sessionId: String, transactionId: String, otp: String): Result<AadhaarVerificationResult>
    suspend fun getKycStatus(userId: String): Result<KycStatus>
    suspend fun getKycSession(sessionId: String): Result<KycSession>
}

/**
 * Mock implementation for development
 */
class KycRepositoryImpl(
    private val baseUrl: String = "https://api.bero.app"
) : KycRepository {
    
    private val sessions = mutableMapOf<String, KycSession>()
    private val pendingTransactions = mutableMapOf<String, String>()
    
    override suspend fun initiateKyc(userId: String): Result<KycSession> {
        val session = KycSession(
            sessionId = "kyc-${System.currentTimeMillis()}",
            userId = userId,
            status = KycStatus.PENDING,
            createdAt = System.currentTimeMillis(),
            expiresAt = System.currentTimeMillis() + (30 * 60 * 1000)
        )
        sessions[session.sessionId] = session
        return Result.success(session)
    }
    
    override suspend fun requestAadhaarOtp(sessionId: String, aadhaarNumber: String): Result<AadhaarOtpRequest> {
        if (!isValidAadhaarNumber(aadhaarNumber)) {
            return Result.failure(KycError.InvalidAadhaarNumber)
        }
        
        if (!sessions.containsKey(sessionId)) {
            return Result.failure(KycError.SessionExpired)
        }
        
        val transactionId = "txn-${System.currentTimeMillis()}"
        pendingTransactions[transactionId] = aadhaarNumber
        
        return Result.success(
            AadhaarOtpRequest(
                transactionId = transactionId,
                aadhaarNumber = maskAadhaarNumber(aadhaarNumber),
                otpSent = true
            )
        )
    }
    
    override suspend fun verifyAadhaarOtp(
        sessionId: String,
        transactionId: String,
        otp: String
    ): Result<AadhaarVerificationResult> {
        if (otp.length != 6 || !otp.all { it.isDigit() }) {
            return Result.failure(KycError.InvalidOtp)
        }
        
        val session = sessions[sessionId] ?: return Result.failure(KycError.SessionExpired)
        val aadhaarNumber = pendingTransactions[transactionId] ?: return Result.failure(KycError.SessionExpired)
        
        val updatedSession = session.copy(
            status = KycStatus.VERIFIED,
            aadhaarLastFour = aadhaarNumber.takeLast(4),
            fullName = "Demo User"
        )
        sessions[sessionId] = updatedSession
        pendingTransactions.remove(transactionId)
        
        return Result.success(
            AadhaarVerificationResult(
                verified = true,
                fullName = "Demo User",
                gender = "M",
                dateOfBirth = "01-01-1990",
                address = "Sample Address, City, State - 123456"
            )
        )
    }
    
    override suspend fun getKycStatus(userId: String): Result<KycStatus> {
        val session = sessions.values.find { it.userId == userId }
        return Result.success(session?.status ?: KycStatus.NONE)
    }
    
    override suspend fun getKycSession(sessionId: String): Result<KycSession> {
        val session = sessions[sessionId] ?: return Result.failure(KycError.SessionExpired)
        return Result.success(session)
    }
    
    private fun isValidAadhaarNumber(aadhaar: String): Boolean {
        val cleaned = aadhaar.replace(" ", "").replace("-", "")
        return cleaned.length == 12 && cleaned.all { it.isDigit() }
    }
    
    private fun maskAadhaarNumber(aadhaar: String): String {
        val cleaned = aadhaar.replace(" ", "").replace("-", "")
        return "XXXX XXXX ${cleaned.takeLast(4)}"
    }
}

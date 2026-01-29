package com.bero.kyc

import com.bero.auth.currentTimeMillis
import com.bero.domain.models.KycStatus

/**
 * Repository for KYC verification operations.
 * Integrates with Aadhaar/Digilocker APIs.
 */
interface KycRepository {
    /**
     * Initiate KYC session
     */
    suspend fun initiateKyc(userId: String): Result<KycSession>
    
    /**
     * Request OTP for Aadhaar verification
     * @param aadhaarNumber 12-digit Aadhaar number
     */
    suspend fun requestAadhaarOtp(sessionId: String, aadhaarNumber: String): Result<AadhaarOtpRequest>
    
    /**
     * Verify Aadhaar OTP
     * @param transactionId From requestAadhaarOtp
     * @param otp 6-digit OTP received on registered mobile
     */
    suspend fun verifyAadhaarOtp(
        sessionId: String, 
        transactionId: String, 
        otp: String
    ): Result<AadhaarVerificationResult>
    
    /**
     * Get current KYC status for user
     */
    suspend fun getKycStatus(userId: String): Result<KycStatus>
    
    /**
     * Get KYC session details
     */
    suspend fun getKycSession(sessionId: String): Result<KycSession>
}

/**
 * Mock implementation for development
 * Will be replaced with real Digilocker/HyperVerge integration
 */
class KycRepositoryImpl(
    private val baseUrl: String = "https://api.bero.app"
) : KycRepository {
    
    private val sessions = mutableMapOf<String, KycSession>()
    private val pendingTransactions = mutableMapOf<String, String>() // transactionId -> aadhaarNumber
    
    override suspend fun initiateKyc(userId: String): Result<KycSession> {
        val session = KycSession(
            sessionId = "kyc-${currentTimeMillis()}",
            userId = userId,
            status = KycStatus.PENDING,
            createdAt = currentTimeMillis(),
            expiresAt = currentTimeMillis() + (30 * 60 * 1000) // 30 minutes
        )
        sessions[session.sessionId] = session
        return Result.success(session)
    }
    
    override suspend fun requestAadhaarOtp(sessionId: String, aadhaarNumber: String): Result<AadhaarOtpRequest> {
        // Validate Aadhaar number format
        if (!isValidAadhaarNumber(aadhaarNumber)) {
            return Result.failure(KycError.InvalidAadhaarNumber)
        }
        
        // Check session exists
        if (!sessions.containsKey(sessionId)) {
            return Result.failure(KycError.SessionExpired)
        }
        
        // TODO: Call actual Aadhaar OTP API
        val transactionId = "txn-${currentTimeMillis()}"
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
        // Validate OTP format
        if (otp.length != 6 || !otp.all { it.isDigit() }) {
            return Result.failure(KycError.InvalidOtp)
        }
        
        // Check session and transaction
        val session = sessions[sessionId] ?: return Result.failure(KycError.SessionExpired)
        val aadhaarNumber = pendingTransactions[transactionId] ?: return Result.failure(KycError.SessionExpired)
        
        // TODO: Call actual Aadhaar verification API
        // For development, accept any 6-digit OTP
        
        // Update session as verified
        val updatedSession = session.copy(
            status = KycStatus.VERIFIED,
            aadhaarLastFour = aadhaarNumber.takeLast(4),
            fullName = "Demo User" // Would come from Aadhaar response
        )
        sessions[sessionId] = updatedSession
        
        // Clear pending transaction
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
    
    /**
     * Validate 12-digit Aadhaar number format
     */
    private fun isValidAadhaarNumber(aadhaar: String): Boolean {
        val cleaned = aadhaar.replace(" ", "").replace("-", "")
        return cleaned.length == 12 && cleaned.all { it.isDigit() }
    }
    
    /**
     * Mask Aadhaar number for display (XXXX XXXX 1234)
     */
    private fun maskAadhaarNumber(aadhaar: String): String {
        val cleaned = aadhaar.replace(" ", "").replace("-", "")
        return "XXXX XXXX ${cleaned.takeLast(4)}"
    }
}

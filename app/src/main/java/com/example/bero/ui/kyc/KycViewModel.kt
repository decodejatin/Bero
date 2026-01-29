package com.example.bero.ui.kyc

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.bero.kyc.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * ViewModel for KYC verification flow
 */
class KycViewModel : ViewModel() {
    
    private val kycRepository: KycRepository = KycRepositoryImpl()
    
    private val _uiState = MutableStateFlow(KycUiState())
    val uiState: StateFlow<KycUiState> = _uiState.asStateFlow()
    
    private var currentSessionId: String? = null
    private var currentTransactionId: String? = null
    private var currentAadhaar: String? = null
    
    /**
     * Initialize KYC session
     */
    fun initSession(userId: String) {
        viewModelScope.launch {
            kycRepository.initiateKyc(userId).fold(
                onSuccess = { session ->
                    currentSessionId = session.sessionId
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        error = getErrorMessage(error)
                    )
                }
            )
        }
    }
    
    /**
     * Request OTP for Aadhaar verification
     */
    fun requestAadhaarOtp(aadhaarNumber: String) {
        val sessionId = currentSessionId
        if (sessionId == null) {
            _uiState.value = _uiState.value.copy(error = "Session not initialized")
            return
        }
        
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)
            
            kycRepository.requestAadhaarOtp(sessionId, aadhaarNumber).fold(
                onSuccess = { otpRequest ->
                    currentTransactionId = otpRequest.transactionId
                    currentAadhaar = aadhaarNumber
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        step = KycStep.VERIFY_OTP,
                        maskedAadhaar = otpRequest.aadhaarNumber
                    )
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = getErrorMessage(error)
                    )
                }
            )
        }
    }
    
    /**
     * Verify OTP
     */
    fun verifyOtp(otp: String) {
        val sessionId = currentSessionId
        val transactionId = currentTransactionId
        
        if (sessionId == null || transactionId == null) {
            _uiState.value = _uiState.value.copy(error = "Session expired")
            return
        }
        
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)
            
            kycRepository.verifyAadhaarOtp(sessionId, transactionId, otp).fold(
                onSuccess = { result ->
                    if (result.verified) {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            step = KycStep.SUCCESS,
                            verifiedName = result.fullName
                        )
                    } else {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            step = KycStep.FAILED,
                            error = result.errorMessage ?: "Verification failed"
                        )
                    }
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = getErrorMessage(error)
                    )
                }
            )
        }
    }
    
    /**
     * Resend OTP
     */
    fun resendOtp() {
        currentAadhaar?.let { aadhaar ->
            requestAadhaarOtp(aadhaar)
        }
    }
    
    /**
     * Reset to start over
     */
    fun reset() {
        _uiState.value = KycUiState()
        currentTransactionId = null
        currentAadhaar = null
    }
    
    private fun getErrorMessage(error: Throwable): String {
        return when (error) {
            is KycError.InvalidAadhaarNumber -> "Please enter a valid 12-digit Aadhaar number"
            is KycError.InvalidOtp -> "Invalid OTP. Please try again."
            is KycError.OtpExpired -> "OTP has expired. Please request a new one."
            is KycError.SessionExpired -> "Session expired. Please start again."
            is KycError.VerificationFailed -> "Verification failed. Please try again."
            is KycError.NetworkError -> "Network error. Please check your connection."
            is KycError.AlreadyVerified -> "Your Aadhaar is already verified!"
            is KycError.Unknown -> error.message ?: "An error occurred"
            else -> "An unexpected error occurred"
        }
    }
}

/**
 * KYC verification steps
 */
enum class KycStep {
    ENTER_AADHAAR,
    VERIFY_OTP,
    SUCCESS,
    FAILED
}

/**
 * UI state for KYC verification
 */
data class KycUiState(
    val step: KycStep = KycStep.ENTER_AADHAAR,
    val isLoading: Boolean = false,
    val error: String? = null,
    val maskedAadhaar: String? = null,
    val verifiedName: String? = null
)

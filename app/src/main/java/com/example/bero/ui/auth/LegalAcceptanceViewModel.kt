package com.example.bero.ui.auth

import android.os.Build
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * ViewModel for legal acceptance flow.
 * Manages checkbox states, API calls, and compliance checking.
 */
class LegalAcceptanceViewModel : ViewModel() {

    private val apiClient by lazy { com.example.bero.di.AppContainer.instance.apiClient }

    private val _uiState = MutableStateFlow(LegalAcceptanceUiState())
    val uiState: StateFlow<LegalAcceptanceUiState> = _uiState.asStateFlow()

    /**
     * Toggle the general legal agreement checkbox
     */
    fun toggleGeneralAcceptance(accepted: Boolean) {
        _uiState.value = _uiState.value.copy(generalAccepted = accepted)
    }

    /**
     * Toggle the worker responsibility policy checkbox
     */
    fun toggleWorkerPolicyAcceptance(accepted: Boolean) {
        _uiState.value = _uiState.value.copy(workerPolicyAccepted = accepted)
    }

    /**
     * Submit legal document acceptances to the backend
     */
    fun submitAcceptances(isWorker: Boolean, onSuccess: () -> Unit) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            try {
                val deviceInfo = "${Build.MANUFACTURER} ${Build.MODEL} / Android ${Build.VERSION.RELEASE}"

                // Submit general legal documents acceptance
                val generalDocs = listOf(
                    "terms-conditions",
                    "privacy-policy",
                    "liability-disclaimer",
                    "dispute-resolution"
                )
                val acceptBody = buildAcceptBody(generalDocs, deviceInfo)
                val acceptResult = apiClient.acceptLegalDocuments(acceptBody)
                if (acceptResult.isFailure) {
                    throw acceptResult.exceptionOrNull() ?: Exception("Failed to accept documents")
                }

                // Submit worker policy if worker
                if (isWorker) {
                    val workerBody = """{"device_info":"$deviceInfo"}"""
                    val workerResult = apiClient.acceptWorkerPolicy(workerBody)
                    if (workerResult.isFailure) {
                        throw workerResult.exceptionOrNull() ?: Exception("Failed to accept worker policy")
                    }
                }

                _uiState.value = _uiState.value.copy(isLoading = false, submittedSuccessfully = true)
                onSuccess()
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "Failed to record acceptance: ${e.message}"
                )
            }
        }
    }

    /**
     * Check compliance status from backend
     */
    fun checkCompliance(onNeedsReAcceptance: () -> Unit) {
        viewModelScope.launch {
            try {
                val result = apiClient.getLegalComplianceStatus()
                result.onSuccess { json ->
                    if (json.contains("\"is_compliant\":false") || json.contains("\"is_compliant\": false")) {
                        _uiState.value = _uiState.value.copy(needsReAcceptance = true)
                        onNeedsReAcceptance()
                    }
                }
                // Silently fail on error — don't block the user
            } catch (e: Exception) {
                // Silently fail
            }
        }
    }

    /**
     * Accept re-acceptance modal
     */
    fun acceptReAcceptance(onSuccess: () -> Unit) {
        submitAcceptances(isWorker = false, onSuccess = {
            _uiState.value = _uiState.value.copy(needsReAcceptance = false)
            onSuccess()
        })
    }

    /**
     * Dismiss re-acceptance modal (mark as seen but still show reminder)
     */
    fun dismissReAcceptance() {
        _uiState.value = _uiState.value.copy(needsReAcceptance = false)
    }

    private fun buildAcceptBody(slugs: List<String>, deviceInfo: String): String {
        val acceptances = slugs.joinToString(",") { slug ->
            """{"document_slug":"$slug","version":"v1.0"}"""
        }
        return """{"acceptances":[$acceptances],"device_info":"$deviceInfo"}"""
    }
}

/**
 * UI state for legal acceptance screens
 */
data class LegalAcceptanceUiState(
    val generalAccepted: Boolean = false,
    val workerPolicyAccepted: Boolean = false,
    val isLoading: Boolean = false,
    val error: String? = null,
    val submittedSuccessfully: Boolean = false,
    val needsReAcceptance: Boolean = false
)

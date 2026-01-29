package com.bero.auth

import com.bero.domain.models.KycStatus

/**
 * Interface for managing user session persistence.
 * Implemented differently per platform (DataStore on Android, UserDefaults on iOS).
 */
interface SessionManager {
    /**
     * Save session after successful authentication
     */
    suspend fun saveSession(session: Session)
    
    /**
     * Get current session if exists
     */
    suspend fun getSession(): Session?
    
    /**
     * Clear session on logout
     */
    suspend fun clearSession()
    
    /**
     * Update KYC status in saved session
     */
    suspend fun updateKycStatus(status: KycStatus)
    
    /**
     * Update video bio status in saved session
     */
    suspend fun updateVideoBioStatus(hasVideoBio: Boolean)
    
    /**
     * Check if user is logged in
     */
    suspend fun isLoggedIn(): Boolean = getSession() != null
}

/**
 * In-memory session manager for testing and development
 */
class InMemorySessionManager : SessionManager {
    private var currentSession: Session? = null
    
    override suspend fun saveSession(session: Session) {
        currentSession = session
    }
    
    override suspend fun getSession(): Session? = currentSession
    
    override suspend fun clearSession() {
        currentSession = null
    }
    
    override suspend fun updateKycStatus(status: KycStatus) {
        currentSession = currentSession?.copy(kycStatus = status)
    }
    
    override suspend fun updateVideoBioStatus(hasVideoBio: Boolean) {
        currentSession = currentSession?.copy(hasVideoBio = hasVideoBio)
    }
}


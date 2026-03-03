package com.example.bero.data.location

import android.util.Log
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.NearbyWorkersResponse
import com.example.bero.data.network.TokenManager
import com.example.bero.data.network.UpdateLocationResponse
import kotlinx.serialization.json.Json

/**
 * Repository for location-related API operations.
 * Wraps BeroApiClient calls with error handling and deserialization.
 *
 * Used by:
 * - LocationTrackingService (every 10s location update)
 * - Map UI (query nearby workers)
 * - Worker availability toggle
 */
class LocationRepository(tokenManager: TokenManager) {

    companion object {
        private const val TAG = "LocationRepository"
        private val json = Json { ignoreUnknownKeys = true }
    }

    private val apiClient = BeroApiClient(tokenManager)

    /**
     * Send the worker's current location to the backend.
     * Called every 10 seconds by LocationTrackingService.
     */
    suspend fun updateWorkerLocation(latitude: Double, longitude: Double): UpdateLocationResponse? {
        return try {
            val responseBody = apiClient.updateWorkerLocation(latitude, longitude)
            json.decodeFromString<UpdateLocationResponse>(responseBody)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to update worker location", e)
            null
        }
    }

    /**
     * Query nearby available workers around a point.
     * Used by the client map to show worker markers and
     * internally by the matching system to build cost matrices.
     *
     * @param latitude  Query center latitude
     * @param longitude Query center longitude
     * @param radius    Search radius in meters (default: 2000)
     */
    suspend fun getNearbyWorkers(
        latitude: Double,
        longitude: Double,
        radius: Int = 2000
    ): NearbyWorkersResponse? {
        return try {
            val responseBody = apiClient.getNearbyWorkers(latitude, longitude, radius)
            json.decodeFromString<NearbyWorkersResponse>(responseBody)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to get nearby workers", e)
            null
        }
    }

    /**
     * Toggle worker availability flag on the backend.
     */
    suspend fun setWorkerAvailability(available: Boolean): Boolean {
        return try {
            apiClient.setWorkerAvailability(available)
            true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to set availability", e)
            false
        }
    }

    /**
     * Mark worker as offline (called when service stops).
     */
    suspend fun setWorkerOffline() {
        try {
            apiClient.setWorkerAvailability(false)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to set worker offline", e)
        }
    }
}

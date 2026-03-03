package com.example.bero.data.location

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Context
import android.content.Intent
import android.os.Build
import android.os.IBinder
import android.os.Looper
import android.util.Log
import com.example.bero.data.network.TokenManager
import com.google.android.gms.location.FusedLocationProviderClient
import com.google.android.gms.location.LocationCallback
import com.google.android.gms.location.LocationRequest
import com.google.android.gms.location.LocationResult
import com.google.android.gms.location.LocationServices
import com.google.android.gms.location.Priority
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

/**
 * Foreground service that tracks worker location every 10 seconds.
 *
 * Architecture:
 * - Uses FusedLocationProviderClient for battery-efficient GPS
 * - Sends location updates to Go backend via PUT /api/v1/workers/location
 * - Runs as a foreground service to survive app backgrounding
 * - Controlled by worker availability toggle in the UI
 *
 * Integration with Hungarian matching:
 * - Each location update refreshes the worker's position in the DB
 * - The matching system reads these positions via ST_DWithin spatial queries
 * - Fresher positions → more accurate distance weights → better assignments
 */
class LocationTrackingService : Service() {

    companion object {
        private const val TAG = "LocationTracking"
        private const val CHANNEL_ID = "bero_location_channel"
        private const val NOTIFICATION_ID = 1001
        private const val LOCATION_INTERVAL_MS = 10_000L // 10 seconds
        private const val FASTEST_INTERVAL_MS = 5_000L   // 5 seconds minimum

        fun start(context: Context) {
            val intent = Intent(context, LocationTrackingService::class.java)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                context.startForegroundService(intent)
            } else {
                context.startService(intent)
            }
        }

        fun stop(context: Context) {
            context.stopService(Intent(context, LocationTrackingService::class.java))
        }
    }

    private lateinit var fusedLocationClient: FusedLocationProviderClient
    private lateinit var locationCallback: LocationCallback
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var locationRepository: LocationRepository? = null

    override fun onCreate() {
        super.onCreate()
        Log.d(TAG, "LocationTrackingService created")

        fusedLocationClient = LocationServices.getFusedLocationProviderClient(this)
        createNotificationChannel()

        // Initialize repository with TokenManager
        val tokenManager = TokenManager(applicationContext)
        locationRepository = LocationRepository(tokenManager)

        locationCallback = object : LocationCallback() {
            override fun onLocationResult(result: LocationResult) {
                result.lastLocation?.let { location ->
                    Log.d(TAG, "Location update: ${location.latitude}, ${location.longitude}")
                    serviceScope.launch {
                        try {
                            locationRepository?.updateWorkerLocation(
                                latitude = location.latitude,
                                longitude = location.longitude
                            )
                        } catch (e: Exception) {
                            Log.e(TAG, "Failed to send location update", e)
                        }
                    }
                }
            }
        }
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.d(TAG, "LocationTrackingService started")

        // Start foreground with notification
        startForeground(NOTIFICATION_ID, createNotification())

        // Request location updates
        requestLocationUpdates()

        return START_STICKY
    }

    @Suppress("MissingPermission") // Permission is checked before starting the service
    private fun requestLocationUpdates() {
        val locationRequest = LocationRequest.Builder(
            Priority.PRIORITY_HIGH_ACCURACY,
            LOCATION_INTERVAL_MS
        )
            .setMinUpdateIntervalMillis(FASTEST_INTERVAL_MS)
            .setWaitForAccurateLocation(false)
            .build()

        try {
            fusedLocationClient.requestLocationUpdates(
                locationRequest,
                locationCallback,
                Looper.getMainLooper()
            )
            Log.d(TAG, "Location updates requested (${LOCATION_INTERVAL_MS}ms interval)")
        } catch (e: SecurityException) {
            Log.e(TAG, "Location permission not granted", e)
            stopSelf()
        }
    }

    override fun onDestroy() {
        Log.d(TAG, "LocationTrackingService destroyed")
        fusedLocationClient.removeLocationUpdates(locationCallback)
        serviceScope.cancel()

        // Notify backend that worker is offline
        CoroutineScope(Dispatchers.IO).launch {
            try {
                locationRepository?.setWorkerOffline()
            } catch (e: Exception) {
                Log.e(TAG, "Failed to set worker offline", e)
            }
        }

        super.onDestroy()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "Location Tracking",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "Keeps your location updated for job matching"
                setShowBadge(false)
            }

            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    private fun createNotification(): Notification {
        val builder = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            Notification.Builder(this, CHANNEL_ID)
        } else {
            @Suppress("DEPRECATION")
            Notification.Builder(this)
        }

        return builder
            .setContentTitle("Bero — You're Online")
            .setContentText("Looking for nearby jobs...")
            .setSmallIcon(android.R.drawable.ic_menu_mylocation)
            .setOngoing(true)
            .build()
    }
}

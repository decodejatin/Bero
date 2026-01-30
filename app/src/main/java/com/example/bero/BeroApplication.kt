package com.example.bero

import android.app.Application
import com.example.bero.di.AppContainer

/**
 * Application class for initialization
 */
class BeroApplication : Application() {
    
    override fun onCreate() {
        super.onCreate()
        
        // Initialize dependency injection
        AppContainer.initialize(this)
    }
}

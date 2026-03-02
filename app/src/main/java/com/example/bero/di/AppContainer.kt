package com.example.bero.di

import android.content.Context
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.TokenManager
import com.example.bero.data.repository.AuthRepository
import com.example.bero.data.repository.JobRepository
import com.example.bero.ui.viewmodels.AuthViewModel
import com.example.bero.ui.viewmodels.JobViewModel
import com.example.bero.ui.viewmodels.WorkerHomeViewModel

/**
 * Simple dependency injection container
 * In production, consider using Hilt or Koin
 */
object AppContainer {
    
    private var _instance: AppDependencies? = null
    
    fun initialize(context: Context) {
        if (_instance == null) {
            _instance = AppDependencies(context.applicationContext)
        }
    }
    
    val instance: AppDependencies
        get() = _instance ?: throw IllegalStateException("AppContainer not initialized. Call initialize() first.")
}

class AppDependencies(private val context: Context) {
    
    // Network
    val tokenManager: TokenManager by lazy {
        TokenManager(context)
    }
    
    val apiClient: BeroApiClient by lazy {
        BeroApiClient(tokenManager)
    }
    
    val webSocketClient: com.example.bero.data.network.WebSocketClient by lazy {
        com.example.bero.data.network.WebSocketClient(tokenManager)
    }
    
    // Repositories
    val authRepository: AuthRepository by lazy {
        AuthRepository(apiClient, tokenManager)
    }
    
    val jobRepository: JobRepository by lazy {
        JobRepository(apiClient)
    }
    
    // ViewModels - Note: In production, use ViewModelFactory with Hilt
    fun createAuthViewModel(): AuthViewModel {
        return AuthViewModel(authRepository)
    }
    
    fun createJobViewModel(): JobViewModel {
        return JobViewModel(jobRepository)
    }
    
    fun createWorkerHomeViewModel(): WorkerHomeViewModel {
        return WorkerHomeViewModel()
    }
    
    fun createChatViewModel(): com.example.bero.ui.chat.ChatViewModel {
        return com.example.bero.ui.chat.ChatViewModel(apiClient, webSocketClient, tokenManager)
    }
}

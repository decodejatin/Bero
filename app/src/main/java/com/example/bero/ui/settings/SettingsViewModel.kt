package com.example.bero.ui.settings

import android.app.Application
import android.content.Context
import android.content.res.Configuration
import androidx.appcompat.app.AppCompatDelegate
import androidx.core.os.LocaleListCompat
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class AppSettings(
    val isDarkMode: Boolean = false,
    val languageCode: String = "en",
    val soundEnabled: Boolean = true,
    val locationEnabled: Boolean = true
)

class SettingsViewModel(application: Application) : AndroidViewModel(application) {
    
    private val prefs = application.getSharedPreferences("bero_settings", Context.MODE_PRIVATE)
    
    private val _settings = MutableStateFlow(loadSettings())
    val settings: StateFlow<AppSettings> = _settings.asStateFlow()
    
    private fun loadSettings(): AppSettings {
        return AppSettings(
            isDarkMode = prefs.getBoolean("dark_mode", false),
            languageCode = prefs.getString("language_code", "en") ?: "en",
            soundEnabled = prefs.getBoolean("sound_enabled", true),
            locationEnabled = prefs.getBoolean("location_enabled", true)
        )
    }
    
    fun setDarkMode(enabled: Boolean) {
        viewModelScope.launch {
            prefs.edit().putBoolean("dark_mode", enabled).apply()
            _settings.value = _settings.value.copy(isDarkMode = enabled)
        }
    }
    
    fun setLanguage(languageCode: String) {
        viewModelScope.launch {
            prefs.edit().putString("language_code", languageCode).apply()
            _settings.value = _settings.value.copy(languageCode = languageCode)
            
            // Apply locale change using AppCompat
            val localeList = LocaleListCompat.forLanguageTags(languageCode)
            AppCompatDelegate.setApplicationLocales(localeList)
        }
    }
    
    fun setSoundEnabled(enabled: Boolean) {
        viewModelScope.launch {
            prefs.edit().putBoolean("sound_enabled", enabled).apply()
            _settings.value = _settings.value.copy(soundEnabled = enabled)
        }
    }
    
    fun setLocationEnabled(enabled: Boolean) {
        viewModelScope.launch {
            prefs.edit().putBoolean("location_enabled", enabled).apply()
            _settings.value = _settings.value.copy(locationEnabled = enabled)
        }
    }
    
    fun getLanguageDisplayName(): String {
        return when (_settings.value.languageCode) {
            "hi" -> "हिंदी"
            "ta" -> "தமிழ்"
            "bn" -> "বাংলা"
            "te" -> "తెలుగు"
            else -> "English"
        }
    }
}

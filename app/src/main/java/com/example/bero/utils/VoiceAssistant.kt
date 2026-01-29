package com.example.bero.utils

import android.content.Context
import android.speech.tts.TextToSpeech
import android.speech.tts.UtteranceProgressListener
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.util.Locale
import java.util.UUID

/**
 * Voice Assistant for text-to-speech functionality.
 * Designed for low-literacy users who benefit from audio feedback.
 * 
 * Supports Hindi and English with automatic language detection.
 */
class VoiceAssistant(
    private val context: Context,
    private val defaultLocale: Locale = Locale("hi", "IN") // Hindi as default for Indian workers
) {
    private var tts: TextToSpeech? = null
    
    private val _isReady = MutableStateFlow(false)
    val isReady: StateFlow<Boolean> = _isReady.asStateFlow()
    
    private val _isSpeaking = MutableStateFlow(false)
    val isSpeaking: StateFlow<Boolean> = _isSpeaking.asStateFlow()
    
    /**
     * Initialize the TTS engine
     */
    fun initialize() {
        tts = TextToSpeech(context) { status ->
            if (status == TextToSpeech.SUCCESS) {
                val result = tts?.setLanguage(defaultLocale)
                if (result == TextToSpeech.LANG_MISSING_DATA || result == TextToSpeech.LANG_NOT_SUPPORTED) {
                    // Fall back to English
                    tts?.setLanguage(Locale.US)
                }
                
                // Set up progress listener
                tts?.setOnUtteranceProgressListener(object : UtteranceProgressListener() {
                    override fun onStart(utteranceId: String?) {
                        _isSpeaking.value = true
                    }
                    
                    override fun onDone(utteranceId: String?) {
                        _isSpeaking.value = false
                    }
                    
                    override fun onError(utteranceId: String?) {
                        _isSpeaking.value = false
                    }
                })
                
                _isReady.value = true
            }
        }
    }
    
    /**
     * Speak the given text
     * @param text Text to speak
     * @param queueMode Whether to queue after current speech or interrupt
     */
    fun speak(text: String, queueMode: Int = TextToSpeech.QUEUE_FLUSH) {
        if (_isReady.value) {
            val utteranceId = UUID.randomUUID().toString()
            tts?.speak(text, queueMode, null, utteranceId)
        }
    }
    
    /**
     * Speak text in specific language
     */
    fun speakInLanguage(text: String, locale: Locale) {
        if (_isReady.value) {
            val currentLocale = tts?.voice?.locale
            tts?.setLanguage(locale)
            speak(text)
            // Restore default locale
            currentLocale?.let { tts?.setLanguage(it) }
        }
    }
    
    /**
     * Pre-defined voice alerts for common scenarios
     */
    object Alerts {
        // Hindi alerts
        const val NEW_JOB_AVAILABLE_HI = "नया काम आया है! देखने के लिए टैप करें।"
        const val JOB_ACCEPTED_HI = "काम मिल गया! जल्दी पहुंचें।"
        const val PAYMENT_RECEIVED_HI = "पैसे आ गए! अपना बैलेंस चेक करें।"
        const val VERIFICATION_PENDING_HI = "अपना KYC पूरा करें काम पाने के लिए।"
        const val PROFILE_INCOMPLETE_HI = "वीडियो बनाएं ज्यादा काम पाने के लिए।"
        
        // English alerts
        const val NEW_JOB_AVAILABLE_EN = "New job available! Tap to view."
        const val JOB_ACCEPTED_EN = "Job confirmed! Please arrive on time."
        const val PAYMENT_RECEIVED_EN = "Payment received! Check your balance."
        const val VERIFICATION_PENDING_EN = "Complete KYC to start getting jobs."
        const val PROFILE_INCOMPLETE_EN = "Record a video to get more jobs."
    }
    
    /**
     * Announce new job availability
     */
    fun announceNewJob(isHindi: Boolean = true) {
        speak(if (isHindi) Alerts.NEW_JOB_AVAILABLE_HI else Alerts.NEW_JOB_AVAILABLE_EN)
    }
    
    /**
     * Announce job confirmation
     */
    fun announceJobAccepted(isHindi: Boolean = true) {
        speak(if (isHindi) Alerts.JOB_ACCEPTED_HI else Alerts.JOB_ACCEPTED_EN)
    }
    
    /**
     * Announce payment received
     */
    fun announcePaymentReceived(isHindi: Boolean = true) {
        speak(if (isHindi) Alerts.PAYMENT_RECEIVED_HI else Alerts.PAYMENT_RECEIVED_EN)
    }
    
    /**
     * Stop speaking
     */
    fun stop() {
        tts?.stop()
        _isSpeaking.value = false
    }
    
    /**
     * Release TTS resources
     */
    fun shutdown() {
        tts?.stop()
        tts?.shutdown()
        tts = null
        _isReady.value = false
        _isSpeaking.value = false
    }
}

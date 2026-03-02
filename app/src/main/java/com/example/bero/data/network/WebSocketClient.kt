package com.example.bero.data.network

import android.util.Log
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import java.util.concurrent.TimeUnit

/**
 * WebSocket client for real-time chat messaging
 * Features: auto-reconnect with exponential backoff, message flows
 */
class WebSocketClient(private val tokenManager: TokenManager) {

    private val TAG = "WebSocketClient"
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    
    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
    }

    private val client = OkHttpClient.Builder()
        .readTimeout(0, TimeUnit.MILLISECONDS) // No timeout for WebSocket
        .pingInterval(30, TimeUnit.SECONDS)
        .build()

    private var webSocket: WebSocket? = null
    private var isConnecting = false
    private var reconnectAttempts = 0
    private val maxReconnectAttempts = 10

    // State
    private val _connectionState = MutableStateFlow(ConnectionState.DISCONNECTED)
    val connectionState: StateFlow<ConnectionState> = _connectionState

    // Incoming messages flow
    private val _incomingMessages = MutableSharedFlow<WsIncomingMessage>(extraBufferCapacity = 64)
    val incomingMessages: SharedFlow<WsIncomingMessage> = _incomingMessages

    // Typing indicators flow
    private val _typingIndicators = MutableSharedFlow<TypingIndicator>(extraBufferCapacity = 16)
    val typingIndicators: SharedFlow<TypingIndicator> = _typingIndicators

    fun connect() {
        if (isConnecting || _connectionState.value == ConnectionState.CONNECTED) return
        
        val token = tokenManager.getAccessToken() ?: run {
            Log.w(TAG, "No access token, cannot connect WebSocket")
            return
        }

        isConnecting = true
        _connectionState.value = ConnectionState.CONNECTING

        val wsUrl = "${ApiConfig.wsBaseUrl}?token=$token"
        val request = Request.Builder()
            .url(wsUrl)
            .build()

        webSocket = client.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                Log.d(TAG, "WebSocket connected")
                _connectionState.value = ConnectionState.CONNECTED
                isConnecting = false
                reconnectAttempts = 0
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                try {
                    val baseMsg = json.decodeFromString<WsBaseMessage>(text)
                    when (baseMsg.type) {
                        "new_message", "message_sent" -> {
                            val msg = json.decodeFromString<WsIncomingMessage>(text)
                            scope.launch { _incomingMessages.emit(msg) }
                        }
                        "typing" -> {
                            val typing = json.decodeFromString<TypingIndicator>(text)
                            scope.launch { _typingIndicators.emit(typing) }
                        }
                    }
                } catch (e: Exception) {
                    Log.e(TAG, "Failed to parse WS message: ${e.message}")
                }
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                Log.d(TAG, "WebSocket closing: $code $reason")
                webSocket.close(1000, null)
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                Log.d(TAG, "WebSocket closed: $code $reason")
                _connectionState.value = ConnectionState.DISCONNECTED
                isConnecting = false
                scheduleReconnect()
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                Log.e(TAG, "WebSocket failure: ${t.message}")
                _connectionState.value = ConnectionState.DISCONNECTED
                isConnecting = false
                scheduleReconnect()
            }
        })
    }

    fun sendMessage(conversationId: String, content: String) {
        val msg = WsOutgoingMessage(
            type = "message",
            conversation_id = conversationId,
            content = content
        )
        val text = json.encodeToString(WsOutgoingMessage.serializer(), msg)
        webSocket?.send(text)
    }

    fun sendTyping(conversationId: String) {
        val msg = WsOutgoingMessage(
            type = "typing",
            conversation_id = conversationId,
            content = ""
        )
        val text = json.encodeToString(WsOutgoingMessage.serializer(), msg)
        webSocket?.send(text)
    }

    fun disconnect() {
        reconnectAttempts = maxReconnectAttempts // Prevent auto-reconnect
        webSocket?.close(1000, "User disconnected")
        webSocket = null
        _connectionState.value = ConnectionState.DISCONNECTED
    }

    private fun scheduleReconnect() {
        if (reconnectAttempts >= maxReconnectAttempts) return
        
        reconnectAttempts++
        val delayMs = minOf(1000L * (1 shl reconnectAttempts), 30_000L) // Exponential backoff, max 30s
        
        scope.launch {
            Log.d(TAG, "Reconnecting in ${delayMs}ms (attempt $reconnectAttempts)")
            delay(delayMs)
            connect()
        }
    }

    enum class ConnectionState {
        DISCONNECTED, CONNECTING, CONNECTED
    }
}

// WebSocket message types
@Serializable
data class WsBaseMessage(val type: String)

@Serializable
data class WsOutgoingMessage(
    val type: String,
    val conversation_id: String,
    val content: String
)

@Serializable
data class WsIncomingMessage(
    val type: String,
    val conversation_id: String,
    val message: WsChatMessage? = null
)

@Serializable
data class WsChatMessage(
    val id: String,
    val conversation_id: String,
    val sender_id: String,
    val content: String,
    val message_type: String = "text",
    val is_read: Boolean = false,
    val created_at: String = ""
)

@Serializable
data class TypingIndicator(
    val type: String,
    val conversation_id: String,
    val user_id: String
)

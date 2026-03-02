package com.example.bero.ui.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.bero.data.models.ChatMessage
import com.example.bero.data.models.Conversation
import com.example.bero.data.models.MessageType
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.ChatMessageDto
import com.example.bero.data.network.ConversationDto
import com.example.bero.data.network.TokenManager
import com.example.bero.data.network.WebSocketClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/**
 * ViewModel for managing chat conversations and messages
 */
class ChatViewModel(
    private val apiClient: BeroApiClient,
    private val webSocketClient: WebSocketClient,
    private val tokenManager: TokenManager
) : ViewModel() {

    private val _conversations = MutableStateFlow<List<Conversation>>(emptyList())
    val conversations: StateFlow<List<Conversation>> = _conversations

    private val _messages = MutableStateFlow<List<ChatMessage>>(emptyList())
    val messages: StateFlow<List<ChatMessage>> = _messages

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading

    private val _currentConversationId = MutableStateFlow<String?>(null)
    val currentConversationId: StateFlow<String?> = _currentConversationId

    private val _unreadCount = MutableStateFlow(0)
    val unreadCount: StateFlow<Int> = _unreadCount

    private val currentUserId: String?
        get() = tokenManager.getUserId()

    init {
        // Connect WebSocket and listen for messages
        webSocketClient.connect()
        collectWebSocketMessages()
        loadConversations()
    }

    fun loadConversations() {
        viewModelScope.launch {
            _isLoading.value = true
            apiClient.getConversations().fold(
                onSuccess = { dtos ->
                    _conversations.value = dtos.map { it.toConversation() }
                },
                onFailure = {
                    // Keep existing list on error
                }
            )
            _isLoading.value = false
        }
        
        // Also refresh unread count
        viewModelScope.launch {
            apiClient.getUnreadCount().fold(
                onSuccess = { count -> _unreadCount.value = count },
                onFailure = { /* ignore */ }
            )
        }
    }

    fun openConversation(conversationId: String) {
        _currentConversationId.value = conversationId
        loadMessages(conversationId)
        markAsRead(conversationId)
    }

    fun openOrCreateConversation(participantId: String, participantName: String, jobId: String? = null) {
        viewModelScope.launch {
            _isLoading.value = true
            apiClient.getOrCreateConversation(participantId, jobId).fold(
                onSuccess = { dto ->
                    val conv = dto.toConversation()
                    // Add to list if not already there
                    _conversations.update { list ->
                        if (list.any { it.id == conv.id }) list
                        else listOf(conv) + list
                    }
                    openConversation(conv.id)
                },
                onFailure = {
                    // Handle error
                }
            )
            _isLoading.value = false
        }
    }

    private fun loadMessages(conversationId: String) {
        viewModelScope.launch {
            _isLoading.value = true
            apiClient.getMessages(conversationId).fold(
                onSuccess = { dtos ->
                    // API returns newest first, reverse for display (oldest first)
                    _messages.value = dtos.reversed().map { it.toChatMessage() }
                },
                onFailure = {
                    _messages.value = emptyList()
                }
            )
            _isLoading.value = false
        }
    }

    fun sendMessage(content: String) {
        val conversationId = _currentConversationId.value ?: return
        val userId = currentUserId ?: return
        
        if (content.isBlank()) return

        // Optimistic: add message to local list immediately
        val tempMessage = ChatMessage(
            senderId = userId,
            receiverId = "", // Unknown until API confirms
            message = content,
            isRead = false,
            messageType = MessageType.TEXT
        )
        _messages.update { it + tempMessage }

        // Send via WebSocket (primary) or REST (fallback)
        if (webSocketClient.connectionState.value == WebSocketClient.ConnectionState.CONNECTED) {
            webSocketClient.sendMessage(conversationId, content)
        } else {
            // Fallback to REST
            viewModelScope.launch {
                apiClient.sendChatMessage(conversationId, content).fold(
                    onSuccess = { dto ->
                        // Replace temp message with confirmed one
                        _messages.update { list ->
                            list.map { if (it.id == tempMessage.id) dto.toChatMessage() else it }
                        }
                    },
                    onFailure = {
                        // Remove temp message on failure
                        _messages.update { list ->
                            list.filter { it.id != tempMessage.id }
                        }
                    }
                )
            }
        }

        // Update conversation list
        _conversations.update { list ->
            list.map { conv ->
                if (conv.id == conversationId) {
                    conv.copy(lastMessage = content, lastMessageTime = System.currentTimeMillis())
                } else conv
            }
        }
    }

    fun markAsRead(conversationId: String) {
        viewModelScope.launch {
            apiClient.markConversationAsRead(conversationId)
            // Update local unread count
            _conversations.update { list ->
                list.map { conv ->
                    if (conv.id == conversationId) conv.copy(unreadCount = 0) else conv
                }
            }
            // Refresh total unread
            apiClient.getUnreadCount().fold(
                onSuccess = { count -> _unreadCount.value = count },
                onFailure = {}
            )
        }
    }

    private fun collectWebSocketMessages() {
        viewModelScope.launch {
            webSocketClient.incomingMessages.collect { wsMsg ->
                val msg = wsMsg.message ?: return@collect
                val chatMessage = ChatMessage(
                    id = msg.id,
                    senderId = msg.sender_id,
                    receiverId = currentUserId ?: "",
                    message = msg.content,
                    isRead = msg.is_read,
                    messageType = when (msg.message_type) {
                        "image" -> MessageType.IMAGE
                        "system" -> MessageType.SYSTEM
                        else -> MessageType.TEXT
                    }
                )

                // Add to current conversation if active
                if (wsMsg.conversation_id == _currentConversationId.value) {
                    _messages.update { it + chatMessage }
                }

                // Update conversation list
                _conversations.update { list ->
                    list.map { conv ->
                        if (conv.id == wsMsg.conversation_id) {
                            conv.copy(
                                lastMessage = msg.content,
                                lastMessageTime = System.currentTimeMillis(),
                                unreadCount = if (wsMsg.conversation_id == _currentConversationId.value) 0 else conv.unreadCount + 1
                            )
                        } else conv
                    }
                }

                // Refresh unread count
                _unreadCount.update { it + 1 }
            }
        }
    }

    fun closeConversation() {
        _currentConversationId.value = null
        _messages.value = emptyList()
    }

    override fun onCleared() {
        super.onCleared()
        webSocketClient.disconnect()
    }

    // DTO mapping helpers
    private fun ConversationDto.toConversation() = Conversation(
        id = id,
        participantId = participant_id,
        participantName = participant_name,
        lastMessage = last_message,
        lastMessageTime = try {
            java.text.SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss", java.util.Locale.US)
                .parse(last_message_at)?.time ?: System.currentTimeMillis()
        } catch (e: Exception) { System.currentTimeMillis() },
        unreadCount = unread_count.toInt()
    )

    private fun ChatMessageDto.toChatMessage() = ChatMessage(
        id = id,
        senderId = sender_id,
        receiverId = "", // Will be filled by conversation context
        message = content,
        isRead = is_read,
        messageType = when (message_type) {
            "image" -> MessageType.IMAGE
            "system" -> MessageType.SYSTEM
            else -> MessageType.TEXT
        }
    )
}

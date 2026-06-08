package com.example.bero.ui.auth

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.slideInVertically
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.interaction.collectIsPressedAsState
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.models.UserType
import com.example.bero.ui.theme.*

/**
 * Premium Role Selection Screen — luxury dark design with legal acceptance checkboxes
 */
@Composable
fun RoleSelectionScreen(
    onRoleSelected: (UserType) -> Unit,
    isLoading: Boolean = false,
    error: String? = null,
    onDocumentClick: (slug: String, title: String) -> Unit = { _, _ -> }
) {
    var selectedRole by remember { mutableStateOf<UserType?>(null) }
    var generalAccepted by remember { mutableStateOf(false) }
    var workerPolicyAccepted by remember { mutableStateOf(false) }
    var showValidationError by remember { mutableStateOf(false) }

    // Determine which checkboxes are required
    val isWorkerSelected = selectedRole == UserType.WORKER
    val canProceed = generalAccepted && (!isWorkerSelected || workerPolicyAccepted)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                brush = Brush.verticalGradient(
                    colors = listOf(
                        LuxuryCharcoal,
                        LuxuryOnyx,
                        LuxuryCharcoal
                    )
                )
            )
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(48.dp))

            Text(
                text = "WELCOME",
                style = MaterialTheme.typography.headlineLarge,
                fontWeight = FontWeight.ExtraBold,
                color = Color(0xFFF5F0E8),
                letterSpacing = 6.sp
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "How would you like to use Bero?",
                style = MaterialTheme.typography.bodyLarge,
                color = Color(0xFFF5F0E8).copy(alpha = 0.5f)
            )

            Spacer(modifier = Modifier.height(36.dp))

            // Worker Card
            PremiumRoleCard(
                title = "I want to Work",
                description = "Find jobs, grow your reputation, and earn money",
                emoji = "🛠️",
                accentColor = LuxuryGold,
                isSelected = selectedRole == UserType.WORKER,
                onClick = {
                    selectedRole = UserType.WORKER
                    showValidationError = false
                }
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Client Card
            PremiumRoleCard(
                title = "I want to Hire",
                description = "Post jobs and find skilled professionals nearby",
                emoji = "🔍",
                accentColor = LuxuryGoldLight,
                isSelected = selectedRole == UserType.CLIENT,
                onClick = {
                    selectedRole = UserType.CLIENT
                    showValidationError = false
                }
            )

            // Legal acceptance section
            AnimatedVisibility(
                visible = selectedRole != null,
                enter = fadeIn() + slideInVertically(initialOffsetY = { it / 2 })
            ) {
                Column(
                    modifier = Modifier.padding(top = 24.dp)
                ) {
                    // Divider
                    HorizontalDivider(
                        color = Color(0xFFF5F0E8).copy(alpha = 0.1f),
                        modifier = Modifier.padding(vertical = 8.dp)
                    )

                    Text(
                        text = "Legal Agreement",
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.Bold,
                        color = LuxuryGold,
                        letterSpacing = 2.sp,
                        modifier = Modifier.padding(bottom = 12.dp)
                    )

                    // General legal checkbox
                    LegalCheckboxItem(
                        checked = generalAccepted,
                        onCheckedChange = {
                            generalAccepted = it
                            showValidationError = false
                        },
                        content = {
                            val annotatedText = buildAnnotatedString {
                                append("I have read and agree to the ")
                                withStyle(style = SpanStyle(
                                    color = LuxuryGold,
                                    textDecoration = TextDecoration.Underline
                                )) {
                                    pushStringAnnotation("doc", "terms-conditions")
                                    append("Terms & Conditions")
                                    pop()
                                }
                                append(", ")
                                withStyle(style = SpanStyle(
                                    color = LuxuryGold,
                                    textDecoration = TextDecoration.Underline
                                )) {
                                    pushStringAnnotation("doc", "privacy-policy")
                                    append("Privacy Policy")
                                    pop()
                                }
                                append(", ")
                                withStyle(style = SpanStyle(
                                    color = LuxuryGold,
                                    textDecoration = TextDecoration.Underline
                                )) {
                                    pushStringAnnotation("doc", "liability-disclaimer")
                                    append("Liability Disclaimer")
                                    pop()
                                }
                                append(", and ")
                                withStyle(style = SpanStyle(
                                    color = LuxuryGold,
                                    textDecoration = TextDecoration.Underline
                                )) {
                                    pushStringAnnotation("doc", "dispute-resolution")
                                    append("Dispute Resolution Policy")
                                    pop()
                                }
                                append(".")
                            }

                            // Use ClickableText for individual document clicks
                            androidx.compose.foundation.text.ClickableText(
                                text = annotatedText,
                                style = MaterialTheme.typography.bodySmall.copy(
                                    color = Color(0xFFF5F0E8).copy(alpha = 0.7f),
                                    lineHeight = 20.sp
                                ),
                                onClick = { offset ->
                                    annotatedText.getStringAnnotations("doc", offset, offset)
                                        .firstOrNull()?.let { annotation ->
                                            val title = when (annotation.item) {
                                                "terms-conditions" -> "Terms & Conditions"
                                                "privacy-policy" -> "Privacy Policy"
                                                "liability-disclaimer" -> "Liability Disclaimer"
                                                "dispute-resolution" -> "Dispute Resolution Policy"
                                                else -> annotation.item
                                            }
                                            onDocumentClick(annotation.item, title)
                                        }
                                }
                            )
                        }
                    )

                    // Worker responsibility checkbox (only for workers)
                    AnimatedVisibility(
                        visible = isWorkerSelected,
                        enter = fadeIn() + slideInVertically(initialOffsetY = { it / 2 })
                    ) {
                        LegalCheckboxItem(
                            checked = workerPolicyAccepted,
                            onCheckedChange = {
                                workerPolicyAccepted = it
                                showValidationError = false
                            },
                            content = {
                                val workerText = buildAnnotatedString {
                                    append("I acknowledge and agree to the ")
                                    withStyle(style = SpanStyle(
                                        color = LuxuryGold,
                                        textDecoration = TextDecoration.Underline
                                    )) {
                                        pushStringAnnotation("doc", "worker-responsibility")
                                        append("Worker Responsibility Policy")
                                        pop()
                                    }
                                    append(" and confirm I am an independent contractor.")
                                }
                                androidx.compose.foundation.text.ClickableText(
                                    text = workerText,
                                    style = MaterialTheme.typography.bodySmall.copy(
                                        color = Color(0xFFF5F0E8).copy(alpha = 0.7f),
                                        lineHeight = 20.sp
                                    ),
                                    onClick = { offset ->
                                        workerText.getStringAnnotations("doc", offset, offset)
                                            .firstOrNull()?.let { annotation ->
                                                onDocumentClick(annotation.item, "Worker Responsibility Policy")
                                            }
                                    }
                                )
                            },
                            modifier = Modifier.padding(top = 8.dp)
                        )
                    }

                    // Validation error
                    if (showValidationError) {
                        Text(
                            text = "Please accept all required agreements to continue",
                            color = BeroError,
                            style = MaterialTheme.typography.bodySmall,
                            modifier = Modifier.padding(top = 8.dp, start = 40.dp)
                        )
                    }

                    Spacer(modifier = Modifier.height(24.dp))

                    // Continue button
                    Button(
                        onClick = {
                            if (canProceed && selectedRole != null) {
                                onRoleSelected(selectedRole!!)
                            } else {
                                showValidationError = true
                            }
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(52.dp),
                        shape = RoundedCornerShape(16.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = if (canProceed) LuxuryGold else LuxuryGold.copy(alpha = 0.3f),
                            contentColor = LuxuryCharcoal
                        ),
                        enabled = !isLoading
                    ) {
                        if (isLoading) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(24.dp),
                                color = LuxuryCharcoal,
                                strokeWidth = 2.dp
                            )
                        } else {
                            Text(
                                text = "Continue",
                                fontWeight = FontWeight.Bold,
                                fontSize = 16.sp
                            )
                        }
                    }
                }
            }

            if (error != null) {
                Spacer(modifier = Modifier.height(16.dp))
                Text(
                    text = error,
                    color = BeroError,
                    style = MaterialTheme.typography.bodyMedium,
                    textAlign = TextAlign.Center
                )
            }

            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@Composable
private fun LegalCheckboxItem(
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
    content: @Composable () -> Unit,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier.fillMaxWidth(),
        verticalAlignment = Alignment.Top
    ) {
        Checkbox(
            checked = checked,
            onCheckedChange = onCheckedChange,
            colors = CheckboxDefaults.colors(
                checkedColor = LuxuryGold,
                uncheckedColor = Color(0xFFF5F0E8).copy(alpha = 0.4f),
                checkmarkColor = LuxuryCharcoal
            ),
            modifier = Modifier.padding(end = 4.dp)
        )
        Box(modifier = Modifier.padding(top = 12.dp)) {
            content()
        }
    }
}

@Composable
private fun PremiumRoleCard(
    title: String,
    description: String,
    emoji: String,
    accentColor: Color,
    isSelected: Boolean = false,
    onClick: () -> Unit
) {
    val interactionSource = remember { MutableInteractionSource() }
    val isPressed by interactionSource.collectIsPressedAsState()

    val scale by animateFloatAsState(
        targetValue = if (isPressed) 0.97f else 1f,
        animationSpec = tween(100),
        label = "cardScale"
    )

    val borderColor by animateColorAsState(
        targetValue = if (isSelected) accentColor.copy(alpha = 0.6f) else accentColor.copy(alpha = 0.2f),
        label = "borderColor"
    )

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .height(120.dp)
            .scale(scale)
            .clickable(
                interactionSource = interactionSource,
                indication = null,
                onClick = onClick
            ),
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(
            containerColor = if (isSelected) Color(0xFF1E1E32) else Color(0xFF1E1E32).copy(alpha = 0.65f)
        ),
        border = BorderStroke(if (isSelected) 2.dp else 1.dp, borderColor),
        elevation = CardDefaults.cardElevation(
            defaultElevation = if (isPressed) 2.dp else if (isSelected) 8.dp else 4.dp
        )
    ) {
        Row(
            modifier = Modifier
                .fillMaxSize()
                .padding(20.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(52.dp)
                    .background(
                        color = accentColor.copy(alpha = if (isSelected) 0.2f else 0.1f),
                        shape = RoundedCornerShape(14.dp)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Text(emoji, fontSize = 26.sp)
            }

            Spacer(modifier = Modifier.width(14.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFFF5F0E8)
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = description,
                    style = MaterialTheme.typography.bodySmall,
                    color = Color(0xFFF5F0E8).copy(alpha = 0.5f),
                    lineHeight = 18.sp
                )
            }

            if (isSelected) {
                Surface(
                    shape = RoundedCornerShape(8.dp),
                    color = accentColor.copy(alpha = 0.2f)
                ) {
                    Text(
                        text = "✓",
                        color = accentColor,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)
                    )
                }
            } else {
                Icon(
                    Icons.Default.ChevronRight,
                    contentDescription = null,
                    tint = accentColor.copy(alpha = 0.6f),
                    modifier = Modifier.size(24.dp)
                )
            }
        }
    }
}

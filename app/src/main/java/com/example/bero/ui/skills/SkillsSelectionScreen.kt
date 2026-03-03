@file:OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)

package com.example.bero.ui.skills

import androidx.compose.animation.animateColorAsState
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

/**
 * Full-screen skills selection with predefined categories + custom skill input.
 * Used on first worker login and from settings.
 */
@Composable
fun SkillsSelectionScreen(
    initialSkills: List<String> = emptyList(),
    isFirstTime: Boolean = false,
    onSave: (List<String>) -> Unit = {},
    onBackClick: () -> Unit = {}
) {
    val predefinedSkills = listOf(
        "🔧" to "Plumbing",
        "⚡" to "Electrical",
        "🪚" to "Carpentry",
        "🎨" to "Painting",
        "🧹" to "Cleaning",
        "❄️" to "AC Repair",
        "🔩" to "Mechanic",
        "🏗️" to "Construction",
        "🌿" to "Gardening",
        "🚚" to "Moving",
        "📦" to "Delivery",
        "👨‍🍳" to "Cooking",
        "💇" to "Beauty & Grooming",
        "📱" to "Phone Repair",
        "💻" to "Computer Repair",
        "🚰" to "Water Purifier",
        "🔑" to "Locksmith",
        "🐾" to "Pet Care"
    )

    var selectedSkills by remember { mutableStateOf(initialSkills.toMutableSet()) }
    var customSkillText by remember { mutableStateOf("") }
    var isSaving by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        if (isFirstTime) "Select Your Skills" else "Edit Skills",
                        fontWeight = FontWeight.Bold
                    )
                },
                navigationIcon = {
                    if (!isFirstTime) {
                        IconButton(onClick = onBackClick) {
                            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                        }
                    }
                }
            )
        },
        bottomBar = {
            Surface(
                shadowElevation = 16.dp,
                color = MaterialTheme.colorScheme.surface
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp)
                ) {
                    if (selectedSkills.isEmpty()) {
                        Text(
                            "Select at least 1 skill to continue",
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            fontSize = 13.sp,
                            modifier = Modifier.padding(bottom = 8.dp)
                        )
                    } else {
                        Text(
                            "${selectedSkills.size} skill${if (selectedSkills.size > 1) "s" else ""} selected",
                            color = MaterialTheme.colorScheme.primary,
                            fontWeight = FontWeight.SemiBold,
                            fontSize = 13.sp,
                            modifier = Modifier.padding(bottom = 8.dp)
                        )
                    }
                    Button(
                        onClick = {
                            isSaving = true
                            onSave(selectedSkills.toList())
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(52.dp),
                        shape = RoundedCornerShape(12.dp),
                        enabled = selectedSkills.isNotEmpty() && !isSaving
                    ) {
                        if (isSaving) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(20.dp),
                                strokeWidth = 2.dp,
                                color = Color.White
                            )
                        } else {
                            Icon(Icons.Default.Check, contentDescription = null)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(
                                if (isFirstTime) "Continue" else "Save Skills",
                                fontSize = 16.sp,
                                fontWeight = FontWeight.SemiBold
                            )
                        }
                    }
                }
            }
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            // Header
            item {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    if (isFirstTime) "What can you do? Select all skills that apply."
                    else "Tap to toggle skills. Add custom ones below.",
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    fontSize = 14.sp
                )
                Spacer(modifier = Modifier.height(16.dp))
            }

            // Predefined skills as a flow layout simulated with LazyColumn items
            item {
                Text(
                    "Popular Skills",
                    fontWeight = FontWeight.Bold,
                    fontSize = 16.sp,
                    modifier = Modifier.padding(bottom = 8.dp)
                )
            }

            item {
                FlowRow(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    predefinedSkills.forEach { (emoji, skill) ->
                        val isSelected = skill in selectedSkills
                        val bgColor by animateColorAsState(
                            if (isSelected) MaterialTheme.colorScheme.primary
                            else MaterialTheme.colorScheme.surface,
                            label = "chipBg"
                        )
                        val contentColor by animateColorAsState(
                            if (isSelected) MaterialTheme.colorScheme.onPrimary
                            else MaterialTheme.colorScheme.onSurface,
                            label = "chipContent"
                        )

                        Surface(
                            onClick = {
                                selectedSkills = selectedSkills.toMutableSet().apply {
                                    if (isSelected) remove(skill) else add(skill)
                                }
                            },
                            shape = RoundedCornerShape(20.dp),
                            color = bgColor,
                            border = BorderStroke(
                                1.dp,
                                if (isSelected) MaterialTheme.colorScheme.primary
                                else MaterialTheme.colorScheme.outlineVariant
                            )
                        ) {
                            Row(
                                modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(emoji, fontSize = 16.sp)
                                Spacer(modifier = Modifier.width(6.dp))
                                Text(
                                    skill,
                                    fontSize = 14.sp,
                                    fontWeight = FontWeight.Medium,
                                    color = contentColor
                                )
                                if (isSelected) {
                                    Spacer(modifier = Modifier.width(6.dp))
                                    Icon(
                                        Icons.Default.Check,
                                        contentDescription = null,
                                        modifier = Modifier.size(16.dp),
                                        tint = contentColor
                                    )
                                }
                            }
                        }
                    }
                }
            }

            // Custom skills section
            item {
                Spacer(modifier = Modifier.height(16.dp))
                Text(
                    "Custom Skills",
                    fontWeight = FontWeight.Bold,
                    fontSize = 16.sp,
                    modifier = Modifier.padding(bottom = 8.dp)
                )
                Text(
                    "Add skills not in the list above",
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    fontSize = 13.sp,
                    modifier = Modifier.padding(bottom = 8.dp)
                )
            }

            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    OutlinedTextField(
                        value = customSkillText,
                        onValueChange = { customSkillText = it },
                        placeholder = { Text("e.g. Tile Work, Welding...") },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(12.dp),
                        singleLine = true,
                        keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done),
                        keyboardActions = KeyboardActions(
                            onDone = {
                                val trimmed = customSkillText.trim()
                                if (trimmed.isNotEmpty()) {
                                    selectedSkills = selectedSkills.toMutableSet().apply { add(trimmed) }
                                    customSkillText = ""
                                }
                            }
                        )
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    FilledIconButton(
                        onClick = {
                            val trimmed = customSkillText.trim()
                            if (trimmed.isNotEmpty()) {
                                selectedSkills = selectedSkills.toMutableSet().apply { add(trimmed) }
                                customSkillText = ""
                            }
                        }
                    ) {
                        Icon(Icons.Default.Add, contentDescription = "Add skill")
                    }
                }
            }

            // Show custom skills that are selected but not in predefined
            item {
                val customSelected = selectedSkills.filter { skill ->
                    predefinedSkills.none { it.second == skill }
                }
                if (customSelected.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(8.dp))
                    FlowRow(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        customSelected.forEach { skill ->
                            InputChip(
                                selected = true,
                                onClick = {
                                    selectedSkills = selectedSkills.toMutableSet().apply { remove(skill) }
                                },
                                label = { Text(skill) },
                                trailingIcon = {
                                    Icon(
                                        Icons.Default.Close,
                                        contentDescription = "Remove",
                                        modifier = Modifier.size(16.dp)
                                    )
                                }
                            )
                        }
                    }
                }
            }

            item {
                Spacer(modifier = Modifier.height(100.dp))
            }
        }
    }
}

@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.search

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.bero.data.models.ServiceCategory

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SearchScreen(
    onBackClick: () -> Unit = {},
    onCategoryClick: (ServiceCategory) -> Unit = {}
) {
    var query by remember { mutableStateOf("") }
    var active by remember { mutableStateOf(false) }

    val history = remember { mutableStateListOf("Plumber", "AC Repair", "Cleaner in Indiranagar") }
    val categories = ServiceCategory.entries

    Scaffold(
        topBar = {
            SearchBar(
                query = query,
                onQueryChange = { query = it },
                onSearch = { 
                    if (query.isNotBlank() && !history.contains(query)) {
                        history.add(0, query)
                    }
                    active = false 
                },
                active = true, // Force active for full screen search feel
                onActiveChange = { },
                placeholder = { Text("Search services...") },
                leadingIcon = {
                    if (active) {
                        IconButton(onClick = onBackClick) {
                            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                        }
                    } else {
                        Icon(Icons.Default.Search, contentDescription = "Search")
                    }
                },
                trailingIcon = {
                    if (query.isNotEmpty()) {
                        IconButton(onClick = { query = "" }) {
                            Icon(Icons.Default.Close, contentDescription = "Clear")
                        }
                    }
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp)
            ) {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    if (query.isEmpty()) {
                        // Recent Searches
                        if (history.isNotEmpty()) {
                            item {
                                Text(
                                    text = "Recent Searches",
                                    fontWeight = FontWeight.Bold,
                                    fontSize = 16.sp
                                )
                            }
                            items(history) { recent ->
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .clickable { query = recent }
                                        .padding(vertical = 12.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Icon(
                                        Icons.Default.History,
                                        contentDescription = null,
                                        tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                                    )
                                    Spacer(modifier = Modifier.width(16.dp))
                                    Text(
                                        text = recent,
                                        fontSize = 16.sp
                                    )
                                    Spacer(modifier = Modifier.weight(1f))
                                    IconButton(
                                        onClick = { history.remove(recent) },
                                        modifier = Modifier.size(24.dp)
                                    ) {
                                        Icon(
                                            Icons.Default.Close,
                                            contentDescription = "Remove",
                                            tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f)
                                        )
                                    }
                                }
                            }
                            item {
                                HorizontalDivider()
                            }
                        }

                        // Categories suggestion
                        item {
                            Text(
                                text = "Popular Categories",
                                fontWeight = FontWeight.Bold,
                                fontSize = 16.sp,
                                modifier = Modifier.padding(top = 8.dp)
                            )
                        }

                        items(categories.take(5)) { category ->
                             Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable { onCategoryClick(category) }
                                    .padding(vertical = 8.dp),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Box(
                                    modifier = Modifier
                                        .size(40.dp)
                                        .background(
                                            Color(category.color.toInt()).copy(alpha = 0.2f),
                                            RoundedCornerShape(8.dp)
                                        ),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Text(category.emoji, fontSize = 20.sp)
                                }
                                Spacer(modifier = Modifier.width(16.dp))
                                Text(
                                    text = category.displayName,
                                    fontSize = 16.sp
                                )
                            }
                        }
                    } else {
                        // Search Results (Filtered Categories for now)
                        val filtered = categories.filter { 
                            it.displayName.contains(query, ignoreCase = true) 
                        }
                        
                        if (filtered.isEmpty()) {
                             item {
                                Box(
                                    modifier = Modifier.fillMaxWidth().padding(32.dp),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                        Icon(
                                            Icons.Default.SearchOff,
                                            contentDescription = null,
                                            modifier = Modifier.size(48.dp),
                                            tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                                        )
                                        Spacer(modifier = Modifier.height(16.dp))
                                        Text("No results found for '$query'")
                                    }
                                }
                            }
                        } else {
                            items(filtered) { category ->
                                ListItem(
                                    headlineContent = { Text(category.displayName) },
                                    leadingContent = { Text(category.emoji, fontSize = 24.sp) },
                                    modifier = Modifier.clickable { onCategoryClick(category) }
                                )
                            }
                        }
                    }
                }
            }
        }
    ) {}
}

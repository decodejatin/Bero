package com.example.bero.utils

interface Platform {
    val name: String
}

class AndroidPlatform : Platform {
    override val name: String = "Android ${android.os.Build.VERSION.SDK_INT}"
}

fun getPlatform(): Platform = AndroidPlatform()

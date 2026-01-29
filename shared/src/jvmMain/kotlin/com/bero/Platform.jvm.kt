package com.bero

/**
 * JVM Platform implementation for testing
 */
class JvmPlatform : Platform {
    override val name: String = "JVM (Test)"
}

actual fun getPlatform(): Platform = JvmPlatform()

package org.latentsignal.sampleplugin

import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType
import com.intellij.openapi.diagnostic.thisLogger
import com.intellij.openapi.project.Project
import java.io.File
import java.net.ServerSocket
import java.nio.file.Files
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.atomic.AtomicReference

/**
 * PeriscopeProcessManager — lifecycle manager for the periscope binary.
 *
 * Layer 3 addition (diazMelgarejo/periscope).
 * Starts the periscope server on project open and stops it on project close.
 * Multiple projects share a single process; the last project to close stops it.
 *
 * Usage:
 *   PeriscopeProcessManager.start(project)   // from MyProjectActivity
 *   PeriscopeProcessManager.stop(project)    // from MyProjectActivity (on close)
 *   PeriscopeProcessManager.serverUrl()      // from MyToolWindowFactory
 */
object PeriscopeProcessManager {

    private val logger = thisLogger()

    // Default port periscope listens on.
    private const val DEFAULT_PORT = 8080

    // Number of open projects currently using the process.
    private val refCount = AtomicInteger(0)

    // Live process handle, if any.
    private val processRef = AtomicReference<Process?>(null)

    // Resolved port (may differ from DEFAULT_PORT if port was taken).
    @Volatile private var resolvedPort: Int = DEFAULT_PORT

    // ── Public API ────────────────────────────────────────────────────────────

    /**
     * Start the periscope server (if not already running) and increment
     * the reference count for [project].
     */
    fun start(project: Project) {
        refCount.incrementAndGet()
        if (processRef.get() != null) {
            logger.info("periscope: already running on port $resolvedPort")
            return
        }
        synchronized(this) {
            if (processRef.get() != null) return  // double-checked
            launchProcess(project)
        }
    }

    /**
     * Decrement the reference count for [project]. When all projects are
     * closed the periscope process is stopped.
     */
    fun stop(project: Project) {
        val remaining = refCount.decrementAndGet()
        if (remaining <= 0) {
            refCount.set(0)
            stopProcess()
        }
    }

    /**
     * Returns the URL at which the periscope UI is served.
     * Uses the resolved port after [start] is called; falls back to default.
     */
    fun serverUrl(): String = "http://localhost:$resolvedPort/"

    // ── Internal ──────────────────────────────────────────────────────────────

    private fun launchProcess(project: Project) {
        val binary = resolveBinary()
        if (binary == null) {
            notify(
                project,
                "Periscope binary not found. Install periscope and ensure it is on your PATH.",
                NotificationType.WARNING
            )
            return
        }

        resolvedPort = findFreePort(DEFAULT_PORT)

        val cmd = listOf(binary.absolutePath, "serve", "--port", resolvedPort.toString())
        logger.info("periscope: launching ${cmd.joinToString(" ")}")

        try {
            val process = ProcessBuilder(cmd)
                .redirectErrorStream(true)
                .start()

            processRef.set(process)

            // Drain stdout/stderr so the process does not block on a full pipe.
            Thread {
                process.inputStream.bufferedReader().use { reader ->
                    reader.lines().forEach { line ->
                        logger.debug("periscope: $line")
                    }
                }
                // Process exited — clear the reference.
                processRef.set(null)
                logger.info("periscope: process exited (port $resolvedPort)")
            }.also {
                it.isDaemon = true
                it.name = "periscope-stdout-drain"
                it.start()
            }

            logger.info("periscope: started on port $resolvedPort (pid ${process.pid()})")
            notify(project, "Periscope started on port $resolvedPort", NotificationType.INFORMATION)

        } catch (e: Exception) {
            logger.warn("periscope: failed to start", e)
            processRef.set(null)
            notify(project, "Failed to start periscope: ${e.message}", NotificationType.ERROR)
        }
    }

    private fun stopProcess() {
        val process = processRef.getAndSet(null) ?: return
        logger.info("periscope: stopping process")
        process.destroy()
        // Give it a moment to exit gracefully; force-kill if needed.
        if (!process.waitFor(3, java.util.concurrent.TimeUnit.SECONDS)) {
            process.destroyForcibly()
        }
        logger.info("periscope: stopped")
    }

    /**
     * Locate the periscope binary on the system.
     * Search order: PATH, ~/periscope, ~/.local/bin/periscope, /usr/local/bin/periscope.
     */
    private fun resolveBinary(): File? {
        // Candidates in preference order
        val name = if (System.getProperty("os.name").lowercase().contains("win")) "periscope.exe" else "periscope"
        val homeDir = System.getProperty("user.home")
        val candidates = listOf(
            File(homeDir, name),
            File(homeDir, ".local/bin/$name"),
            File("/usr/local/bin/$name"),
            File("/opt/homebrew/bin/$name"),
        )
        // Prefer PATH first
        val fromPath = findOnPath(name)
        if (fromPath != null) return fromPath
        return candidates.firstOrNull { it.exists() && it.canExecute() }
    }

    private fun findOnPath(name: String): File? {
        val path = System.getenv("PATH") ?: return null
        return path.split(File.pathSeparator)
            .map { File(it, name) }
            .firstOrNull { it.exists() && it.canExecute() }
    }

    /**
     * Find a free TCP port, starting from [preferred].
     * Returns [preferred] if available, otherwise the next free port.
     */
    private fun findFreePort(preferred: Int): Int {
        return try {
            ServerSocket(preferred).use { preferred }
        } catch (_: Exception) {
            ServerSocket(0).use { it.localPort }
        }
    }

    private fun notify(project: Project, message: String, type: NotificationType) {
        try {
            NotificationGroupManager.getInstance()
                .getNotificationGroup("Periscope")
                .createNotification(message, type)
                .notify(project)
        } catch (e: Exception) {
            // NotificationGroup may not be registered in older IDE versions — log only.
            logger.info("periscope notification ($type): $message")
        }
    }
}

package org.latentsignal.sampleplugin.startup

import com.intellij.openapi.diagnostic.thisLogger
import com.intellij.openapi.project.Project
import com.intellij.openapi.project.ProjectManager
import com.intellij.openapi.project.ProjectManagerListener
import com.intellij.openapi.startup.ProjectActivity
import org.latentsignal.sampleplugin.PeriscopeProcessManager

/**
 * MyProjectActivity starts the periscope binary when a project opens and
 * registers a listener to stop it when the project closes.
 *
 * Layer 3 addition (diazMelgarejo/periscope).
 * Updated for IntelliJ Platform 2025.1: ProjectActivity moved to
 * com.intellij.openapi.startup.ProjectActivity; TOPIC on ProjectManager.
 */
class MyProjectActivity : ProjectActivity {

    override suspend fun execute(project: Project) {
        thisLogger().info("Periscope: project opened — starting server")
        PeriscopeProcessManager.start(project)

        // Register shutdown hook for this project.
        project.messageBus.connect().subscribe(
            ProjectManager.TOPIC,
            object : ProjectManagerListener {
                override fun projectClosing(closingProject: Project) {
                    if (closingProject === project) {
                        thisLogger().info("Periscope: project closing — stopping server")
                        PeriscopeProcessManager.stop(project)
                    }
                }
            }
        )
    }
}

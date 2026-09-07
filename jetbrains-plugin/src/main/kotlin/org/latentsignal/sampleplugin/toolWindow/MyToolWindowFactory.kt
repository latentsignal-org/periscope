package org.latentsignal.sampleplugin.toolWindow

import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory
import com.intellij.ui.content.ContentFactory
import com.intellij.ui.jcef.JBCefBrowser
import org.latentsignal.sampleplugin.PeriscopeProcessManager
import java.awt.BorderLayout
import javax.swing.JPanel


class MyToolWindowFactory : ToolWindowFactory {

    override fun createToolWindowContent(project: Project, toolWindow: ToolWindow) {
        val panel = JPanel(BorderLayout())
        // Use the URL from PeriscopeProcessManager so the port is always correct
        // (may differ from 8080 if the default port was already in use).
        val browser = JBCefBrowser(PeriscopeProcessManager.serverUrl())
        panel.add(browser.component, BorderLayout.CENTER)

        val content = ContentFactory.getInstance().createContent(panel, null, false)
        toolWindow.contentManager.addContent(content)
    }

    override fun shouldBeAvailable(project: Project) = true
}

package org.latentsignal.sampleplugin

import org.junit.Assert.assertEquals
import org.junit.Test

class PeriscopeProcessManagerTest {

    @Test
    fun serverUrlUsesDefaultPortBeforeStart() {
        assertEquals("http://localhost:8080/", PeriscopeProcessManager.serverUrl())
    }

    @Test
    fun binaryNamesPreferPeriscopeOverAgentsview() {
        val names = PeriscopeProcessManager.binaryNames()
        assertEquals("periscope", names.first())
        assertEquals("agentsview", names.last())
    }
}

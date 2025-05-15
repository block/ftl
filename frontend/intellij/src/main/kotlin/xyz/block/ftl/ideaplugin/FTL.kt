package xyz.block.ftl.ideaplugin

import com.intellij.openapi.application.ApplicationManager

fun runOnEDT(runnable: () -> Unit) {
  ApplicationManager.getApplication().invokeLater {
    runnable()
  }
}

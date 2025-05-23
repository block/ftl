package xyz.block.ftl.ideaplugin

import com.intellij.openapi.options.Configurable
import org.jetbrains.annotations.Nls
import javax.swing.JComponent

class FTLSettingsConfigurable : Configurable {

  private var mySettingsComponent: FTLSettingsComponent? = null

  @Nls(capitalization = Nls.Capitalization.Title)
  override fun getDisplayName(): String {
    return "FTL"
  }

  override fun getPreferredFocusedComponent(): JComponent? {
    return mySettingsComponent?.getPreferredFocusedComponent()
  }

  override fun createComponent(): JComponent? {
    mySettingsComponent = FTLSettingsComponent()
    return mySettingsComponent?.getPanel()
  }

  override fun isModified(): Boolean {
    val state = AppSettings.getInstance().state
    return mySettingsComponent?.getLspServerPath() != state.lspServerPath
  }

  override fun apply() {
    val state = AppSettings.getInstance().state
    state.lspServerPath = mySettingsComponent?.getLspServerPath() ?: "ftl"
  }

  override fun reset() {
    val state = AppSettings.getInstance().state
    mySettingsComponent?.setLspServerPath(state.lspServerPath)
  }

  override fun disposeUIResources() {
    mySettingsComponent = null
  }
}

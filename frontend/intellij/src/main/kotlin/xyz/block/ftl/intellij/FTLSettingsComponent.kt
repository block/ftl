package xyz.block.ftl.intellij

import com.intellij.openapi.ui.TextFieldWithBrowseButton
import com.intellij.ui.components.JBLabel
import com.intellij.util.ui.FormBuilder
import javax.swing.JComponent
import javax.swing.JPanel

class FTLSettingsComponent {

  private val settingsPanel: JPanel
  private val lspServerPath = TextFieldWithBrowseButton()

  init {
    settingsPanel = FormBuilder.createFormBuilder()
      .addLabeledComponent(JBLabel("LSP Server Path:"), lspServerPath, 1, false)
      .addComponentFillVertically(JPanel(), 0)
      .panel
  }

  fun getPanel(): JPanel {
    return settingsPanel
  }

  fun getPreferredFocusedComponent(): JComponent {
    return lspServerPath
  }

  fun getLspServerPath(): String {
    return lspServerPath.text
  }

  fun setLspServerPath(newPath: String) {
    lspServerPath.text = newPath
  }
}

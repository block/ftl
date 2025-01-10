import type { ExtensionContext } from 'vscode'

import * as vscode from 'vscode'
import { FTLClient } from './client'
import { MIN_FTL_VERSION, checkMinimumVersion, getFTLVersion, getProjectOrWorkspaceRoot, resolveFtlPath } from './config'
import { FTLStatus } from './status'

const extensionId = 'ftl'
let client: FTLClient
let statusBarItem: vscode.StatusBarItem
let outputChannel: vscode.OutputChannel

export const activate = async (context: ExtensionContext) => {
  outputChannel = vscode.window.createOutputChannel('FTL', 'log')
  outputChannel.appendLine('FTL extension activated')

  statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100)
  statusBarItem.command = 'ftl.statusItemClicked'
  statusBarItem.show()

  client = new FTLClient(statusBarItem, outputChannel)

  const showLogsCommand = vscode.commands.registerCommand('ftl.showLogs', () => {
    outputChannel.show()
  })

  const showCommands = vscode.commands.registerCommand('ftl.statusItemClicked', () => {
    const ftlCommands = [{ label: 'FTL: Show Logs', command: 'ftl.showLogs' }]

    vscode.window.showQuickPick(ftlCommands, { placeHolder: 'Select an FTL command' }).then((selected) => {
      if (selected) {
        vscode.commands.executeCommand(selected.command)
      }
    })
  })

  await startClient(context)

  context.subscriptions.push(statusBarItem, showCommands, showLogsCommand)
}

export const deactivate = async () => client.stop()

const FTLPreflightCheck = async (ftlPath: string) => {
  let version: string
  try {
    version = await getFTLVersion(ftlPath)
    // biome-ignore lint/suspicious/noExplicitAny: <explanation>
  } catch (error: any) {
    vscode.window.showErrorMessage(`${error.message}`)
    return false
  }

  const versionOK = checkMinimumVersion(version, MIN_FTL_VERSION)
  if (!versionOK) {
    vscode.window.showErrorMessage(`FTL version ${version} is not supported. Please upgrade to at least ${MIN_FTL_VERSION}.`)
    return false
  }

  return true
}

const startClient = async (context: ExtensionContext) => {
  FTLStatus.ftlStarting(statusBarItem)

  const ftlConfig = vscode.workspace.getConfiguration('ftl')
  const workspaceRootPath = await getProjectOrWorkspaceRoot()
  const resolvedFtlPath = await resolveFtlPath(workspaceRootPath, ftlConfig)

  outputChannel.appendLine(`VSCode workspace root path: ${workspaceRootPath}`)
  outputChannel.appendLine(`FTL path: ${resolvedFtlPath}`)

  const ftlOK = await FTLPreflightCheck(resolvedFtlPath)
  if (!ftlOK) {
    FTLStatus.ftlStopped(statusBarItem)
    return
  }

  const userFlags = ftlConfig.get<string[]>('lspCommandFlags') ?? []

  const flags = [...userFlags]

  return client.start(resolvedFtlPath, workspaceRootPath, flags, context)
}

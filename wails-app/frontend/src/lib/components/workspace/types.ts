export interface FileNode {
  name: string;
  path: string;
  isDirectory: boolean;
  children?: FileNode[];
  isExpanded?: boolean;
}

export interface WorkspaceTab {
  name: string;
  path: string;
}

export interface IDEShellState {
  isExplorerOpen: boolean;
  isBottomPanelOpen: boolean;
  isCopilotOpen: boolean;
}

export interface IDEStatus {
  left: {
    workspaceName?: string;
    sttStatus?: string;
    docStats?: string;
  };
  right: {
    scrollerStatus?: string;
    ragStatus?: string;
    stealthStatus?: string;
  };
}

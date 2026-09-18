export interface HeaderAction {
  icon: string;          // Material Symbols icon name, e.g. 'star', 'download'
  label: string;         // Tooltip text shown on hover
  onClick: () => void;   // Callback when button is clicked
  active?: boolean;      // Optional: highlights button when true (like toggles)
  activeIcon?: string;   // Optional: alternate icon when active=true
}

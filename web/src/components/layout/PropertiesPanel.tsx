import { Sidebar, SidebarSection } from './Sidebar';

interface PropertiesPanelProps {
  children?: React.ReactNode;
}

export function PropertiesPanel({ children }: PropertiesPanelProps) {
  return (
    <Sidebar side="right" width="w-64">
      {children ?? (
        <SidebarSection title="Properties">
          <p className="text-petit text-ink-muted">
            Select a node to edit its properties
          </p>
        </SidebarSection>
      )}
    </Sidebar>
  );
}

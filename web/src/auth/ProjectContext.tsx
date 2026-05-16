import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from 'react';

import { getStoredProjectCode, setStoredProjectCode } from '../api/client';
import { listProjects } from '../api/project';
import type { Project } from '../types/project';

interface ProjectContextValue {
  projects: Project[];
  currentProject: Project | null;
  loading: boolean;
  error: string | null;
  selectProject: (code: string) => void;
  reloadProjects: () => Promise<void>;
}

const ProjectContext = createContext<ProjectContextValue | null>(null);

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [currentCode, setCurrentCode] = useState<string | null>(() => getStoredProjectCode());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const currentProject = useMemo(
    () => projects.find((project) => project.code === currentCode) ?? null,
    [currentCode, projects],
  );

  async function reloadProjects() {
    setLoading(true);
    setError(null);
    try {
      const result = await listProjects();
      setProjects(result.items);
      const stored = getStoredProjectCode();
      const next = result.items.find((project) => project.code === stored) ?? result.items[0] ?? null;
      const nextCode = next?.code ?? null;
      setCurrentCode(nextCode);
      setStoredProjectCode(nextCode);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void reloadProjects();
  }, []);

  function selectProject(code: string) {
    setCurrentCode(code);
    setStoredProjectCode(code);
  }

  return (
    <ProjectContext.Provider value={{ projects, currentProject, loading, error, selectProject, reloadProjects }}>
      {children}
    </ProjectContext.Provider>
  );
}

export function useProject() {
  const value = useContext(ProjectContext);
  if (!value) {
    throw new Error('useProject must be used inside ProjectProvider');
  }
  return value;
}

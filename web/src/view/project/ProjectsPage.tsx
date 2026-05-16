import { FormEvent, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import { createProject, deleteProject, listProjects, updateProject } from '../../api/project';
import { useProject } from '../../auth/ProjectContext';
import { useI18n } from '../../locale/useI18n';
import type { Project } from '../../types/project';
import styles from './ProjectsPage.module.css';

interface Draft {
  name: string;
  code: string;
  description: string;
}

const emptyDraft: Draft = { name: '', code: '', description: '' };

export function ProjectsPage() {
  const { t } = useI18n();
  const { currentProject, reloadProjects } = useProject();
  const [projects, setProjects] = useState<Project[]>([]);
  const [draft, setDraft] = useState<Draft>(emptyDraft);
  const [editing, setEditing] = useState<Record<string, Draft>>({});
  const [error, setError] = useState<string | null>(null);

  async function load() {
    const result = await listProjects();
    setProjects(result.items);
  }

  useEffect(() => {
    void load().catch((err) => setError(err instanceof Error ? err.message : t('project.list.loadFailed')));
  }, []);

  async function handleCreate(event: FormEvent) {
    event.preventDefault();
    setError(null);
    try {
      await createProject(draft);
      setDraft(emptyDraft);
      await load();
      await reloadProjects();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('project.list.saveFailed'));
    }
  }

  async function handleUpdate(project: Project) {
    const next = editing[project.code];
    if (!next) {
      return;
    }
    setError(null);
    try {
      await updateProject(project.code, { name: next.name, description: next.description });
      await load();
      await reloadProjects();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('project.list.saveFailed'));
    }
  }

  async function handleDelete(project: Project) {
    if (!window.confirm(t('project.list.deleteConfirm'))) {
      return;
    }
    setError(null);
    try {
      await deleteProject(project.code);
      await load();
      await reloadProjects();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('project.list.deleteFailed'));
    }
  }

  const activeCount = projects.filter((project) => project.status === 'active').length;

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div>
          <p className="page-eyebrow">{t('project.list.eyebrow')}</p>
          <h1 className={styles.title}>{t('project.list.title')}</h1>
        </div>
      </header>

      <div className={styles.summary}>
        <div>
          <span>{t('project.list.total')}</span>
          <strong>{projects.length}</strong>
        </div>
        <div>
          <span>{t('project.list.active')}</span>
          <strong>{activeCount}</strong>
        </div>
        <div className={styles.currentSummary}>
          <span>{t('project.list.current')}</span>
          <strong>{currentProject?.name ?? t('project.switcher.empty')}</strong>
        </div>
      </div>

      {error ? <div className={styles.error}>{error}</div> : null}

      <form className={styles.createBar} onSubmit={(event) => void handleCreate(event)}>
        <label>
          <span>{t('project.list.name')}</span>
          <input
            value={draft.name}
            onChange={(event) => setDraft((value) => ({ ...value, name: event.target.value }))}
            placeholder={t('project.list.namePlaceholder')}
            required
          />
        </label>
        <label>
          <span>{t('project.list.code')}</span>
          <input
            value={draft.code}
            onChange={(event) => setDraft((value) => ({ ...value, code: event.target.value }))}
            placeholder={t('project.list.codePlaceholder')}
            required
          />
        </label>
        <label>
          <span>{t('project.list.description')}</span>
          <input
            value={draft.description}
            onChange={(event) => setDraft((value) => ({ ...value, description: event.target.value }))}
            placeholder={t('project.list.descriptionPlaceholder')}
          />
        </label>
        <button type="submit">{t('project.list.create')}</button>
      </form>

      <div className={styles.table}>
        <div className={styles.tableHeader} aria-hidden="true">
          <span>{t('project.list.project')}</span>
          <span>{t('project.list.name')}</span>
          <span>{t('project.list.description')}</span>
          <span>{t('project.list.actions')}</span>
        </div>
        {projects.map((project) => {
          const row = editing[project.code] ?? {
            name: project.name,
            code: project.code,
            description: project.description,
          };
          return (
            <article className={styles.row} key={project.code}>
              <div className={styles.identity}>
                <strong>{project.name}</strong>
                <span>{project.code}</span>
                <em className={project.status === 'active' ? styles.statusActive : undefined}>{project.status}</em>
              </div>
              <label className={styles.field}>
                <span>{t('project.list.name')}</span>
                <input
                  value={row.name}
                  onChange={(event) =>
                    setEditing((value) => ({
                      ...value,
                      [project.code]: { ...row, name: event.target.value },
                    }))
                  }
                  aria-label={t('project.list.name')}
                />
              </label>
              <label className={styles.field}>
                <span>{t('project.list.description')}</span>
                <input
                  value={row.description}
                  onChange={(event) =>
                    setEditing((value) => ({
                      ...value,
                      [project.code]: { ...row, description: event.target.value },
                    }))
                  }
                  aria-label={t('project.list.description')}
                />
              </label>
              <div className={styles.actions}>
                <button type="button" className={styles.primary} onClick={() => void handleUpdate(project)}>
                  {t('project.list.save')}
                </button>
                <Link to={`/projects/${project.code}/members`}>{t('project.list.members')}</Link>
                <button type="button" className={styles.danger} onClick={() => void handleDelete(project)}>
                  {t('project.list.delete')}
                </button>
              </div>
            </article>
          );
        })}
        {projects.length === 0 ? <p className={styles.empty}>{t('project.list.empty')}</p> : null}
      </div>
    </section>
  );
}

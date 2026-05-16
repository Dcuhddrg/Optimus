import { Link } from 'react-router-dom';

import { useProject } from '../../auth/ProjectContext';
import { useI18n } from '../../locale/useI18n';
import styles from './ProjectSwitcher.module.css';

export function ProjectSwitcher() {
  const { projects, currentProject, loading, error, selectProject } = useProject();
  const { t } = useI18n();

  return (
    <div className={styles.switcher} aria-label={t('project.switcher.label')}>
      <div className={styles.context}>
        <span className={styles.label}>{t('project.switcher.current')}</span>
        <strong>{currentProject?.name ?? t('project.switcher.empty')}</strong>
        {currentProject ? <span className={styles.code}>{currentProject.code}</span> : null}
      </div>
      <select
        className={styles.select}
        value={currentProject?.code ?? ''}
        disabled={loading || projects.length === 0}
        onChange={(event) => selectProject(event.target.value)}
      >
        {projects.length === 0 ? (
          <option value="">{loading ? t('project.switcher.loading') : t('project.switcher.empty')}</option>
        ) : (
          projects.map((project) => (
            <option key={project.code} value={project.code}>
              {project.name} ({project.code})
            </option>
          ))
        )}
      </select>
      <Link className={styles.manage} to="/projects">
        {t('project.switcher.manage')}
      </Link>
      <span className={styles.count}>
        {projects.length} {t('project.switcher.count')}
      </span>
      {error ? <span className={styles.error}>{error}</span> : null}
    </div>
  );
}

import { KeyboardEvent, useEffect, useMemo, useRef, useState } from 'react';
import { Link } from 'react-router-dom';

import { useProject } from '../../auth/ProjectContext';
import { useI18n } from '../../locale/useI18n';
import styles from './ProjectSwitcher.module.css';

export function ProjectSwitcher() {
  const { projects, currentProject, loading, error, selectProject } = useProject();
  const { t } = useI18n();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) {
      return projects;
    }
    return projects.filter(
      (project) =>
        project.name.toLowerCase().includes(q) || project.code.toLowerCase().includes(q),
    );
  }, [projects, query]);

  useEffect(() => {
    if (!open) {
      setQuery('');
      return;
    }
    requestAnimationFrame(() => searchInputRef.current?.focus());

    function onDocumentMouseDown(event: MouseEvent) {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }
    function onKey(event: globalThis.KeyboardEvent) {
      if (event.key === 'Escape') {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', onDocumentMouseDown);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDocumentMouseDown);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  function handleSelect(code: string) {
    selectProject(code);
    setOpen(false);
  }

  function handleSearchKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Enter' && filtered.length > 0) {
      event.preventDefault();
      handleSelect(filtered[0].code);
    }
  }

  const triggerLabel = currentProject?.name ?? t('project.switcher.empty');
  const triggerSub = currentProject?.code ?? '';
  const disabled = loading && projects.length === 0;

  return (
    <div ref={containerRef} className={styles.switcher} aria-label={t('project.switcher.label')}>
      <div className={styles.context}>
        <span className={styles.label}>{t('project.switcher.current')}</span>
        {triggerSub ? <span className={styles.code}>{triggerSub}</span> : null}
        <strong>{triggerLabel}</strong>
      </div>
      <button
        type="button"
        className={styles.trigger}
        onClick={() => setOpen((value) => !value)}
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={open}
      >
        <span className={styles.triggerName}>
          {loading && projects.length === 0 ? t('project.switcher.loading') : triggerLabel}
        </span>
        <span className={styles.chevron} aria-hidden="true">{open ? '▴' : '▾'}</span>
      </button>
      <span className={styles.count}>
        {projects.length} {t('project.switcher.count')}
      </span>
      {error ? <span className={styles.error}>{error}</span> : null}

      {open ? (
        <div className={styles.panel} role="listbox">
          <div className={styles.searchRow}>
            <input
              ref={searchInputRef}
              type="search"
              className={styles.search}
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={handleSearchKeyDown}
              placeholder={t('project.switcher.search')}
              aria-label={t('project.switcher.search')}
            />
          </div>
          <ul className={styles.list}>
            {filtered.length === 0 ? (
              <li className={styles.emptyItem}>{t('project.switcher.noResults')}</li>
            ) : (
              filtered.map((project) => {
                const active = project.code === currentProject?.code;
                return (
                  <li key={project.code}>
                    <button
                      type="button"
                      role="option"
                      aria-selected={active}
                      className={active ? styles.itemActive : styles.item}
                      onClick={() => handleSelect(project.code)}
                    >
                      <span className={styles.itemName}>{project.name}</span>
                      <span className={styles.itemCode}>{project.code}</span>
                      {active ? <span className={styles.itemCheck} aria-hidden="true">✓</span> : null}
                    </button>
                  </li>
                );
              })
            )}
          </ul>
          <Link className={styles.footer} to="/projects" onClick={() => setOpen(false)}>
            + {t('project.switcher.manage')}
          </Link>
        </div>
      ) : null}
    </div>
  );
}

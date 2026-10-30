import { FormEvent, useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

import {
  addProjectMember,
  listProjectMembers,
  removeProjectMember,
  updateProjectMemberRole,
} from '../../api/project';
import { listUsers } from '../../api/system';
import { useI18n } from '../../locale/useI18n';
import type { ProjectMember, ProjectRole } from '../../types/project';
import type { SystemUser } from '../../types/system';
import styles from './ProjectMembersPage.module.css';

const roles: ProjectRole[] = ['owner', 'member', 'viewer'];

export function ProjectMembersPage() {
  const { code = '' } = useParams();
  const { t } = useI18n();
  const [members, setMembers] = useState<ProjectMember[]>([]);
  const [users, setUsers] = useState<SystemUser[]>([]);
  const [usersUnavailable, setUsersUnavailable] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);
  const [userSearch, setUserSearch] = useState('');
  const [role, setRole] = useState<ProjectRole>('member');
  const [error, setError] = useState<string | null>(null);

  async function load() {
    if (!code) {
      return;
    }
    const result = await listProjectMembers(code);
    setMembers(result.items);
  }

  async function loadUsers() {
    try {
      const result = await listUsers();
      setUsers(result.list);
      setUsersUnavailable(false);
    } catch {
      setUsers([]);
      setUsersUnavailable(true);
    }
  }

  useEffect(() => {
    void load().catch((err) => setError(err instanceof Error ? err.message : t('project.members.loadFailed')));
    void loadUsers();
  }, [code]);

  const eligibleUsers = useMemo(() => {
    const memberIds = new Set(members.map((member) => member.userId));
    const query = userSearch.trim().toLowerCase();
    return users.filter((user) => {
      if (memberIds.has(user.id)) {
        return false;
      }
      if (!query) {
        return true;
      }
      return (
        user.username.toLowerCase().includes(query) ||
        (user.displayName ?? '').toLowerCase().includes(query)
      );
    });
  }, [users, members, userSearch]);

  useEffect(() => {
    if (selectedUserId !== null && !eligibleUsers.some((user) => user.id === selectedUserId)) {
      setSelectedUserId(null);
    }
  }, [eligibleUsers, selectedUserId]);

  async function handleAdd(event: FormEvent) {
    event.preventDefault();
    if (selectedUserId === null) {
      setError(t('project.members.selectUserRequired'));
      return;
    }
    setError(null);
    try {
      await addProjectMember(code, { userId: selectedUserId, projectRole: role });
      setSelectedUserId(null);
      setUserSearch('');
      setRole('member');
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('project.members.saveFailed'));
    }
  }

  async function handleRoleChange(member: ProjectMember, nextRole: ProjectRole) {
    setError(null);
    try {
      await updateProjectMemberRole(code, member.userId, { projectRole: nextRole });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('project.members.saveFailed'));
    }
  }

  async function handleRemove(member: ProjectMember) {
    if (!window.confirm(t('project.members.removeConfirm'))) {
      return;
    }
    setError(null);
    try {
      await removeProjectMember(code, member.userId);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('project.members.removeFailed'));
    }
  }

  const ownerCount = members.filter((member) => member.projectRole === 'owner').length;
  const viewerCount = members.filter((member) => member.projectRole === 'viewer').length;

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div>
          <p className="page-eyebrow">{t('project.members.eyebrow')}</p>
          <h1 className={styles.title}>{t('project.members.title')} · {code}</h1>
        </div>
        <Link className={styles.back} to="/projects">{t('project.members.back')}</Link>
      </header>

      {error ? <div className={styles.error}>{error}</div> : null}

      <div className={styles.summary}>
        <div>
          <span>{t('project.members.total')}</span>
          <strong>{members.length}</strong>
        </div>
        <div>
          <span>{t('project.roles.owner')}</span>
          <strong>{ownerCount}</strong>
        </div>
        <div>
          <span>{t('project.roles.viewer')}</span>
          <strong>{viewerCount}</strong>
        </div>
      </div>

      <form className={styles.addBar} onSubmit={(event) => void handleAdd(event)}>
        <label className={styles.userPicker}>
          <span>{t('project.members.selectUser')}</span>
          <input
            value={userSearch}
            onChange={(event) => setUserSearch(event.target.value)}
            placeholder={t('project.members.searchUser')}
            disabled={usersUnavailable}
          />
          <select
            value={selectedUserId ?? ''}
            onChange={(event) =>
              setSelectedUserId(event.target.value ? Number(event.target.value) : null)
            }
            disabled={usersUnavailable || eligibleUsers.length === 0}
          >
            <option value="">
              {usersUnavailable
                ? t('project.members.usersUnavailable')
                : eligibleUsers.length === 0
                  ? t('project.members.noEligibleUsers')
                  : t('project.members.selectUserPlaceholder')}
            </option>
            {eligibleUsers.map((user) => (
              <option key={user.id} value={user.id}>
                {user.displayName ? `${user.displayName} · ${user.username}` : user.username} · #{user.id}
              </option>
            ))}
          </select>
        </label>
        <select value={role} onChange={(event) => setRole(event.target.value as ProjectRole)}>
          {roles.map((item) => (
            <option key={item} value={item}>{t(`project.roles.${item}`)}</option>
          ))}
        </select>
        <button type="submit" disabled={selectedUserId === null}>{t('project.members.add')}</button>
      </form>

      <div className={styles.list}>
        <div className={styles.listHeader} aria-hidden="true">
          <span>{t('project.members.user')}</span>
          <span>{t('project.members.currentRole')}</span>
          <span>{t('project.members.changeRole')}</span>
          <span>{t('project.list.actions')}</span>
        </div>
        {members.map((member) => (
          <article className={styles.row} key={member.userId}>
            <div className={styles.identity}>
              <strong>{member.displayName || member.username}</strong>
              <span>{member.username} · #{member.userId}</span>
            </div>
            <span className={styles.roleBadge}>{t(`project.roles.${member.projectRole as ProjectRole}`)}</span>
            <label className={styles.roleField}>
              <span>{t('project.members.changeRole')}</span>
              <select
                value={member.projectRole}
                onChange={(event) => void handleRoleChange(member, event.target.value as ProjectRole)}
              >
                {roles.map((item) => (
                  <option key={item} value={item}>{t(`project.roles.${item}`)}</option>
                ))}
              </select>
            </label>
            <button type="button" className={styles.danger} onClick={() => void handleRemove(member)}>
              {t('project.members.remove')}
            </button>
          </article>
        ))}
        {members.length === 0 ? <p className={styles.empty}>{t('project.members.empty')}</p> : null}
      </div>
    </section>
  );
}

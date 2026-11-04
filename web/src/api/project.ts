import { apiRequest } from './client';
import type {
  AddProjectMemberPayload,
  CreateProjectPayload,
  Project,
  ProjectMember,
  UpdateProjectMemberRolePayload,
  UpdateProjectPayload,
} from '../types/project';

type ProjectList<T> = { list: T[] };

export function listProjects() {
  return apiRequest<ProjectList<Project>>('/projects');
}

export function createProject(payload: CreateProjectPayload) {
  return apiRequest<Project>('/projects', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateProject(code: string, payload: UpdateProjectPayload) {
  return apiRequest<Project>(`/projects/${code}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteProject(code: string) {
  return apiRequest<Record<string, never>>(`/projects/${code}`, { method: 'DELETE' });
}

export function listProjectMembers(code: string) {
  return apiRequest<ProjectList<ProjectMember>>(`/projects/${code}/members`);
}

export function addProjectMember(code: string, payload: AddProjectMemberPayload) {
  return apiRequest<Record<string, never>>(`/projects/${code}/members`, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateProjectMemberRole(
  code: string,
  userId: number,
  payload: UpdateProjectMemberRolePayload,
) {
  return apiRequest<Record<string, never>>(`/projects/${code}/members/${userId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
}

export function removeProjectMember(code: string, userId: number) {
  return apiRequest<Record<string, never>>(`/projects/${code}/members/${userId}`, {
    method: 'DELETE',
  });
}

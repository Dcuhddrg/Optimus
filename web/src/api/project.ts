import { apiRequestEnvelope } from './client';
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
  return apiRequestEnvelope<ProjectList<Project>>('/projects');
}

export function createProject(payload: CreateProjectPayload) {
  return apiRequestEnvelope<Project>('/projects', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateProject(code: string, payload: UpdateProjectPayload) {
  return apiRequestEnvelope<Project>(`/projects/${code}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteProject(code: string) {
  return apiRequestEnvelope<Record<string, never>>(`/projects/${code}`, { method: 'DELETE' });
}

export function listProjectMembers(code: string) {
  return apiRequestEnvelope<ProjectList<ProjectMember>>(`/projects/${code}/members`);
}

export function addProjectMember(code: string, payload: AddProjectMemberPayload) {
  return apiRequestEnvelope<Record<string, never>>(`/projects/${code}/members`, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateProjectMemberRole(
  code: string,
  userId: number,
  payload: UpdateProjectMemberRolePayload,
) {
  return apiRequestEnvelope<Record<string, never>>(`/projects/${code}/members/${userId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
}

export function removeProjectMember(code: string, userId: number) {
  return apiRequestEnvelope<Record<string, never>>(`/projects/${code}/members/${userId}`, {
    method: 'DELETE',
  });
}

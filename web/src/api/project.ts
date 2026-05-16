import { apiRequest } from './client';
import type {
  AddProjectMemberPayload,
  CreateProjectPayload,
  Project,
  ProjectMember,
  UpdateProjectMemberRolePayload,
  UpdateProjectPayload,
} from '../types/project';
import type { ListResponse } from '../types/system';

export function listProjects() {
  return apiRequest<ListResponse<Project>>('/projects');
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
  return apiRequest<{ message: string }>(`/projects/${code}`, { method: 'DELETE' });
}

export function listProjectMembers(code: string) {
  return apiRequest<ListResponse<ProjectMember>>(`/projects/${code}/members`);
}

export function addProjectMember(code: string, payload: AddProjectMemberPayload) {
  return apiRequest<{ message: string }>(`/projects/${code}/members`, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateProjectMemberRole(
  code: string,
  userId: number,
  payload: UpdateProjectMemberRolePayload,
) {
  return apiRequest<{ message: string }>(`/projects/${code}/members/${userId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
}

export function removeProjectMember(code: string, userId: number) {
  return apiRequest<{ message: string }>(`/projects/${code}/members/${userId}`, {
    method: 'DELETE',
  });
}

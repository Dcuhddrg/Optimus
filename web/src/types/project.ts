export type ProjectRole = 'owner' | 'member' | 'viewer';

export interface Project {
  id: number;
  name: string;
  code: string;
  description: string;
  ownerId: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectMember {
  userId: number;
  username: string;
  displayName: string;
  projectRole: ProjectRole | string;
}

export interface CreateProjectPayload {
  name: string;
  code: string;
  description: string;
}

export interface UpdateProjectPayload {
  name?: string;
  description?: string;
}

export interface AddProjectMemberPayload {
  userId: number;
  projectRole: ProjectRole;
}

export interface UpdateProjectMemberRolePayload {
  projectRole: ProjectRole;
}

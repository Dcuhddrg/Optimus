export interface CloudAccount {
  id: number;
  projectId: number;
  name: string;
  provider: string;
  accountId: string;
  credentialType: string;
  credential: string;
  tags: Record<string, string>;
  status: string;
  lastSyncedAt: string | null;
}

export interface CreateCloudAccountPayload {
  name: string;
  provider: string;
  accountId: string;
  credentialType: string;
  credential: Record<string, string>;
  tags?: Record<string, string>;
}

export interface UpdateCloudAccountTagsPayload {
  tags: Record<string, string>;
}

export interface CloudAsset {
  id: number;
  accountId: number;
  provider: string;
  region: string;
  resourceType: string;
  resourceId: string;
  name: string;
  status: string;
  configJson: string;
  tagsJson: string;
  discoveredAt: string;
  lastSeenAt: string;
  deletedAt: string | null;
}

export interface CloudAssetPage {
  items: CloudAsset[];
  total: number;
  page: number;
  pageSize: number;
}

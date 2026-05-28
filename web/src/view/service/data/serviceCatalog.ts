// web/src/view/service/data/serviceCatalog.ts

import type { ServiceCategory, ServiceCatalogItem } from '../ticketTypes';

export const SERVICE_CATEGORIES: ServiceCategory[] = ['account', 'cloud', 'network', 'env', 'support'];

export const CATEGORY_I18N_KEYS: Record<ServiceCategory | 'popular', string> = {
  popular: 'servicePortal.category.popular',
  account: 'servicePortal.category.account',
  cloud: 'servicePortal.category.cloud',
  network: 'servicePortal.category.network',
  env: 'servicePortal.category.env',
  support: 'servicePortal.category.support',
};

export const SERVICE_CATALOG: ServiceCatalogItem[] = [
  {
    id: 'account.role-grant',
    category: 'account',
    i18nKey: 'servicePortal.catalog.accountRoleGrant',
    iconToken: 'shield',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.oneToTwoDays'],
    hot: true,
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'account.new',
    category: 'account',
    i18nKey: 'servicePortal.catalog.accountNew',
    iconToken: 'shield',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.trackable'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'account.offboarding',
    category: 'account',
    i18nKey: 'servicePortal.catalog.accountOffboarding',
    iconToken: 'shield',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.trackable'],
    estimatedSla: 'servicePortal.sla.threeBusinessDays',
  },
  {
    id: 'cloud.compute-resize',
    category: 'cloud',
    i18nKey: 'servicePortal.catalog.cloudComputeResize',
    iconToken: 'cloud',
    tags: ['servicePortal.tag.projectScope', 'servicePortal.tag.costReview'],
    hot: true,
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'cloud.storage-quota',
    category: 'cloud',
    i18nKey: 'servicePortal.catalog.cloudStorageQuota',
    iconToken: 'cloud',
    tags: ['servicePortal.tag.costReview'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'cloud.new-instance',
    category: 'cloud',
    i18nKey: 'servicePortal.catalog.cloudNewInstance',
    iconToken: 'cloud',
    tags: ['servicePortal.tag.projectScope', 'servicePortal.tag.costReview'],
    estimatedSla: 'servicePortal.sla.threeBusinessDays',
  },
  {
    id: 'cloud.release',
    category: 'cloud',
    i18nKey: 'servicePortal.catalog.cloudRelease',
    iconToken: 'cloud',
    tags: ['servicePortal.tag.approval'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'network.vpn-access',
    category: 'network',
    i18nKey: 'servicePortal.catalog.networkVpnAccess',
    iconToken: 'network',
    tags: ['servicePortal.tag.securityReview', 'servicePortal.tag.frequent'],
    hot: true,
    estimatedSla: 'servicePortal.sla.sameDay',
  },
  {
    id: 'network.acl',
    category: 'network',
    i18nKey: 'servicePortal.catalog.networkAcl',
    iconToken: 'network',
    tags: ['servicePortal.tag.securityReview'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'network.bastion',
    category: 'network',
    i18nKey: 'servicePortal.catalog.networkBastion',
    iconToken: 'network',
    tags: ['servicePortal.tag.securityReview', 'servicePortal.tag.trackable'],
    estimatedSla: 'servicePortal.sla.sameDay',
  },
  {
    id: 'env.deploy-window',
    category: 'env',
    i18nKey: 'servicePortal.catalog.envDeployWindow',
    iconToken: 'env',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.trackable'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'support.workplace',
    category: 'support',
    i18nKey: 'servicePortal.catalog.supportWorkplace',
    iconToken: 'support',
    tags: ['servicePortal.tag.fastResponse'],
    estimatedSla: 'servicePortal.sla.sameDay',
  },
  {
    id: 'support.incident',
    category: 'support',
    i18nKey: 'servicePortal.catalog.supportIncident',
    iconToken: 'support',
    tags: ['servicePortal.tag.fastResponse', 'servicePortal.tag.trackable'],
    hot: true,
    estimatedSla: 'servicePortal.sla.sameDay',
  },
];

export function getCatalogById(id: string): ServiceCatalogItem | undefined {
  return SERVICE_CATALOG.find((item) => item.id === id);
}

export function listByCategory(category: ServiceCategory | 'popular'): ServiceCatalogItem[] {
  if (category === 'popular') {
    return SERVICE_CATALOG.filter((item) => item.hot);
  }
  return SERVICE_CATALOG.filter((item) => item.category === category);
}

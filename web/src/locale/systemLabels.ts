import type { AppLocale } from './types';

const EN_STATUS_LABELS = {
  disabled: 'Disabled',
  enabled: 'Enabled',
} as const;

type StatusValue = keyof typeof EN_STATUS_LABELS;

const STATUS_LABELS = createLocaleLabelTable<StatusValue>({
  'zh-CN': {
    disabled: '禁用',
    enabled: '启用',
  },
  en: EN_STATUS_LABELS,
});

const EN_USER_SOURCE_LABELS = {
  ldap: 'LDAP',
  local: 'Local',
} as const;

type UserSourceValue = keyof typeof EN_USER_SOURCE_LABELS;

const USER_SOURCE_LABELS = createLocaleLabelTable<UserSourceValue>({
  'zh-CN': {
    ldap: 'LDAP',
    local: '本地',
  },
  en: EN_USER_SOURCE_LABELS,
});

const EN_ROLE_CODE_LABELS = {
  admin: 'Administrator',
  user: 'User',
} as const;

type RoleCode = keyof typeof EN_ROLE_CODE_LABELS;

const ROLE_CODE_LABELS = createLocaleLabelTable<RoleCode>({
  'zh-CN': {
    admin: '管理员',
    user: '用户',
  },
  en: EN_ROLE_CODE_LABELS,
});

const EN_PERMISSION_CODE_LABELS = {
  'dashboard.view': 'Dashboard',
  'system.audit.menu': 'Audit Logs',
  'system.audit.view': 'View audit logs',
  'system.auth.logout': 'Logout',
  'system.auth.me': 'Current user',
  'system.ldap.menu': 'LDAP',
  'system.ldap.test': 'Test LDAP config',
  'system.ldap.update': 'Update LDAP config',
  'system.ldap.view': 'View LDAP config',
  'system.menu': 'System',
  'system.permissions.view': 'View permissions',
  'system.roles.create': 'Create roles',
  'system.roles.menu': 'Roles',
  'system.roles.permissions.update': 'Update role permissions',
  'system.roles.status.update': 'Update role status',
  'system.roles.update': 'Update roles',
  'system.roles.view': 'View roles',
  'system.users.menu': 'Users',
  'system.users.roles.update': 'Update user roles',
  'system.users.status.update': 'Update user status',
  'system.users.update': 'Update users',
  'system.users.view': 'View users',
} as const;

type PermissionCode = keyof typeof EN_PERMISSION_CODE_LABELS;

const PERMISSION_CODE_LABELS = createLocaleLabelTable<PermissionCode>({
  'zh-CN': {
    'dashboard.view': '仪表盘',
    'system.audit.menu': '审计日志',
    'system.audit.view': '查看审计日志',
    'system.auth.logout': '退出登录',
    'system.auth.me': '当前用户',
    'system.ldap.menu': 'LDAP',
    'system.ldap.test': '测试 LDAP 配置',
    'system.ldap.update': '更新 LDAP 配置',
    'system.ldap.view': '查看 LDAP 配置',
    'system.menu': '系统',
    'system.permissions.view': '查看权限',
    'system.roles.create': '创建角色',
    'system.roles.menu': '角色',
    'system.roles.permissions.update': '更新角色权限',
    'system.roles.status.update': '更新角色状态',
    'system.roles.update': '更新角色',
    'system.roles.view': '查看角色',
    'system.users.menu': '用户',
    'system.users.roles.update': '更新用户角色',
    'system.users.status.update': '更新用户状态',
    'system.users.update': '更新用户',
    'system.users.view': '查看用户',
  },
  en: EN_PERMISSION_CODE_LABELS,
});

const EN_PERMISSION_TYPE_LABELS = {
  api: 'API',
  menu: 'Menu',
} as const;

type PermissionTypeValue = keyof typeof EN_PERMISSION_TYPE_LABELS;

const PERMISSION_TYPE_LABELS = createLocaleLabelTable<PermissionTypeValue>({
  'zh-CN': {
    api: 'API',
    menu: '菜单',
  },
  en: EN_PERMISSION_TYPE_LABELS,
});

export const AUDIT_ACTION_CODES = [
  'auth.login.success', 'auth.login.failure', 'auth.logout',
  'user.update', 'user.status.update', 'user.roles.update',
  'role.create', 'role.update', 'role.status.update', 'role.permissions.update',
  'ldap.config.update', 'ldap.test.success', 'ldap.test.failure',
] as const;

type AuditActionCode = typeof AUDIT_ACTION_CODES[number];
const AUDIT_ACTION_LABELS = createLocaleLabelTable<AuditActionCode>({
  en: {
    'auth.login.success': 'Sign in', 'auth.login.failure': 'Failed sign-in', 'auth.logout': 'Sign out',
    'user.update': 'Update user', 'user.status.update': 'Update user status', 'user.roles.update': 'Update user roles',
    'role.create': 'Create role', 'role.update': 'Update role', 'role.status.update': 'Update role status', 'role.permissions.update': 'Update role permissions',
    'ldap.config.update': 'Update LDAP configuration', 'ldap.test.success': 'LDAP connection test', 'ldap.test.failure': 'Failed LDAP connection test',
  },
  'zh-CN': {
    'auth.login.success': '登录', 'auth.login.failure': '登录失败', 'auth.logout': '退出登录',
    'user.update': '修改用户', 'user.status.update': '修改用户状态', 'user.roles.update': '修改用户角色',
    'role.create': '创建角色', 'role.update': '修改角色', 'role.status.update': '修改角色状态', 'role.permissions.update': '修改角色权限',
    'ldap.config.update': '修改 LDAP 配置', 'ldap.test.success': 'LDAP 连接测试', 'ldap.test.failure': 'LDAP 连接测试失败',
  },
});

export function getAuditActionLabel(value: string, locale: AppLocale): string {
  return getLocalizedLabel(AUDIT_ACTION_LABELS, value, locale);
}

const EN_AUDIT_RESULT_LABELS = {
  failure: 'Failure',
  success: 'Success',
} as const;

type AuditResultValue = keyof typeof EN_AUDIT_RESULT_LABELS;

const AUDIT_RESULT_LABELS = createLocaleLabelTable<AuditResultValue>({
  'zh-CN': {
    failure: '失败',
    success: '成功',
  },
  en: EN_AUDIT_RESULT_LABELS,
});

const EN_AUDIT_RESOURCE_TYPE_LABELS = {
  ldap: 'LDAP',
  role: 'Role',
  user: 'User',
} as const;

type AuditResourceTypeValue = keyof typeof EN_AUDIT_RESOURCE_TYPE_LABELS;

const AUDIT_RESOURCE_TYPE_LABELS = createLocaleLabelTable<AuditResourceTypeValue>({
  'zh-CN': {
    ldap: 'LDAP',
    role: '角色',
    user: '用户',
  },
  en: EN_AUDIT_RESOURCE_TYPE_LABELS,
});

export function getStatusLabel(value: string, locale: AppLocale): string {
  return getLocalizedLabel(STATUS_LABELS, value, locale);
}

export function getUserSourceLabel(value: string, locale: AppLocale): string {
  return getLocalizedLabel(USER_SOURCE_LABELS, value, locale);
}

export function getRoleCodeLabel(code: string, locale: AppLocale): string {
  return getLocalizedLabel(ROLE_CODE_LABELS, code, locale);
}

export function getPermissionCodeLabel(code: string, locale: AppLocale): string {
  return getLocalizedLabel(PERMISSION_CODE_LABELS, code, locale);
}

export function getPermissionTypeLabel(type: string, locale: AppLocale): string {
  return getLocalizedLabel(PERMISSION_TYPE_LABELS, type, locale);
}

export function getAuditResultLabel(value: string, locale: AppLocale): string {
  return getLocalizedLabel(AUDIT_RESULT_LABELS, value, locale);
}

export function getAuditResourceTypeLabel(value: string, locale: AppLocale): string {
  return getLocalizedLabel(AUDIT_RESOURCE_TYPE_LABELS, value, locale);
}

function getLocalizedLabel<Key extends string>(
  table: Record<AppLocale, Record<Key, string>>,
  value: string,
  locale: AppLocale,
): string {
  return hasLabel(table[locale], value) ? table[locale][value] : value;
}

function createLocaleLabelTable<Key extends string>(
  table: Record<AppLocale, Record<Key, string>>,
): Record<AppLocale, Record<Key, string>> {
  return table;
}

function hasLabel<Key extends string>(
  labels: Record<Key, string>,
  value: string,
): value is Key {
  return Object.prototype.hasOwnProperty.call(labels, value);
}

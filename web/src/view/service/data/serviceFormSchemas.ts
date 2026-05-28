// web/src/view/service/data/serviceFormSchemas.ts

import type { RequestFormField, ServiceFormSchemas } from '../ticketTypes';
import { GENERIC_FORM_KEY } from '../ticketTypes';

const accountRoleGrant: RequestFormField[] = [
  {
    name: 'projectCode',
    type: 'select',
    i18nLabelKey: 'servicePortal.form.projectCode.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.projectCode.placeholder',
    options: [
      { value: 'web', i18nLabelKey: 'servicePortal.form.projectCode.options.web' },
      { value: 'data', i18nLabelKey: 'servicePortal.form.projectCode.options.data' },
      { value: 'mobile', i18nLabelKey: 'servicePortal.form.projectCode.options.mobile' },
    ],
  },
  {
    name: 'role',
    type: 'select',
    i18nLabelKey: 'servicePortal.form.role.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.role.placeholder',
    options: [
      { value: 'developer', i18nLabelKey: 'servicePortal.form.role.options.developer' },
      { value: 'ops', i18nLabelKey: 'servicePortal.form.role.options.ops' },
      { value: 'admin', i18nLabelKey: 'servicePortal.form.role.options.admin' },
      { value: 'viewer', i18nLabelKey: 'servicePortal.form.role.options.viewer' },
    ],
  },
  {
    name: 'effectiveAt',
    type: 'date',
    i18nLabelKey: 'servicePortal.form.effectiveAt.label',
    required: false,
  },
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
];

const cloudComputeResize: RequestFormField[] = [
  {
    name: 'projectCode',
    type: 'select',
    i18nLabelKey: 'servicePortal.form.projectCode.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.projectCode.placeholder',
    options: [
      { value: 'web', i18nLabelKey: 'servicePortal.form.projectCode.options.web' },
      { value: 'data', i18nLabelKey: 'servicePortal.form.projectCode.options.data' },
      { value: 'mobile', i18nLabelKey: 'servicePortal.form.projectCode.options.mobile' },
    ],
  },
  {
    name: 'instanceSpec',
    type: 'text',
    i18nLabelKey: 'servicePortal.form.instanceSpec.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.instanceSpec.placeholder',
    maxLength: 60,
  },
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
];

const networkVpnAccess: RequestFormField[] = [
  {
    name: 'duration',
    type: 'select',
    i18nLabelKey: 'servicePortal.form.duration.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.duration.placeholder',
    options: [
      { value: 'temporary', i18nLabelKey: 'servicePortal.form.duration.options.temporary' },
      { value: 'longTerm', i18nLabelKey: 'servicePortal.form.duration.options.longTerm' },
    ],
  },
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
];

const generic: RequestFormField[] = [
  {
    name: 'title',
    type: 'text',
    i18nLabelKey: 'servicePortal.form.title.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.title.placeholder',
    maxLength: 80,
  },
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
  {
    name: 'contactPhone',
    type: 'text',
    i18nLabelKey: 'servicePortal.form.contactPhone.label',
    required: false,
    placeholderI18nKey: 'servicePortal.form.contactPhone.placeholder',
    maxLength: 40,
  },
];

// 默认 schema：含 reason + reasonAttachment，用于未显式定义自己 schema 的服务
const defaultSchema: RequestFormField[] = [
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
  {
    name: 'reasonAttachment',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reasonAttachment.label',
    required: false,
    placeholderI18nKey: 'servicePortal.form.reasonAttachment.placeholder',
    maxLength: 300,
  },
];

export const SERVICE_FORM_SCHEMAS: ServiceFormSchemas = {
  [GENERIC_FORM_KEY]: generic,
  'account.role-grant': accountRoleGrant,
  'cloud.compute-resize': cloudComputeResize,
  'network.vpn-access': networkVpnAccess,
};

export function getFormSchema(serviceId: string | null): RequestFormField[] {
  if (!serviceId) {
    return SERVICE_FORM_SCHEMAS[GENERIC_FORM_KEY];
  }
  return SERVICE_FORM_SCHEMAS[serviceId] ?? defaultSchema;
}

import { PageHeader } from "../../components/page/PageHeader";

import styles from "./SystemLdapPage.module.css";

const connectionItems = [
  { label: "Directory host", value: "ldap.internal.example.com" },
  { label: "Port", value: "389 / StartTLS" },
  { label: "Bind DN", value: "cn=admin,dc=example,dc=com" },
  { label: "Base DN", value: "ou=people,dc=example,dc=com" },
];

const syncItems = [
  { label: "Last sync", value: "Not connected yet" },
  { label: "Sync scope", value: "Users only" },
  { label: "Status", value: "Waiting for backend integration" },
];

const rolloutNotes = [
  "This page is a product shell for layout review only.",
  "Connection testing is a placeholder action in this phase.",
  "No real LDAP credentials, validation, or synchronization logic is wired yet.",
];

export function SystemLdapPage() {
  return (
    <div className={styles.page}>
      <PageHeader
        eyebrow="System / LDAP"
        title="Directory access will be configured here"
        description="The first frontend iteration defines the LDAP management surface before any real integration work begins. This page focuses on display structure, content hierarchy, and operator-facing copy."
        actions={
          <button className={styles.primaryAction} type="button">
            Test Connection
          </button>
        }
      />

      <div className={styles.grid}>
        <section className={styles.panel}>
          <div className={styles.panelHeader}>
            <p className={styles.sectionEyebrow}>Connection settings</p>
            <h2 className={styles.sectionTitle}>Directory endpoint details</h2>
          </div>
          <p className={styles.sectionDescription}>
            Keep the core connection fields visible so operators can quickly
            understand how LDAP will be connected once backend support is ready.
          </p>

          <dl className={styles.definitionList}>
            {connectionItems.map((item) => (
              <div className={styles.definitionItem} key={item.label}>
                <dt className={styles.definitionLabel}>{item.label}</dt>
                <dd className={styles.definitionValue}>{item.value}</dd>
              </div>
            ))}
          </dl>

          <div className={styles.notice}>
            <p className={styles.noticeTitle}>Implementation note</p>
            <p className={styles.noticeCopy}>
              The values shown here are static placeholders for the page shell.
              Saving, validation, and secure credential handling are intentionally
              out of scope in this phase.
            </p>
          </div>
        </section>

        <section className={styles.panel}>
          <div className={styles.panelHeader}>
            <p className={styles.sectionEyebrow}>Sync and status</p>
            <h2 className={styles.sectionTitle}>Readiness before activation</h2>
          </div>
          <p className={styles.sectionDescription}>
            Reserve space for sync visibility and operational feedback without
            implying that directory sync is already implemented.
          </p>

          <dl className={styles.definitionList}>
            {syncItems.map((item) => (
              <div className={styles.definitionItem} key={item.label}>
                <dt className={styles.definitionLabel}>{item.label}</dt>
                <dd className={styles.definitionValue}>{item.value}</dd>
              </div>
            ))}
          </dl>

          <div className={styles.callout}>
            <p className={styles.calloutLabel}>Current phase</p>
            <ul className={styles.calloutList}>
              {rolloutNotes.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </div>
        </section>
      </div>
    </div>
  );
}

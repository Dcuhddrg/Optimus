import { PageHeader } from "../../components/page/PageHeader";
import { EmptyState } from "../../components/state/EmptyState";

import styles from "./DashboardPage.module.css";

export function DashboardPage() {
  return (
    <div className={styles.page}>
      <PageHeader
        eyebrow="Dashboard"
        title="Platform overview is under construction"
        description="This area is reserved for the future operations overview. The first frontend iteration keeps the dashboard intentionally empty while the shell, navigation, and system settings surfaces are being established."
      />

      <EmptyState
        eyebrow="Coming soon"
        title="A focused ops overview will land here"
        description="The dashboard will become the platform's shared working surface after the first set of foundational pages is in place. For now, this placeholder keeps the structure clear without inventing metrics, charts, or operational summaries that do not exist yet."
        aside={
          <div className={styles.aside}>
            <p className={styles.asideLabel}>Planned use</p>
            <ul className={styles.asideList}>
              <li>Platform-wide status summary</li>
              <li>High-priority operational context</li>
              <li>Action-oriented overview modules</li>
            </ul>
          </div>
        }
      />
    </div>
  );
}

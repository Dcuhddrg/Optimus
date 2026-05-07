export function ForbiddenPage() {
  return (
    <section className="page-card">
      <p className="page-eyebrow">Access</p>
      <h1 className="page-title">Permission required</h1>
      <p className="page-description">
        Your account is signed in, but this area is not included in your current role permissions.
      </p>
    </section>
  );
}

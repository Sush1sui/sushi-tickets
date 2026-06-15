"use client";

const REDIRECT_URL = process.env.NEXT_PUBLIC_REDIRECT_URL ?? "";

export default function RedirectPage() {
  return (
    <main style={styles.main}>
      <div style={styles.card}>
        {/* Logo / icon area */}
        <div style={styles.iconWrap}>
          <span style={styles.icon}>🍣</span>
        </div>

        <h1 style={styles.heading}>We&apos;ve moved!</h1>

        <p style={styles.body}>
          <strong>Finesse Tickets</strong> has transitioned into{" "}
          <strong>Sushi Tickets</strong>. The old dashboard is no longer active.
        </p>

        <p style={styles.sub}>
          Head over to the new platform to manage your tickets and panels.
        </p>

        {REDIRECT_URL ? (
          <a href={REDIRECT_URL} style={styles.button}>
            Go to Sushi Tickets →
          </a>
        ) : (
          <p style={styles.warning}>
            Redirect URL not configured. Please contact the administrator.
          </p>
        )}

        <p style={styles.footer}>
          The new platform is rebuilt from scratch. You will need to recreate your configurations.
        </p>
      </div>
    </main>
  );
}

const styles: Record<string, React.CSSProperties> = {
  main: {
    minHeight: "100vh",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    padding: "24px",
    background: "radial-gradient(ellipse at 50% 0%, #1f1a2e 0%, #0e0e10 60%)",
  },
  card: {
    background: "#1a1a1f",
    border: "1px solid #2a2a32",
    borderRadius: "16px",
    padding: "48px 40px",
    maxWidth: "480px",
    width: "100%",
    textAlign: "center",
    boxShadow: "0 24px 64px rgba(0,0,0,0.5)",
  },
  iconWrap: {
    marginBottom: "24px",
  },
  icon: {
    fontSize: "56px",
    lineHeight: 1,
  },
  heading: {
    fontSize: "28px",
    fontWeight: 700,
    color: "#f0f0f5",
    marginBottom: "16px",
    letterSpacing: "-0.3px",
  },
  body: {
    fontSize: "15px",
    color: "#c0c0cc",
    lineHeight: 1.6,
    marginBottom: "10px",
  },
  sub: {
    fontSize: "14px",
    color: "#8e8ea0",
    lineHeight: 1.6,
    marginBottom: "32px",
  },
  button: {
    display: "inline-block",
    background: "#FF5A36",
    color: "#ffffff",
    fontWeight: 600,
    fontSize: "15px",
    padding: "13px 28px",
    borderRadius: "10px",
    textDecoration: "none",
    transition: "background 0.15s ease, transform 0.1s ease",
    cursor: "pointer",
    marginBottom: "28px",
  },
  warning: {
    fontSize: "13px",
    color: "#e0a000",
    background: "#2a2200",
    border: "1px solid #4a3800",
    borderRadius: "8px",
    padding: "10px 16px",
    marginBottom: "28px",
  },
  footer: {
    fontSize: "12px",
    color: "#5a5a6e",
    lineHeight: 1.5,
  },
};

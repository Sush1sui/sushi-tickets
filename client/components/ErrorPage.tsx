import Footer from "./Footer";
import "./error-page.css";

export default function ErrorPage({ title, message, reset }: {
  title: string;
  message: string;
  reset?: () => void;
}) {
  return (
    <div className="failure-page">
      <div className="failure-frame">
        <header>
          <a href="/" className="failure-brand">Sushi Tickets</a>
        </header>
        <main className="failure-content">
          <h1>{title}</h1>
          <p>{message}</p>
          <div className="failure-actions">
            {reset && <button type="button" onClick={reset}>Try again</button>}
            <a href="/">Go to homepage</a>
          </div>
        </main>
        <Footer />
      </div>
    </div>
  );
}

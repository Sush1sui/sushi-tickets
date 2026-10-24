"use client";

import ErrorPage from "../components/ErrorPage";
import "./globals.css";
import "./overhaul.css";

export default function GlobalError({ reset }: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="en">
      <head><title>Something went wrong | Sushi Tickets</title></head>
      <body style={{ margin: 0 }}>
        <ErrorPage
          title="Something went wrong"
          message="Sushi Tickets couldn't start this page. Try again, or return to the homepage."
          reset={reset}
        />
      </body>
    </html>
  );
}

"use client";

import ErrorPage from "../components/ErrorPage";

export default function Error({ reset }: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return <ErrorPage
    title="Something went wrong"
    message="We couldn't load this page. Try again, or return to the homepage. If the problem continues, contact our support server."
    reset={reset}
  />;
}

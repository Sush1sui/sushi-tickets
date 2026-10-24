import type { Metadata } from "next";
import ErrorPage from "../components/ErrorPage";

export const metadata: Metadata = {
  title: "Page not found",
  robots: { index: false, follow: true },
};

export default function NotFound() {
  return <ErrorPage
    title="Page not found"
    message="This link may be outdated, or the page may have moved. Head back to Sushi Tickets to continue."
  />;
}

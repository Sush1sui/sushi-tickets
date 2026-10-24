import type { Metadata } from "next";
import Link from "next/link";
import LegalPage from "../../components/LegalPage";

export const metadata: Metadata = {
  title: "Terms of Service",
  description: "Terms for using the Sushi Tickets Discord bot and dashboard.",
  alternates: { canonical: "/terms" },
};

export default function TermsPage() {
  return <LegalPage
    title="Terms of Service"
    introduction="The rules for using Sushi Tickets, our Discord ticket bot and management dashboard."
    sections={[
      { id: "agreement", title: "Your agreement", content: <>
        <p>These terms govern your use of Sushi Tickets (the Service), operated by the Sushi Tickets developer team. By installing, configuring, or using the Service, you agree to these terms. If you do not agree, stop using the Service and remove the bot from servers you manage.</p>
        <p>You must meet Discord&apos;s minimum age requirements and the minimum age required in your country. Where required, a parent or guardian must consent to your use. You must have permission to manage any server you configure.</p>
      </> },
      { id: "service", title: "What the Service does", content: <>
        <p>Sushi Tickets lets Discord communities create ticket channels, configure panels and questions, manage staff access and opening schedules, and save ticket transcripts when transcript logging is enabled. Features depend on server settings, bot permissions, and Discord availability.</p>
        <p>Sushi Tickets is an independent application. It is not operated, sponsored, or endorsed by Discord.</p>
      </> },
      { id: "responsibilities", title: "Server and user responsibilities", content: <>
        <p>Server administrators control ticket settings, staff access, and transcript logging. Administrators must tell members when their ticket content will be saved, obtain any legally required permissions, and limit access to people who need it.</p>
        <p>You are responsible for your account, your server configuration, and the content you submit. Share only information needed to resolve your ticket. Do not submit passwords, payment credentials, access tokens, or other highly sensitive information.</p>
      </> },
      { id: "acceptable-use", title: "Acceptable use", content: <>
        <p>You must follow applicable laws, <a href="https://discord.com/terms">Discord&apos;s Terms of Service</a>, and <a href="https://discord.com/guidelines">Discord&apos;s Community Guidelines</a>. You must not use the Service to:</p>
        <ul>
          <li>Harass, threaten, scam, spam, or distribute unlawful or harmful content.</li>
          <li>Collect, expose, or share someone&apos;s personal information without authorization.</li>
          <li>Access tickets or transcripts without permission, bypass access controls, or steal credentials.</li>
          <li>Disrupt the Service, abuse automated requests, or evade usage limits.</li>
          <li>Infringe intellectual property, privacy, or other rights.</li>
        </ul>
      </> },
      { id: "content-and-privacy", title: "Your content and privacy", content: <>
        <p>You retain any rights you hold in content you submit. You give us permission to process, store, and display that content only as needed to provide and protect the Service, including delivering tickets and generating configured transcripts.</p>
        <p>Our <Link href="/privacy">Privacy Policy</Link> explains what information we process and how to request access, correction, or deletion. Closing a ticket or removing the bot does not automatically erase previously saved transcripts or copies held by server staff.</p>
      </> },
      { id: "availability", title: "Availability and changes", content: <>
        <p>We may change, suspend, or discontinue features, or restrict access to address abuse, security risks, legal obligations, or violations of these terms. Where practical, we will give notice of material changes or discontinuation.</p>
        <p>We do not guarantee uninterrupted availability, successful delivery of every action, or permanent storage. Discord outages, missing permissions, hosting failures, and changes to third-party services can affect operation. Keep independent copies of records you are required to retain.</p>
      </> },
      { id: "disclaimers", title: "Disclaimers and liability", content: <>
        <p>To the extent permitted by applicable law, the Service is provided &quot;as is&quot; and &quot;as available&quot; without warranties of merchantability, fitness for a particular purpose, or non-infringement.</p>
        <p>To the extent permitted by applicable law, the developer team is not liable for indirect or consequential losses, lost profits, or loss of data arising from use of the Service. Nothing in these terms limits liability or rights that cannot lawfully be limited, including mandatory consumer protections.</p>
      </> },
      { id: "ending-use", title: "Ending your use", content: <>
        <p>You can stop using the Service at any time. Server administrators can remove the bot; dashboard users can sign out and revoke Sushi Tickets access in Discord&apos;s Authorized Apps settings. To request deletion of stored information, contact the developer team as described below.</p>
      </> },
      { id: "updates", title: "Updates to these terms", content: <>
        <p>We will publish revised terms here and update the date above. Material changes will be announced through the dashboard or support server where practical. Changes apply prospectively; continued use after they take effect means you accept the revised terms, subject to applicable law.</p>
      </> },
    ]}
  />;
}

import type { Metadata } from "next";
import LegalPage from "../../components/LegalPage";

export const metadata: Metadata = {
  title: "Privacy Policy",
  description: "How Sushi Tickets handles Discord account information, ticket content, transcripts, and data deletion requests.",
  alternates: { canonical: "/privacy" },
};

export default function PrivacyPage() {
  return <LegalPage
    title="Privacy Policy"
    introduction="What Sushi Tickets processes, why we need it, and how you can ask us to delete it."
    sections={[
      { id: "scope", title: "Who this policy covers", content: <>
        <p>This policy explains how the Sushi Tickets developer team handles information when you use our Discord bot, ticket features, and management dashboard. Server administrators separately decide how their communities use tickets and any copies of conversations they retain.</p>
      </> },
      { id: "information", title: "Information we process", content: <>
        <ul>
          <li><strong>Discord account information:</strong> user ID, username or display name, avatar, and, when you sign into the dashboard, the email address supplied by Discord. Dashboard authorization also lets us retrieve your server list and relevant permissions.</li>
          <li><strong>Server configuration:</strong> server, channel, role, staff, and panel IDs; ticket settings; welcome messages; configured questions; and opening schedules.</li>
          <li><strong>Ticket information:</strong> the opener&apos;s user ID, ticket channel ID, creation time, submitted form answers, and content sent in the ticket channel.</li>
          <li><strong>Transcripts, when enabled:</strong> ticket messages, author identifiers and names, timestamps, attachment links and metadata, embeds, participant information, and ticket opening and closing details. Transcripts may contain personal information that participants choose to share.</li>
          <li><strong>Login and security information:</strong> session identifiers, encrypted Discord OAuth access tokens, session creation and expiry times, and request information used for abuse prevention. IP addresses are processed for rate limiting; operational logs may include identifiers and error details.</li>
        </ul>
        <p>Information comes from Discord, server administrators, your dashboard activity, and content you submit. We do not ask for your Discord password. The dashboard requests the Discord OAuth scopes for identity, email, and server membership.</p>
      </> },
      { id: "use", title: "How we use information", content: <>
        <p>We use information to sign you in, check server permissions, create and manage tickets, deliver form answers, apply server settings, generate configured transcripts, respond to support and privacy requests, and keep the Service secure and working.</p>
        <p>We do not sell personal information, use ticket content for advertising, or use Discord message content to train AI models. Where applicable privacy law requires a legal basis, processing supports providing the Service you request, our legitimate interests in operating and securing it, compliance with legal obligations, or consent where required.</p>
      </> },
      { id: "sharing", title: "Who can access information", content: <>
        <p>Ticket content is available to participants and people granted access through the server&apos;s Discord permissions. Dashboard settings and saved transcripts are available to server owners, members with administrator or manage-server permissions, and members or roles authorized in Sushi Tickets. Administrators are responsible for choosing that access carefully.</p>
        <p>Discord receives information required for bot actions and login. Our hosting, database, and storage providers process information to run the Service. Authorized developers may access information when necessary for support, maintenance, security, or a privacy request. We may disclose information when required by law.</p>
        <p>Server staff may export, copy, or share content outside the Service. Ask the relevant server administrators about those copies. Discord handles information under its own <a href="https://discord.com/privacy">Privacy Policy</a>.</p>
      </> },
      { id: "retention", title: "Storage and retention", content: <>
        <p>Server settings, active-ticket records, and transcript metadata are stored in our database. Saved transcript files are stored on the Service&apos;s hosting infrastructure.</p>
        <ul>
          <li>Dashboard sessions expire after seven days. Signing out deletes the current server-side session; expired sessions are removed by scheduled cleanup.</li>
          <li>Active-ticket records are removed during normal ticket closure or cleanup of confirmed deleted channels. These records are separate from saved transcripts.</li>
          <li>Server configuration and saved transcripts currently have no automatic expiry. We retain them to provide the configured ticket and conversation-history features, and remove them when no longer needed or in response to a valid deletion request, subject to legal requirements.</li>
        </ul>
        <p>Removing the bot, deleting a Discord channel, or closing a ticket does not by itself delete an existing saved transcript. Copies held by Discord or server staff follow their own retention practices. If information exists in infrastructure backups, removal from those copies depends on the provider&apos;s backup lifecycle and applicable legal requirements.</p>
      </> },
      { id: "cookies", title: "Cookies and login", content: <>
        <p>The dashboard uses cookies for your login session, OAuth sign-in verification, and protection against forged requests. Blocking these cookies may prevent sign-in or dashboard changes. You can sign out, clear cookies in your browser, and revoke access through Discord&apos;s Authorized Apps settings.</p>
      </> },
      { id: "choices", title: "Your choices and deletion requests", content: <>
        <p>You can ask the developer team for access to, correction of, or deletion of your personal information through the contact options below. Include your Discord user ID, the relevant server ID, and ticket or transcript IDs if known. Do not send your password or access token. We may ask for proportionate proof of identity or authority before acting.</p>
        <p>You may contact server administrators about their copies, but you do not need their permission to submit a personal-data request to us. We will address verified requests promptly and within applicable legal deadlines, explaining any lawful reason we cannot fulfill a request in full. Deletion of a saved transcript may affect other participants&apos; conversation history.</p>
        <p>Depending on your location, you may also have rights to portability, restriction, objection, withdrawal of consent, or complaint to a data protection authority. Withdrawing consent does not affect processing that was lawful before withdrawal.</p>
      </> },
      { id: "security", title: "Security and international processing", content: <>
        <p>We use access controls, server authorization checks, and encryption of stored OAuth access tokens to help protect information. No system can guarantee complete security. Share only information needed for your ticket.</p>
        <p>Discord and infrastructure providers may process information outside your country. Applicable legal requirements govern those transfers. Contact us if you need more information about where your data is processed.</p>
      </> },
      { id: "children", title: "Age requirements", content: <>
        <p>The Service is intended for people who meet Discord&apos;s minimum age and the requirements of their country. If you believe a child below that age has provided personal information, contact us so we can investigate and remove it where required.</p>
      </> },
      { id: "updates", title: "Policy updates", content: <>
        <p>We will post changes on this page and update the date above. Where practical, material changes will also be announced through the dashboard or support server. We will obtain consent for new processing where required by law.</p>
      </> },
    ]}
  />;
}

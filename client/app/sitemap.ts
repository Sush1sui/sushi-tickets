import { MetadataRoute } from "next";

export default function sitemap(): MetadataRoute.Sitemap {
  const siteUrl = process.env.NEXT_PUBLIC_SITE_URL || "https://sushitickets.app";

  return [
    {
      url: siteUrl,
      lastModified: new Date(),
      changeFrequency: "weekly",
      priority: 1,
    },
    ...["terms", "privacy"].map((path) => ({
      url: `${siteUrl}/${path}`,
      lastModified: new Date("2026-10-24T00:00:00+08:00"),
      changeFrequency: "yearly" as const,
      priority: 0.3,
    })),
  ];
}

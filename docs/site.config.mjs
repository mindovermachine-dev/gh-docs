export default {
  theme: "mindovermachine-dev/mom-doc-theme@main",
  title: "gh-docs",
  site: "https://docs.mindovermachine.dev",
  rootRedirect: null,
  defaultLocale: "root",
  locales: { root: { label: "English", lang: "en" } },
  astro: { base: "/gh-docs", server: { host: true } },
  social: [
    {
      icon: "github",
      label: "GitHub",
      href: "https://github.com/mindovermachine-dev/gh-docs",
    },
  ],
  sidebar: [
    {
      label: "Commands",
      items: [{ autogenerate: { directory: "commands" } }],
    },
    {
      label: "Output Formats",
      items: [{ autogenerate: { directory: "output-formats" } }],
    },
  ],
};

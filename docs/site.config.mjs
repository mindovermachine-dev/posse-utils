// Steering config for remote-astro-theme: builds docs/content with the theme below.
export default {
  //theme: "mindovermachine-dev/mom-doc-theme@main",
  theme: "mindovermachine-dev/mom-doc-theme@main",
  // Raw Astro config, merged over the theme's astro.config.mjs.
  astro: {
    site: "https://docs.mindovermachine.dev",
    base: "/posse-utils",
    server: { host: true },
  },
  title: "POSSE utilities docs",
  // null (not undefined) so the theme's default "/da/" redirect is overridden.
  rootRedirect: null,
  defaultLocale: "root",
  locales: {
    root: { label: "English", lang: "en" },
  },
  social: [
    {
      icon: "github",
      label: "GitHub",
      href: "https://github.com/mindovermachine-dev",
    },
  ],
  // One group per folder under docs/content; order pages with `sidebar.order` frontmatter.
  sidebar: [
    {
      label: "Development",
      items: [{ autogenerate: { directory: "development" } }],
    },
    {
      label: "Design",
      items: [{ autogenerate: { directory: "design" } }],
    },
    { 
      label: "Installation", 
      items: [{ autogenerate: { directory: "installation" } }] 
    },
    {
      slug: "contributing",
    },
  ],
};

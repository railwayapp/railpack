// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwindcss from "@tailwindcss/vite";

// https://astro.build/config
export default defineConfig({
  prefetch: {
    prefetchAll: true,
    defaultStrategy: "hover",
  },

  integrations: [
    starlight({
      title: "Railpack Docs",
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/railwayapp/railpack",
        },
      ],
      editLink: {
        baseUrl: "https://github.com/railwayapp/railpack/edit/main/docs/",
      },
      favicon: "/favicon.svg?v=2",
      customCss: [
        "./src/tailwind.css",

        "@fontsource/inter/400.css",
        "@fontsource/inter/600.css",
      ],
      sidebar: [
        {
          label: "Getting Started",
          link: "/getting-started",
        },
        {
          label: "Installation",
          link: "/installation",
        },
        {
          label: "Guides",
          items: [
            {
              label: "Installing Additional Packages",
              link: "/guides/installing-packages",
            },
            {
              label: "Adding Steps",
              link: "/guides/adding-steps",
            },
            {
              label: "Developing Locally",
              link: "/guides/developing-locally",
            },
            {
              label: "Running Railpack in Production",
              link: "/guides/running-railpack-in-production",
            },
          ],
        },
        {
          label: "Configuration",
          items: [
            { label: "Configuration File", link: "/config/file" },
            {
              label: "Environment Variables",
              link: "/config/environment-variables",
            },
            { label: "Procfile", link: "/config/procfile" },
          ],
        },
        {
          label: "Languages",
          items: [
            { label: "Node", link: "/languages/node" },
            { label: "Python", link: "/languages/python" },
            { label: "Go", link: "/languages/golang" },
            { label: "PHP", link: "/languages/php" },
            { label: "Java", link: "/languages/java" },
            { label: "Ruby", link: "/languages/ruby" },
            { label: "Deno", link: "/languages/deno" },
            { label: "Rust", link: "/languages/rust" },
            { label: "Elixir", link: "/languages/elixir" },
            { label: "Staticfile", link: "/languages/staticfile" },
            { label: "Shell Scripts", link: "/languages/shell" },
          ],
        },
        {
          label: "Reference",
          items: [
            { label: "CLI Commands", link: "/reference/cli" },
            { label: "BuildKit Frontend", link: "/reference/frontend" },
          ],
        },
        {
          label: "Architecture",
          items: [
            { label: "High Level Overview", link: "/architecture/overview" },
            {
              label: "Package Resolution",
              link: "/architecture/package-resolution",
            },
            {
              label: "Secrets and Variables",
              link: "/architecture/secrets",
            },
            { label: "BuildKit Generation", link: "/architecture/buildkit" },
            { label: "Caching", link: "/architecture/caching" },
            { label: "User Config", link: "/architecture/user-config" },
          ],
        },
        {
          label: "Contributing",
          link: "/contributing",
        },
      ],
    }),
  ],

  vite: {
    plugins: [tailwindcss()],
  },
});

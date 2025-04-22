// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwind from "@astrojs/tailwind";

// https://astro.build/config
export default defineConfig({
  prefetch: {
    prefetchAll: true,
    defaultStrategy: "hover",
  },

  integrations: [
    starlight({
      title: "Railpack Docs",
      social: {
        github: "https://github.com/railwayapp/railpack",
      },
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
            // {
            //   label: "Building with CLI and BuildKit",
            //   link: "/guides/building-with-cli",
            // },
            // {
            //   label: "Building with a Custom Frontend",
            //   link: "/guides/custom-frontend",
            // },
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
          ],
        },
        {
          label: "Languages",
          items: [
            { label: "PHP", link: "/languages/php" },
            { label: "Go", link: "/languages/golang" },
            { label: "Java", link: "/languages/java" },
            { label: "Python", link: "/languages/python" },
            { label: "Deno", link: "/languages/deno" },
            { label: "Ruby", link: "/languages/ruby" },
            { label: "Rust", link: "/languages/rust" },
            { label: "Node", link: "/languages/node" },
            { label: "Dotnet", link: "/languages/dotnet" },
            { label: "Staticfile", link: "/languages/staticfile" },
            { label: "Shell Scripts", link: "/languages/shell" },
          ],
        },
        {
          label: "Reference",
          items: [{ label: "CLI Commands", link: "/reference/cli" }],
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

    tailwind({
      // Disable the default base styles:
      applyBaseStyles: false,
    }),
  ],
});

// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwindcss from "@tailwindcss/vite";
import starlightLlmsTxt from "starlight-llms-txt";

// https://astro.build/config
export default defineConfig({
  site: "https://railpack.com",

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
      plugins: [
        starlightLlmsTxt({
          projectName: "Railpack",
          description:
            "Zero-config application builder that automatically analyzes your code and turns it into a container image. Built on BuildKit with support for Node, Python, Go, PHP, and more.",
          details:
            "Railpack provides a seamless way to build container images from your source code without complex configuration. It automatically detects your project type and generates appropriate build steps.",
          customSets: [
            {
              label: "Languages Reference",
              description:
                "Language-specific documentation for all supported platforms",
              paths: ["languages/**"],
            },
            {
              label: "Architecture",
              description:
                "Technical details about Railpack's internal architecture",
              paths: ["architecture/**"],
            },
            {
              label: "Guides",
              description: "Step-by-step guides for common tasks",
              paths: ["guides/**"],
            },
            {
              label: "Configuration",
              description: "Configuration options and environment variables",
              paths: ["config/**"],
            },
            {
              label: "Deploying",
              description: "Deployment guides for Railway and GitHub Actions",
              paths: ["deploying/**"],
            },
            {
              label: "Reference",
              description: "CLI commands and BuildKit frontend reference",
              paths: ["reference/**"],
            },
          ],
          optionalLinks: [
            {
              label: "Railpack GitHub Repository",
              url: "https://github.com/railwayapp/railpack",
              description: "Source code and issue tracking for Railpack",
            },
            {
              label: "Railway",
              url: "https://railway.com",
              description: "Cloud platform that created Railpack",
            },
            {
              label: "Railway Railpack Guide",
              url: "https://docs.railway.com/guides/build-configuration#railpack",
              description: "How to use Railpack on Railway platform",
            },
          ],
          promote: ["index*", "getting-started*", "installation*", "config/**"],
        }),
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
            { label: "Dotnet", link: "/languages/dotnet" },
            { label: "Deno", link: "/languages/deno" },
            { label: "Rust", link: "/languages/rust" },
            { label: "Elixir", link: "/languages/elixir" },
            { label: "Gleam", link: "/languages/gleam" },
            { label: "C/C++", link: "/languages/cpp" },
            { label: "Staticfile", link: "/languages/staticfile" },
            { label: "Shell Scripts", link: "/languages/shell" },
          ],
        },
        {
          label: "Deploying",
          items: [
            { label: "Railway", link: "/deploying/railway" },
            { label: "GitHub Actions", link: "/deploying/github-actions" },
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

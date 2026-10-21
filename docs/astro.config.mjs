// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwindcss from "@tailwindcss/vite";
import starlightLlmsTxt from "starlight-llms-txt";
import starlightPageActions from "starlight-page-actions";

// https://astro.build/config
export default defineConfig({
  site: "https://railpack.com",

  redirects: {
    "/architecture/buildkit": "/platforms/buildkit",
    "/architecture/caching": "/platforms/caching",
    "/architecture/recommendations": "/config/recommendations",
    "/guides/running-railpack-in-production":
      "/platforms/running-railpack-in-production",
    "/reference/frontend": "/platforms/buildkit-frontend",
    "/resolving-errors": "/config/resolving-errors",
  },

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
        "@fontsource/inter/500.css",
        "@fontsource/inter/600.css",
        "@fontsource/ibm-plex-serif/400.css",
        "@fontsource/ibm-plex-serif/500.css",
        "@fontsource/ibm-plex-serif/600.css",
        "@fontsource/jetbrains-mono/400.css",
        "@fontsource/jetbrains-mono/500.css",
      ],
      expressiveCode: {
        // Shell blocks use terminal frames so we can render a Railway-style
        // top bar (label left, copy right). Other langs keep auto/code frames.
        defaultProps: {
          overridesByLang: {
            "bash,sh,shell,zsh": {
              frame: "terminal",
            },
          },
        },
        styleOverrides: {
          borderRadius: "0.5rem",
          // Match railpack.com code metrics
          codeFontFamily: "var(--font-mono)",
          codeFontSize: "0.875rem",
          codeLineHeight: "1.75",
          codePaddingBlock: "0.75rem",
          codePaddingInline: "1rem",
          frames: {
            shadowColor: "transparent",
            frameBoxShadowCssValue: "none",
          },
        },
      },
      plugins: [
        starlightPageActions({
          position: "table-of-contents",
        }),
        starlightLlmsTxt({
          projectName: "Railpack",
          description:
            "Zero-config application builder that analyzes code and builds an image. Built on BuildKit with support for Node, Python, Go, PHP, and more.",
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
              label: "Platforms",
              description:
                "Documentation for integrating Railpack into hosting platforms",
              paths: ["platforms/**"],
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
          label: "Help",
          link: "/help",
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
          ],
        },
        {
          label: "Configuration",
          items: [
            { label: "Configuration Options", link: "/config/options" },
            { label: "Configuration File", link: "/config/file" },
            {
              label: "Environment Variables",
              link: "/config/environment-variables",
            },
            { label: "Mise", link: "/config/mise" },
            { label: "Procfile", link: "/config/procfile" },
            { label: "Excluding Files", link: "/config/excluding-files" },
            { label: "Recommendations", link: "/config/recommendations" },
            { label: "Resolving Errors", link: "/config/resolving-errors" },
          ],
        },
        {
          label: "Languages",
          items: [
            { label: "Node", link: "/languages/node" },
            { label: "Bun", link: "/languages/bun" },
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
            { label: "Changelog", link: "/changelog" },
            { label: "FAQ", link: "/faq" },
          ],
        },
        {
          label: "Architecture",
          items: [
            { label: "Design Goals", link: "/architecture/design-goals" },
            { label: "High Level Overview", link: "/architecture/overview" },
            {
              label: "Package Resolution",
              link: "/architecture/package-resolution",
            },
            {
              label: "Secrets and Variables",
              link: "/architecture/secrets",
            },
          ],
        },
        {
          label: "Contributing",
          link: "/contributing",
        },
        {
          label: "Platforms",
          items: [
            {
              label: "Build with Railpack",
              link: "/platforms/build-with-railpack",
            },
            {
              label: "Running Railpack in Production",
              link: "/platforms/running-railpack-in-production",
            },
            {
              label: "BuildKit Frontend",
              link: "/platforms/buildkit-frontend",
            },
            {
              label: "BuildKit Generation",
              link: "/platforms/buildkit",
            },
            { label: "Caching", link: "/platforms/caching" },
            {
              label: "Package Version Resolution",
              link: "/platforms/package-version-resolution",
            },
          ],
        },
      ],
      components: {
        MobileTableOfContents: "./src/components/MobileTableOfContents.astro",
        TableOfContents: "./src/components/TableOfContents.astro",
      },
    }),
  ],

  vite: {
    plugins: [tailwindcss()],
  },
});

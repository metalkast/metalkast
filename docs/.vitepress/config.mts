import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "metalkast",
  description: "Kubernetes Baremetal Clusters Deployer",
  themeConfig: {
    logo: '/metalkast.png',
    socialLinks: [
      { icon: 'github', link: 'https://github.com/metalkast/metalkast' }
    ],
    sidebar: [
      {
        text: 'Guide',
        items: [
          { text: 'Get started', link: '/get-started' },
        ]
      },
      {
        text: 'Reference',
        items: [
          { text: 'Annotations', link: '/annotations' },
        ]
      }
    ],
    docFooter: {
      prev: false,
      next: false,
    },
  }
})

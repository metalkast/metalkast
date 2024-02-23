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
          { text: 'Get started', link: '/get-started/' },
          { text: 'SOPS Integration', link: '/sops' },
        ]
      },
      {
        text: 'Reference',
        items: [
          { text: 'Annotations', link: '/annotations' },
          { text: 'Image Releases', link: '/image-releases' },
        ]
      }
    ],
    docFooter: {
      prev: false,
      next: false,
    },
  },
})

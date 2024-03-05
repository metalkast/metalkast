import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "metalkast",
  description: "Kubernetes on Bare Metal Made Easy",
  head: [
    ['link', { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32x32.png' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '16x16', href: '/favicon-16x16.png' }],
    ['link', { rel: 'manifest', href: '/site.webmanifest' }],
  ],
  themeConfig: {
    logo: '/favicon.ico',
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

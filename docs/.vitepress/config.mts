import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "metalkast",
  description: "Kubernetes Baremetal Clusters Deployer",
  themeConfig: {
    logo: '/metalkast.png',
    socialLinks: [
      { icon: 'github', link: 'https://github.com/metalkast/metalkast' }
    ]
  }
})

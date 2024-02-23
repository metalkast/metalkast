import { createMarkdownRenderer } from "vitepress";

export async function renderMarkdown(markdown: string): Promise<string> {
    const config = global.VITEPRESS_CONFIG;
    let renderer = await createMarkdownRenderer(
        config.srcDir,
        config.markdown,
        config.site.base,
        config.logger
    );
    return renderer.render(markdown);
}

/**
 * Reads English release body from release_body_en.md (written by CI),
 * calls an OpenAI-compatible Chat Completions API for Chinese translation,
 * and writes translated_release.md with bilingual content
 * (English original + --- separator + Chinese translation).
 *
 * Env:
 *   LLM_API_KEY   — required
 *   LLM_BASE_URL  — default https://api.openai.com/v1
 *     MiniMax China: https://api.minimaxi.com/v1
 *   LLM_MODEL     — default gpt-4o-mini (MiniMax example: MiniMax-M2.7)
 *   LLM_TEMPERATURE — default 0.1
 *   TRANSLATION_CONTEXT_FILES — optional; comma-separated repo-relative paths
 *     Default: .github/translation-context.md, README.md
 *   TRANSLATION_CONTEXT_MAX_CHARS — default 8000
 */
import fs from "node:fs";
import path from "node:path";

const DEFAULT_CONTEXT_FILES = [".github/translation-context.md", "README.md"];

const apiKey = process.env.LLM_API_KEY;
const baseUrl = (process.env.LLM_BASE_URL || "https://api.openai.com/v1").replace(/\/$/, "");
const model = process.env.LLM_MODEL || "gpt-4o-mini";
const rawTemp = process.env.LLM_TEMPERATURE;
const parsedTemp = rawTemp != null && rawTemp !== "" ? Number.parseFloat(rawTemp) : Number.NaN;
const temperature = Number.isFinite(parsedTemp) ? parsedTemp : 0.1;
const inputPath = process.env.RELEASE_BODY_PATH || "release_body_en.md";
const outPath = process.env.TRANSLATED_OUT_PATH || "translated_release.md";

const isMiniMaxHost = /minimaxi\.com|minimax\.io/i.test(baseUrl);

const contextMaxChars = (() => {
  const raw = process.env.TRANSLATION_CONTEXT_MAX_CHARS;
  const n = raw != null && raw !== "" ? Number.parseInt(raw, 10) : Number.NaN;
  return Number.isFinite(n) && n > 0 ? n : 8000;
})();

function loadTranslationContext() {
  const rawList = process.env.TRANSLATION_CONTEXT_FILES?.trim();
  const paths = rawList
    ? rawList.split(",").map(p => p.trim()).filter(Boolean)
    : DEFAULT_CONTEXT_FILES;

  const chunks = [];
  for (const rel of paths) {
    const abs = path.resolve(process.cwd(), rel);
    if (!(fs.existsSync(abs) && fs.statSync(abs).isFile())) continue;
    let text = fs.readFileSync(abs, "utf8");
    if (text.length > contextMaxChars) {
      text = `${text.slice(0, contextMaxChars)}\n\n… [truncated, ${contextMaxChars} chars max per file]`;
    }
    chunks.push(`### File: ${rel}\n\n${text.trim()}`);
  }
  if (chunks.length === 0) return "";
  return chunks.join("\n\n---\n\n");
}

function stripInterleavedThinking(text) {
  return text.replace(/<redacted_thinking>[\s\S]*?<\/redacted_thinking>/gi, "").trim();
}

const rawBody = fs.existsSync(inputPath) ? fs.readFileSync(inputPath, "utf8") : "";

if (!rawBody.trim()) {
  console.log("No release body to translate; skipping.");
  process.exit(0);
}

// If the body already contains a bilingual separator, extract only the English portion.
const sepIndex = rawBody.indexOf("\n---\n");
const originalBody = sepIndex !== -1 ? rawBody.slice(0, sepIndex) : rawBody;

if (!originalBody.trim()) {
  console.log("No English portion to translate; skipping.");
  process.exit(0);
}

if (!apiKey) {
  console.error("Missing LLM_API_KEY.");
  process.exit(1);
}

const translationContext = loadTranslationContext();
if (translationContext) {
  console.log(`Loaded translation context (${translationContext.length} chars from project docs).`);
}

const contextSection = translationContext
  ? `\n以下是本项目参考文档摘录（用于统一产品名、功能译法与语气；请勿把本段当作要翻译的正文去输出）。\n\n${translationContext}\n\n---\n`
  : "";

const prompt = `你是一个资深的技术研发文档翻译专家。请将以下 GitHub Release / Changelog 从英文翻译为中文。

规则：
1. 保持原有的 Markdown 结构（缩进、列表、加粗、链接、代码块）不变。
2. 必须将 Changelog 的标准英文大纲标题翻译为固定中文（默认不新增 Emoji，除非原文标题本身已包含 Emoji）：
   - "Features" -> "新特性"
   - "Bug Fixes" -> "问题修复"
   - "Performance Improvements" -> "性能优化"
   - "Documentation" -> "文档"
3. 专业术语（如 API、S3、CDN、CSP、SSO、OAuth、JWT、Webhook、MCP 等）在中文语境下可保留英文或常见译法，保持一致性；若「参考文档」中有约定译名或产品名，请优先遵循参考文档。
4. 语气专业、自然，符合中国开发者阅读习惯。
5. 只输出译文正文，不要前言、不要 "Here is the translation" 等套话。
${contextSection}
待翻译原文：

${originalBody}`;

console.log(`Calling LLM (model=${model}, base=${baseUrl})...`);

try {
  const requestBody = {
    model,
    messages: [{ role: "user", content: prompt }],
    temperature,
  };
  if (isMiniMaxHost) {
    requestBody.reasoning_split = true;
  }

  const response = await fetch(`${baseUrl}/chat/completions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${apiKey}`,
    },
    body: JSON.stringify(requestBody),
  });

  const data = await response.json();

  if (!response.ok) {
    console.error("API error:", response.status, JSON.stringify(data));
    process.exit(1);
  }

  let translatedBody = data.choices?.[0]?.message?.content;
  if (!translatedBody || typeof translatedBody !== "string") {
    console.error("Model returned no text:", JSON.stringify(data));
    process.exit(1);
  }

  const stripped = stripInterleavedThinking(translatedBody);
  translatedBody = stripped || translatedBody.trim();

  if (!translatedBody) {
    console.error("Translation empty after processing.");
    process.exit(1);
  }

  const finalMarkdown = `${originalBody.trimEnd()}

---

${translatedBody.trim()}
`;

  fs.writeFileSync(outPath, finalMarkdown, "utf8");
  console.log(`Wrote ${outPath}`);
} catch (err) {
  console.error("translate_release failed:", err);
  process.exit(1);
}

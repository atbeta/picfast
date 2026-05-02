import { z } from "zod";

const envSchema = z.object({
  PICFAST_BASE_URL: z.string().min(1, "PICFAST_BASE_URL is required"),
  PICFAST_API_TOKEN: z.string().optional(),
});

export type Config = z.infer<typeof envSchema>;

export function loadConfig(): Config {
  const env = {
    PICFAST_BASE_URL: process.env.PICFAST_BASE_URL ?? "",
    PICFAST_API_TOKEN: process.env.PICFAST_API_TOKEN || undefined,
  };

  const result = envSchema.safeParse(env);
  if (!result.success) {
    const issues = result.error.issues
      .map((i) => `${i.path.join(".")}: ${i.message}`)
      .join(", ");
    throw new Error(`Invalid configuration: ${issues}`);
  }

  return result.data;
}
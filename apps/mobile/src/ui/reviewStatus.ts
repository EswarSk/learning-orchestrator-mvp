export function reviewStatusLabel(status?: string): string | null {
  if (status === "reviewed") return null;
  if (status === "ai_generated") return "AI-generated · Not educator-reviewed";
  return "Not educator-reviewed";
}

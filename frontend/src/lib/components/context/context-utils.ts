export const CATEGORY_LABELS: Record<string, string> = {
  system_prompt_and_tool_definitions: "System + tools",
  user_messages: "User",
  assistant_messages: "Assistant",
  thinking: "Thinking",
  tool_calls: "Tool calls",
  tool_outputs: "Tool outputs",
  file_reads: "File reads",
  search_results: "Search / grep",
  summaries_and_compacted_handoffs: "Summary / compact",
  subagent_outputs: "Subagent outputs",
  free_space: "Free space",
  other: "Other",
};

export const CATEGORY_COLORS: Record<string, string> = {
  system_prompt_and_tool_definitions: "var(--text-muted)",
  user_messages: "var(--accent-teal)",
  assistant_messages: "var(--accent-blue)",
  thinking: "var(--accent-purple)",
  tool_calls: "var(--accent-amber)",
  tool_outputs: "var(--accent-orange)",
  file_reads: "var(--accent-sky)",
  search_results: "var(--accent-indigo)",
  summaries_and_compacted_handoffs: "var(--accent-rose)",
  subagent_outputs: "var(--accent-pink)",
  free_space: "var(--border-muted)",
  other: "var(--text-muted)",
};

export function categoryLabel(category: string): string {
  return CATEGORY_LABELS[category] ?? category;
}

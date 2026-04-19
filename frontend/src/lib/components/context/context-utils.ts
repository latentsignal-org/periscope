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
  system_prompt_and_tool_definitions: "#6b7280",
  user_messages: "#0f766e",
  assistant_messages: "#1d4ed8",
  thinking: "#7c3aed",
  tool_calls: "#b45309",
  tool_outputs: "#ea580c",
  file_reads: "#2563eb",
  search_results: "#0891b2",
  summaries_and_compacted_handoffs: "#be123c",
  subagent_outputs: "#4f46e5",
  free_space: "#d1d5db",
  other: "#64748b",
};

export function categoryLabel(category: string): string {
  return CATEGORY_LABELS[category] ?? category;
}

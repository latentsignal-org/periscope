export interface ContextCapacity {
  max_tokens: number;
  provenance: string;
  model?: string;
  agent?: string;
}

export interface ContextSummary {
  tokens_in_use: number;
  tokens_provenance: string;
  percent_consumed: number;
  percent_provenance: string;
  remaining_tokens: number;
  remaining_known: boolean;
  visible_row_count: number;
  visible_since_ordinal: number;
  last_updated_at?: string;
  row_granularity: string;
}

export interface ContextCompositionItem {
  category: string;
  tokens: number;
  percentage: number;
  provenance: string;
}

export interface ContextSupports {
  live_updates: boolean;
  standalone_route: boolean;
  embedded_tab: boolean;
  transcript_jump: boolean;
  row_granularity: string;
  compaction_trimmed: boolean;
}

export interface SessionContextResponse {
  summary: ContextSummary;
  capacity: ContextCapacity;
  composition: ContextCompositionItem[];
  supports: ContextSupports;
  warnings?: string[];
}

export interface ContextCategoryValue {
  category: string;
  tokens: number;
}

export interface ContextTimelineRow {
  ordinal: number;
  timestamp?: string;
  label: string;
  granularity: string;
  delta_tokens: number;
  delta_provenance: string;
  cumulative_tokens: number;
  cumulative_provenance: string;
  dominant_category?: string;
  categories: ContextCategoryValue[];
  markers?: string[];
  annotations?: string[];
}

export interface SessionContextTimelineResponse {
  timeline: ContextTimelineRow[];
  supports: ContextSupports;
  warnings?: string[];
}

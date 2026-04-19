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

export interface RewindSignal {
  should_rewind: boolean;
  confidence: string;
  reasons: string[];
  tokens_recoverable: number;
  score: number;
  rewind_to_turn?: number;
  rewind_to_reason?: string;
  bad_stretch_from?: number;
  bad_stretch_to?: number;
}

export interface CompactSignal {
  should_compact: boolean;
  confidence: string;
  reasons: string[];
  score: number;
  estimated_reclaimable: number;
  compact_focus?: string[];
}

export interface SessionContextResponse {
  summary: ContextSummary;
  capacity: ContextCapacity;
  composition: ContextCompositionItem[];
  supports: ContextSupports;
  warnings?: string[];
  rewind_signal?: RewindSignal;
  compact_signal?: CompactSignal;
}

export interface ContextCategoryValue {
  category: string;
  tokens: number;
}

export interface ContextTimelineMessagePreview {
  ordinal: number;
  preview: string;
}

export interface ContextTimelineToolPreview {
  ordinal: number;
  tool_name: string;
  snippet?: string;
}

export interface ContextTimelineEntry {
  kind: string;
  ordinal: number;
  label: string;
  preview?: string;
  output_preview?: string;
}

export interface ContextTimelineTurn {
  turn: number;
  start_ordinal: number;
  end_ordinal: number;
  timestamp?: string;
  label: string;
  delta_tokens: number;
  delta_provenance: string;
  cumulative_tokens: number;
  cumulative_provenance: string;
  dominant_category?: string;
  categories: ContextCategoryValue[];
  markers?: string[];
  annotations?: string[];
  user_message?: ContextTimelineMessagePreview;
  assistant_message?: ContextTimelineMessagePreview;
  tool_calls?: ContextTimelineToolPreview[];
  entries?: ContextTimelineEntry[];
}

export interface SessionContextTimelineResponse {
  timeline: ContextTimelineTurn[];
  supports: ContextSupports;
  warnings?: string[];
}

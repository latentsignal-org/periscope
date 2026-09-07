/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DbCallTiming } from './DbCallTiming';
export type DbSessionTiming = {
  by_category: any[] | null;
  running: boolean;
  session_id: string;
  slowest_call: DbCallTiming;
  subagent_count: number;
  tool_call_count: number;
  tool_duration_ms: number;
  total_duration_ms: number;
  turn_count: number;
  turns: any[] | null;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DbPercentiles } from './DbPercentiles';
export type DbVelocityOverview = {
  chars_per_active_min: number;
  first_response_sec: DbPercentiles;
  msgs_per_active_min: number;
  tool_calls_per_active_min: number;
  turn_cycle_sec: DbPercentiles;
};


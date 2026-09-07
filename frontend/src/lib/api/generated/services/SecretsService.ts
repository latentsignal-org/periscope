/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { ServiceSecretFindingList } from '../models/ServiceSecretFindingList';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class SecretsService {
  /**
   * List secret findings
   * @returns ServiceSecretFindingList OK
   * @throws ApiError
   */
  public static getApiV1Secrets({
    project,
    agent,
    dateFrom,
    dateTo,
    rule,
    confidence,
    reveal,
    limit,
    cursor,
  }: {
    /**
     * Filter by project
     */
    project?: string,
    /**
     * Filter by agent
     */
    agent?: string,
    /**
     * Filter start date
     */
    dateFrom?: string,
    /**
     * Filter end date
     */
    dateTo?: string,
    /**
     * Filter by secret rule
     */
    rule?: string,
    /**
     * Filter by confidence
     */
    confidence?: string,
    /**
     * Return unredacted matches for localhost callers
     */
    reveal?: boolean,
    /**
     * Maximum number of results
     */
    limit?: number,
    /**
     * Pagination cursor
     */
    cursor?: number,
  }): CancelablePromise<ServiceSecretFindingList> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/api/v1/secrets',
      query: {
        'project': project,
        'agent': agent,
        'date_from': dateFrom,
        'date_to': dateTo,
        'rule': rule,
        'confidence': confidence,
        'reveal': reveal,
        'limit': limit,
        'cursor': cursor,
      },
      errors: {
        400: `Bad Request`,
        401: `Unauthorized`,
        403: `Forbidden`,
        404: `Not Found`,
        409: `Conflict`,
        422: `Unprocessable Entity`,
        500: `Internal Server Error`,
        501: `Not Implemented`,
        502: `Bad Gateway`,
        503: `Service Unavailable`,
        504: `Gateway Timeout`,
      },
    });
  }
  /**
   * Scan secrets
   * @returns string OK
   * @throws ApiError
   */
  public static postApiV1SecretsScan({
    backfill,
    project,
    agent,
    dateFrom,
    dateTo,
  }: {
    /**
     * Backfill all matching sessions
     */
    backfill?: boolean,
    /**
     * Filter by project
     */
    project?: string,
    /**
     * Filter by agent
     */
    agent?: string,
    /**
     * Filter start date
     */
    dateFrom?: string,
    /**
     * Filter end date
     */
    dateTo?: string,
  }): CancelablePromise<string> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/api/v1/secrets/scan',
      query: {
        'backfill': backfill,
        'project': project,
        'agent': agent,
        'date_from': dateFrom,
        'date_to': dateTo,
      },
      errors: {
        400: `Bad Request`,
        401: `Unauthorized`,
        403: `Forbidden`,
        404: `Not Found`,
        409: `Conflict`,
        422: `Unprocessable Entity`,
        500: `Internal Server Error`,
        501: `Not Implemented`,
        502: `Bad Gateway`,
        503: `Service Unavailable`,
        504: `Gateway Timeout`,
      },
    });
  }
}

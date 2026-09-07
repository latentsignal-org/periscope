/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { ApplyWorktreeMappingsResponse } from '../models/ApplyWorktreeMappingsResponse';
import type { DbWorktreeProjectMapping } from '../models/DbWorktreeProjectMapping';
import type { SettingsResponse } from '../models/SettingsResponse';
import type { SettingsUpdateRequest } from '../models/SettingsUpdateRequest';
import type { WorktreeMappingRequest } from '../models/WorktreeMappingRequest';
import type { WorktreeMappingsResponse } from '../models/WorktreeMappingsResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class SettingsService {
  /**
   * Get settings
   * @returns SettingsResponse OK
   * @throws ApiError
   */
  public static getApiV1Settings(): CancelablePromise<SettingsResponse> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/api/v1/settings',
      errors: {
        400: `Bad Request`,
        401: `Unauthorized`,
        403: `Forbidden`,
        404: `Not Found`,
        409: `Conflict`,
        500: `Internal Server Error`,
        501: `Not Implemented`,
        502: `Bad Gateway`,
        503: `Service Unavailable`,
        504: `Gateway Timeout`,
      },
    });
  }
  /**
   * Update settings
   * @returns SettingsResponse OK
   * @throws ApiError
   */
  public static putApiV1Settings({
    requestBody,
  }: {
    requestBody: SettingsUpdateRequest,
  }): CancelablePromise<SettingsResponse> {
    return __request(OpenAPI, {
      method: 'PUT',
      url: '/api/v1/settings',
      body: requestBody,
      mediaType: 'application/json',
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
   * List worktree mappings
   * @returns WorktreeMappingsResponse OK
   * @throws ApiError
   */
  public static getApiV1SettingsWorktreeMappings(): CancelablePromise<WorktreeMappingsResponse> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/api/v1/settings/worktree-mappings',
      errors: {
        400: `Bad Request`,
        401: `Unauthorized`,
        403: `Forbidden`,
        404: `Not Found`,
        409: `Conflict`,
        500: `Internal Server Error`,
        501: `Not Implemented`,
        502: `Bad Gateway`,
        503: `Service Unavailable`,
        504: `Gateway Timeout`,
      },
    });
  }
  /**
   * Create worktree mapping
   * @returns DbWorktreeProjectMapping OK
   * @throws ApiError
   */
  public static postApiV1SettingsWorktreeMappings({
    requestBody,
  }: {
    requestBody: WorktreeMappingRequest,
  }): CancelablePromise<DbWorktreeProjectMapping> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/api/v1/settings/worktree-mappings',
      body: requestBody,
      mediaType: 'application/json',
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
   * Apply worktree mappings
   * @returns ApplyWorktreeMappingsResponse OK
   * @throws ApiError
   */
  public static postApiV1SettingsWorktreeMappingsApply(): CancelablePromise<ApplyWorktreeMappingsResponse> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/api/v1/settings/worktree-mappings/apply',
      errors: {
        400: `Bad Request`,
        401: `Unauthorized`,
        403: `Forbidden`,
        404: `Not Found`,
        409: `Conflict`,
        500: `Internal Server Error`,
        501: `Not Implemented`,
        502: `Bad Gateway`,
        503: `Service Unavailable`,
        504: `Gateway Timeout`,
      },
    });
  }
  /**
   * Delete worktree mapping
   * @returns void
   * @throws ApiError
   */
  public static deleteApiV1SettingsWorktreeMappingsId({
    id,
  }: {
    /**
     * Mapping ID
     */
    id: string,
  }): CancelablePromise<void> {
    return __request(OpenAPI, {
      method: 'DELETE',
      url: '/api/v1/settings/worktree-mappings/{id}',
      path: {
        'id': id,
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
   * Update worktree mapping
   * @returns DbWorktreeProjectMapping OK
   * @throws ApiError
   */
  public static putApiV1SettingsWorktreeMappingsId({
    id,
    requestBody,
  }: {
    /**
     * Mapping ID
     */
    id: string,
    requestBody: WorktreeMappingRequest,
  }): CancelablePromise<DbWorktreeProjectMapping> {
    return __request(OpenAPI, {
      method: 'PUT',
      url: '/api/v1/settings/worktree-mappings/{id}',
      path: {
        'id': id,
      },
      body: requestBody,
      mediaType: 'application/json',
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

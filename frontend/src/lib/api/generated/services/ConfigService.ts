/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { GithubConfigResponse } from '../models/GithubConfigResponse';
import type { SetGithubConfigInputBody } from '../models/SetGithubConfigInputBody';
import type { SetGithubConfigResponse } from '../models/SetGithubConfigResponse';
import type { TerminalConfigBody } from '../models/TerminalConfigBody';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ConfigService {
  /**
   * Get GitHub config
   * @returns GithubConfigResponse OK
   * @throws ApiError
   */
  public static getApiV1ConfigGithub(): CancelablePromise<GithubConfigResponse> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/api/v1/config/github',
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
   * Set GitHub config
   * @returns SetGithubConfigResponse OK
   * @throws ApiError
   */
  public static postApiV1ConfigGithub({
    requestBody,
  }: {
    requestBody: SetGithubConfigInputBody,
  }): CancelablePromise<SetGithubConfigResponse> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/api/v1/config/github',
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
   * Get terminal config
   * @returns TerminalConfigBody OK
   * @throws ApiError
   */
  public static getApiV1ConfigTerminal(): CancelablePromise<TerminalConfigBody> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/api/v1/config/terminal',
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
   * Set terminal config
   * @returns TerminalConfigBody OK
   * @throws ApiError
   */
  public static postApiV1ConfigTerminal({
    requestBody,
  }: {
    requestBody: TerminalConfigBody,
  }): CancelablePromise<TerminalConfigBody> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/api/v1/config/terminal',
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

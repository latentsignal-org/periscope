/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ImportService {
  /**
   * Import ChatGPT archive
   * @returns any OK
   * @throws ApiError
   */
  public static postApiV1ImportChatgpt({
    accept,
    formData,
  }: {
    /**
     * Use text/event-stream to stream progress
     */
    accept?: string,
    formData?: {
      file: Blob;
    },
  }): CancelablePromise<Record<string, any>> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/api/v1/import/chatgpt',
      headers: {
        'Accept': accept,
      },
      formData: formData,
      mediaType: 'multipart/form-data',
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
   * Import Claude.ai archive
   * @returns any OK
   * @throws ApiError
   */
  public static postApiV1ImportClaudeAi({
    accept,
    formData,
  }: {
    /**
     * Use text/event-stream to stream progress
     */
    accept?: string,
    formData?: {
      file: Blob;
    },
  }): CancelablePromise<Record<string, any>> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/api/v1/import/claude-ai',
      headers: {
        'Accept': accept,
      },
      formData: formData,
      mediaType: 'multipart/form-data',
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

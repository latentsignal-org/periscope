/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { OpenersResponse } from '../models/OpenersResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class OpenersService {
  /**
   * List openers
   * @returns OpenersResponse OK
   * @throws ApiError
   */
  public static getApiV1Openers(): CancelablePromise<OpenersResponse> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/api/v1/openers',
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
}

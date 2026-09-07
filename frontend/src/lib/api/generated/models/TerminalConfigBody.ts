/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type TerminalConfigBody = {
  /**
   * Argument template containing {cmd} when mode is custom
   */
  custom_args?: string;
  /**
   * Terminal binary path when mode is custom
   */
  custom_bin?: string;
  /**
   * Terminal launch mode
   */
  mode: TerminalConfigBody.mode;
};
export namespace TerminalConfigBody {
  /**
   * Terminal launch mode
   */
  export enum mode {
    AUTO = 'auto',
    CUSTOM = 'custom',
    CLIPBOARD = 'clipboard',
  }
}


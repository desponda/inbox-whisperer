import type { AxiosRequestConfig, AxiosResponse } from 'axios';
import type { GetApiAuthCallbackParams } from '../inboxWhispererAPI.schemas';
export declare const getAuth: () => {
    getApiAuthLogin: <TData = AxiosResponse<unknown, any>>(options?: AxiosRequestConfig) => Promise<TData>;
    getApiAuthCallback: <TData = AxiosResponse<unknown, any>>(params: GetApiAuthCallbackParams, options?: AxiosRequestConfig) => Promise<TData>;
};
export type GetApiAuthLoginResult = AxiosResponse<unknown>;
export type GetApiAuthCallbackResult = AxiosResponse<unknown>;

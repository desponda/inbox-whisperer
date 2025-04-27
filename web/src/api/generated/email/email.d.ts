import type { AxiosRequestConfig, AxiosResponse } from 'axios';
import type { EmailContent, EmailSummary } from '../inboxWhispererAPI.schemas';
export declare const getEmail: () => {
    getApiEmailMessages: <TData = AxiosResponse<EmailSummary[], any>>(options?: AxiosRequestConfig) => Promise<TData>;
    getApiEmailMessagesId: <TData = AxiosResponse<EmailContent, any>>(id: string, options?: AxiosRequestConfig) => Promise<TData>;
};
export type GetApiEmailMessagesResult = AxiosResponse<EmailSummary[]>;
export type GetApiEmailMessagesIdResult = AxiosResponse<EmailContent>;

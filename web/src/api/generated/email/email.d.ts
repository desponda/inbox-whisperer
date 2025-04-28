import type { AxiosRequestConfig, AxiosResponse } from 'axios';
import type { EmailContent, EmailSummary, PutApiEmailsIdBody } from '../inboxWhispererAPI.schemas';
export declare const getEmail: () => {
    getApiEmails: <TData = AxiosResponse<EmailSummary[], any>>(options?: AxiosRequestConfig) => Promise<TData>;
    getApiEmailsId: <TData = AxiosResponse<EmailContent, any>>(id: string, options?: AxiosRequestConfig) => Promise<TData>;
    putApiEmailsId: <TData = AxiosResponse<unknown, any>>(id: string, putApiEmailsIdBody: PutApiEmailsIdBody, options?: AxiosRequestConfig) => Promise<TData>;
    deleteApiEmailsId: <TData = AxiosResponse<unknown, any>>(id: string, options?: AxiosRequestConfig) => Promise<TData>;
};
export type GetApiEmailsResult = AxiosResponse<EmailSummary[]>;
export type GetApiEmailsIdResult = AxiosResponse<EmailContent>;
export type PutApiEmailsIdResult = AxiosResponse<unknown>;
export type DeleteApiEmailsIdResult = AxiosResponse<unknown>;

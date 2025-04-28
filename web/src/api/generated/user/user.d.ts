import type { AxiosRequestConfig, AxiosResponse } from 'axios';
import type { User, UserCreateRequest, UserUpdateRequest } from '../inboxWhispererAPI.schemas';
export declare const getUser: () => {
    getApiUsers: <TData = AxiosResponse<User[], any>>(options?: AxiosRequestConfig) => Promise<TData>;
    postApiUsers: <TData = AxiosResponse<User, any>>(userCreateRequest: UserCreateRequest, options?: AxiosRequestConfig) => Promise<TData>;
    getApiUsersMe: <TData = AxiosResponse<User, any>>(options?: AxiosRequestConfig) => Promise<TData>;
    getApiUsersId: <TData = AxiosResponse<User, any>>(id: string, options?: AxiosRequestConfig) => Promise<TData>;
    putApiUsersId: <TData = AxiosResponse<User, any>>(id: string, userUpdateRequest: UserUpdateRequest, options?: AxiosRequestConfig) => Promise<TData>;
    deleteApiUsersId: <TData = AxiosResponse<void, any>>(id: string, options?: AxiosRequestConfig) => Promise<TData>;
};
export type GetApiUsersResult = AxiosResponse<User[]>;
export type PostApiUsersResult = AxiosResponse<User>;
export type GetApiUsersMeResult = AxiosResponse<User>;
export type GetApiUsersIdResult = AxiosResponse<User>;
export type PutApiUsersIdResult = AxiosResponse<User>;
export type DeleteApiUsersIdResult = AxiosResponse<void>;
